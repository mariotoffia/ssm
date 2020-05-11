package pms

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/mariotoffia/ssm.git/internal/common"
	"github.com/mariotoffia/ssm.git/parser"
	"github.com/mariotoffia/ssm.git/support"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Serializer handles the parameter store communication
type Serializer struct {
	// AWS Config to use when communicating
	config aws.Config
	// The name of the service using this library
	service string
	// Default tier if not specified.
	tier ssm.ParameterTier
}

// SeDefaultTier allows for change the tier. By default the
// serializer uses the standard tier.
func (p *Serializer) SeDefaultTier(tier ssm.ParameterTier) *Serializer {
	p.tier = tier
	return p
}

// NewFromConfig creates a repository using a existing configuration
func NewFromConfig(config aws.Config, service string) *Serializer {
	return &Serializer{config: config, service: service,
		tier: ssm.ParameterTierStandard}
}

// New creates a repository using the default AWS configuration
func New(service string) (*Serializer, error) {
	awscfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return &Serializer{}, errors.Wrapf(err, "Failed to load AWS config")
	}

	return &Serializer{config: awscfg, service: service,
		tier: ssm.ParameterTierStandard}, nil
}

// Get parameters from the parameterstore and populates the node graph with values.
// Any fields that was not able to be set is reported in the FullNameField string map.
// FullNameField do not include those fields filtered out in exclusion filter.
func (p *Serializer) Get(node *parser.StructNode,
	filter *support.FieldFilters) (map[string]support.FullNameField, error) {

	m := map[string]*parser.StructNode{}
	parser.NodesToParameterMap(node, m, filter, []string{"pms"})
	paths := parser.ExtractPaths(m)
	issecure := isSecure(node)

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

func isSecure(node *parser.StructNode) bool {
	if node.HasChildren() {
		for _, n := range node.Childs {
			if isSecure(&n) {
				return true
			}
		}
	}

	if tag, ok := ToPmsTag(node); ok {
		return tag.Secure()
	}

	return false
}

// Upsert stores the node values (after filter is applied). If any
// error occurs it will return that in the support.FullNameField.Error
// field. Thus it is possible to track which fields did not get written
// to the Parameter store and hence needs to be handeled.
func (p *Serializer) Upsert(node *parser.StructNode,
	filter *support.FieldFilters) map[string]support.FullNameField {

	m := map[string]*parser.StructNode{}
	parser.NodesToParameterMap(node, m, filter, []string{"pms"})

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
			im[node.FqName] = p.createFullNameFieldNode(*prm.Name, err, m[*prm.Name])
			log.Debug().Str("svc", p.service).Msgf("Failed to write %v error: %v", im[node.FqName], err)

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
					im[node.FqName] = p.createFullNameFieldNode(*prm.Name, err, m[*prm.Name])
					log.Debug().Str("svc", p.service).Msgf("Failed to write tags on %v error: %v", im[node.FqName], err)

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

func (p *Serializer) handleInvalidRequestParameters(invalid []string,
	m map[string]*parser.StructNode) map[string]support.FullNameField {

	im := map[string]support.FullNameField{}

	if len(invalid) > 0 {
		for _, name := range invalid {
			if val, ok := m[name]; ok {
				im[val.FqName] = support.FullNameField{RemoteName: val.Tag["pms"].GetFullName(),
					LocalName: val.FqName, Field: val.Field, Value: val.Value}
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

func (p *Serializer) populate(node *parser.StructNode, params map[string]ssm.Parameter) error {
	node.EnsureInstance(false)

	if node.HasChildren() {
		for _, n := range node.Childs {
			p.populate(&n, params)
		}
		return nil
	}

	if !node.HasTag("pms") {
		log.Debug().Msgf("Node %s is not of pms type", node.FqName)
		return nil
	}

	if val, ok := params[node.Tag["pms"].GetFullName()]; ok {
		common.SetStructValueFromString(node, *val.Name, *val.Value)
	}

	return nil
}
