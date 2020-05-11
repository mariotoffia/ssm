package pms

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/mariotoffia/ssm.git/internal/common"
	"github.com/mariotoffia/ssm.git/parser"
	"github.com/mariotoffia/ssm.git/support"
	"github.com/pkg/errors"
)

func (p *Serializer) createFullNameFieldNode(remote string, err error,
	node *parser.StructNode) support.FullNameField {
	return support.FullNameField{
		LocalName:  node.FqName,
		RemoteName: remote,
		Field:      node.Field,
		Value:      node.Value,
		Error:      err}
}

func (p *Serializer) toPutParameters(prms map[string]*parser.StructNode) []ssm.PutParameterInput {
	params := []ssm.PutParameterInput{}
	for _, node := range prms {
		if tag, ok := ToPmsTag(node); ok {
			params = append(params, ssm.PutParameterInput{Name: aws.String(tag.FqName()),
				Overwrite: aws.Bool(tag.Overwrite()),
				Tier:      tag.SsmTier(p.tier),
				Tags:      tag.SsmTags(),
				Type:      ParameterType(node),
				Value:     aws.String(common.GetStringValueFromField(node)),
			})
		}
	}

	return params
}

// getFromAws fetches the data from the remote location.
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
