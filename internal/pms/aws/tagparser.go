package awspms

import (
	"github.com/mariotoffia/ssm.git/parser"
)

type tagParser struct {
	realParser parser.TagParser
}

// NewTagParser creates a new AWS parameter store tag
// parser.
func NewTagParser() parser.TagParser {
	return &tagParser{
		realParser: parser.NewTagParser([]string{
			"name",
			"prefix",
			"keyid",
			"description",
			"pattern",
			"overwrite",
		}),
	}
}

func (p *tagParser) ParseTagString(tagstring string,
	prefix string,
	env string,
	svc string) (parser.StructTag, error) {

	s, err := p.realParser.ParseTagString(tagstring, prefix, env, svc)
	if err != nil {
		return nil, err
	}

	q := s.(*parser.StructTagImpl)

	if _, ok := q.Named["overwrite"]; !ok {
		q.Named["overwrite"] = "true"
	}

	return &PmsTagStruct{StructTagImpl: *q}, nil
}
