package asm

import (
	"github.com/mariotoffia/ssm/parser"
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
			"vid",
			"vs",
			"strkey",
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
	return &AsmTagStruct{StructTagImpl: *q}, nil
}
