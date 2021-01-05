package pms

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/mariotoffia/ssm/internal/common"
	"github.com/mariotoffia/ssm/parser"
	"github.com/mariotoffia/ssm/support"
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

func (p *Serializer) toPutParameters(parameters map[string]*parser.StructNode) []ssm.PutParameterInput {
	params := []ssm.PutParameterInput{}
	for _, node := range parameters {

		if tag, ok := ToPmsTag(node); ok {

			params = append(params, ssm.PutParameterInput{Name: aws.String(tag.FqName()),
				Overwrite: tag.Overwrite(),
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
func (p *Serializer) getFromAws(params *ssm.GetParametersInput) (map[string]types.Parameter, []string, error) {

	client := ssm.NewFromConfig(p.config)

	var resp *ssm.GetParametersOutput

	var err error
	success := false

	for i := 0; i < 3 && !success; i++ {

		resp, err = client.GetParameters(context.Background(), params)

		if err != nil {

			return nil, nil, errors.Wrapf(err, "Failed fetch pms config entries %v", params)

		}

		if len(resp.Parameters) == 0 && len(resp.InvalidParameters) > 0 {

			time.Sleep(400 * time.Millisecond)

		} else {

			success = true

		}
	}

	m := map[string]types.Parameter{}

	for _, p := range resp.Parameters {

		key := *p.Name
		m[key] = p

	}

	return m, resp.InvalidParameters, nil
}
