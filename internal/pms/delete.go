package pms

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/mariotoffia/ssm/parser"
	"github.com/mariotoffia/ssm/support"
	"github.com/rs/zerolog/log"
)

// Delete will delete all paths described by _node_ tree. This is the
// "inverse" of `Get`.
func (p *Serializer) Delete(node *parser.StructNode,
	filter *support.FieldFilters) (map[string]support.FullNameField, error) {

	m := map[string]*parser.StructNode{}
	parser.NodesToParameterMap(node, m, filter, []string{"pms"})
	paths := parser.ExtractPaths(m)

	client := ssm.New(p.config)

	input := ssm.DeleteParametersInput{
		Names: paths,
	}

	req := client.DeleteParametersRequest(&input)
	result, err := req.Send(context.Background())
	if err != nil {
		return nil, err
	}

	im := p.handleInvalidRequestParameters(result.InvalidParameters, m, "delete")

	return im, nil
}

// DeleteTree lists all parameters that begins with a certain _prefix_
// and deletes those.
//
// This function accepts a set of prefixes and therefore may delete several
// trees.
func (p *Serializer) DeleteTree(prefixes ...string) error {

	inp := ssm.DescribeParametersInput{
		ParameterFilters: []ssm.ParameterStringFilter{{
			Key:    aws.String("Name"),
			Option: aws.String("BeginsWith"),
			Values: prefixes,
		}}}

	client := ssm.New(p.config)

	for {
		req := client.DescribeParametersRequest(&inp)
		res, err := req.Send(context.Background())
		if err != nil {
			log.Warn().Msgf("got error when listing params for deletion error: %v", err)
			return err
		}

		dprm := ssm.DeleteParametersInput{}
		for _, prm := range res.Parameters {
			log.Debug().Msgf("Deleting pms-param name: %s version %d", *prm.Name, *prm.Version)
			dprm.Names = append(dprm.Names, *prm.Name)
		}

		if len(dprm.Names) > 0 {
			dreq := client.DeleteParametersRequest(&dprm)
			_, err = dreq.Send(context.Background())
			if err != nil {
				log.Warn().Msgf("got error when deleting params error: %v", err)
				return err
			}
		} else {
			log.Debug().Msgf("No pms-parameters to delete")
		}

		if res.NextToken == nil {
			break
		}

		inp.NextToken = res.NextToken
	}

	return nil
}
