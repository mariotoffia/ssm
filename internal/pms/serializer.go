package pms

import (
	"context"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/mariotoffia/ssm.git/internal/common"
	"github.com/mariotoffia/ssm.git/internal/reflectparser"
	"github.com/mariotoffia/ssm.git/internal/tagparser"
	"github.com/mariotoffia/ssm.git/support"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Serializer handles the parameter store communication
type Serializer struct {
	config  aws.Config
	service string
	tier    ssm.ParameterTier
}

// SetTier allows for change the tier. By default PmsRepo uses
// the standard tier.
func (p *Serializer) SetTier(tier ssm.ParameterTier) *Serializer {
	p.tier = tier
	return p
}

// NewFromConfig creates a repository using a existing configuration
func NewFromConfig(config aws.Config, service string) *Serializer {
	return &Serializer{config: config, service: service,
		tier: ssm.ParameterTierStandard}
}

// NewWithRegion creates a repository using the default configuration
// with optional region override.
func NewWithRegion(region string, service string) (*Serializer, error) {
	awscfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return &Serializer{}, errors.Wrapf(err, "Failed to load AWS config")
	}

	if len(region) > 0 {
		awscfg.Region = region
	}

	return &Serializer{config: awscfg, service: service,
		tier: ssm.ParameterTierStandard}, nil
}

// New creates a repository using the default configuration.
func New(service string) (*Serializer, error) {
	return NewWithRegion("", service)
}

// Get parameters from the parameterstore and populates the node graph with values.
// Any fields that was not able to be set is reported in the FullNameField string map.
// FullNameField do not include those fields filtered out in exclusion filter.
func (p *Serializer) Get(node *reflectparser.SsmNode,
	filter *support.FieldFilters) (map[string]support.FullNameField, error) {

	m := map[string]*reflectparser.SsmNode{}
	issecure := common.NodesToParameterMap(node, m, filter, tagparser.Pms)
	paths := common.ExtractParameters(m)

	params := &ssm.GetParametersInput{
		Names:          paths,
		WithDecryption: aws.Bool(issecure),
	}

	log.Debug().Str("svc", p.service).Msgf("Fetching: %v", params)

	prms, invalid, err := p.getFromAws(params)
	if err != nil {
		return nil, err
	}

	im := p.handleInvalidRequestParameters(invalid, m)
	err = p.populate(node, prms)

	return im, err
}

// Upsert stores the node values (after filter is applied). If any
// error occurs it will return that in the support.FullNameField.Error
// field. Thus it is possible to track which fields did not get written
// to the Parameter store and hence needs to be handeled.
func (p *Serializer) Upsert(node *reflectparser.SsmNode,
	filter *support.FieldFilters) map[string]support.FullNameField {

	m := map[string]*reflectparser.SsmNode{}
	common.NodesToParameterMap(node, m, filter, tagparser.Pms)

	im := map[string]support.FullNameField{}
	if len(m) == 0 {
		return im
	}
	params := p.toPutParameters(m)
	client := ssm.New(p.config)

	for _, prm := range params {
		tags := prm.Tags
		prm.Tags = nil

		req := client.PutParameterRequest(&prm)
		resp, err := req.Send(context.TODO())
		if err != nil {
			im[node.FqName()] = p.createFullNameFieldNode(*prm.Name, err, m[*prm.Name])
			log.Debug().Str("svc", p.service).Msgf("Failed to write %v error: %v", im[node.FqName()], err)

		} else {
			log.Debug().Str("svc", p.service).Msgf("Succesfully wrote %v", resp)

			if len(tags) > 0 {
				req := client.AddTagsToResourceRequest(&ssm.AddTagsToResourceInput{
					ResourceId:   prm.Name,
					ResourceType: ssm.ResourceTypeForTaggingParameter,
					Tags:         tags,
				})

				resp, err := req.Send(context.TODO())
				if err != nil {
					im[node.FqName()] = p.createFullNameFieldNode(*prm.Name, err, m[*prm.Name])
					log.Debug().Str("svc", p.service).Msgf("Failed to write tags on %v error: %v", im[node.FqName()], err)

				} else {
					log.Debug().Str("svc", p.service).Msgf("Succesfully wrote tags %v", resp)
				}
			} else {
				log.Debug().Str("svc", p.service).Msgf("No tags to add to %s - skipping", *prm.Name)
			}
		}

	}

	return im
}

func (p *Serializer) createFullNameFieldNode(remote string, err error,
	node *reflectparser.SsmNode) support.FullNameField {
	return support.FullNameField{
		LocalName:  node.FqName(),
		RemoteName: remote,
		Field:      node.Field(),
		Value:      node.Value(),
		Error:      err}
}

func (p *Serializer) toPutParameters(prms map[string]*reflectparser.SsmNode) []ssm.PutParameterInput {

	params := []ssm.PutParameterInput{}
	for _, node := range prms {
		if tag, ok := node.Tag().(*tagparser.PmsTag); ok {

			params = append(params, ssm.PutParameterInput{Name: aws.String(node.Tag().FullName()),
				Overwrite: aws.Bool(tag.Overwrite()),
				Tier:      p.getTierFromTag(tag),
				Tags:      getTagsFromTag(tag),
				Type:      getParameterType(node),
				Value:     aws.String(common.GetStringValueFromField(node)),
			})

		}
	}

	return params
}
func getParameterType(node *reflectparser.SsmNode) ssm.ParameterType {
	if node.Tag().Secure() {
		return ssm.ParameterTypeSecureString
	}
	if node.Value().Kind() == reflect.Slice {
		return ssm.ParameterTypeStringList
	}
	return ssm.ParameterTypeString
}

func (p *Serializer) getTierFromTag(pmstag *tagparser.PmsTag) ssm.ParameterTier {

	switch pmstag.Tier() {
	case tagparser.Default:
		return p.tier
	case tagparser.Std:
		return ssm.ParameterTierStandard
	case tagparser.Adv:
		return ssm.ParameterTierAdvanced
	case tagparser.Eval:
		return ssm.ParameterTierIntelligentTiering
	}

	return p.tier
}

func getTagsFromTag(pmstag *tagparser.PmsTag) []ssm.Tag {
	tags := []ssm.Tag{}

	for key, value := range pmstag.Tags() {
		tags = append(tags, ssm.Tag{Key: aws.String(key), Value: aws.String(value)})
	}

	return tags
}

func (p *Serializer) handleInvalidRequestParameters(invalid []string,
	m map[string]*reflectparser.SsmNode) map[string]support.FullNameField {

	im := map[string]support.FullNameField{}

	if len(invalid) > 0 {
		for _, name := range invalid {
			if val, ok := m[name]; ok {
				im[val.FqName()] = support.FullNameField{RemoteName: val.Tag().FullName(),
					LocalName: val.FqName(), Field: val.Field(), Value: val.Value()}
			} else {
				log.Warn().Str("service", p.service).Msgf("Could not find %s in node map", name)
			}
		}
	}

	if len(im) > 0 {
		for key, val := range im {
			log.Debug().Msgf("not fetched: %s [%s]", key, val.RemoteName)
		}
	}
	return im
}
func (p *Serializer) populate(node *reflectparser.SsmNode, params map[string]ssm.Parameter) error {
	node.EnsureInstance(false)

	if node.HasChildren() {
		for _, n := range node.Children() {
			p.populate(&n, params)
		}
		return nil
	}

	if node.Tag().SsmType() != tagparser.Pms {
		log.Debug().Msgf("Node %s is not of pms type", node.Tag().FullName())
		return nil
	}

	if val, ok := params[node.Tag().FullName()]; ok {
		common.SetStructValueFromString(node, *val.Name, *val.Value)
	}

	return nil
}

// Invoke get towards aws parameter store
func (p *Serializer) getFromAws(params *ssm.GetParametersInput) (map[string]ssm.Parameter, []string, error) {
	client := ssm.New(p.config)

	var resp *ssm.GetParametersResponse
	var err error
	success := false
	for i := 0; i < 3 && !success; i++ {
		req := client.GetParametersRequest(params)

		resp, err = req.Send(context.TODO())
		if err != nil {
			return nil, nil, errors.Wrapf(err, "Failed fetch pms config entries %v", params)
		}

		if len(resp.Parameters) == 0 && len(resp.InvalidParameters) > 0 {
			time.Sleep(400 * time.Millisecond)
		} else {
			success = true
		}
	}

	m := map[string]ssm.Parameter{}
	for _, p := range resp.Parameters {
		key := *p.Name
		m[key] = p
	}

	return m, resp.InvalidParameters, nil
}
