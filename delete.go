package ssm

import (
	"reflect"

	"github.com/mariotoffia/ssm/internal/asm"
	"github.com/mariotoffia/ssm/internal/pms"
	"github.com/mariotoffia/ssm/parser"
	"github.com/mariotoffia/ssm/support"
)

func (s *Serializer) delete(v interface{},
	filter *support.FieldFilters,
	usage []Usage) (map[string]support.FullNameField, *parser.StructNode, error) {

	if len(usage) == 0 {
		if len(s.usage) > 0 {
			usage = s.usage
		} else {
			usage = []Usage{UsePms, UseAsm}
		}
	}

	if nil == filter {
		filter = support.NewFilters()
	}

	tp := reflect.ValueOf(v)
	prs := parser.New(s.service, s.env, s.prefix)

	if _, found := find(usage, UsePms); found {
		prs.RegisterTagParser("pms", pms.NewTagParser())
	}
	if _, found := find(usage, UseAsm); found {
		prs.RegisterTagParser("asm", asm.NewTagParser())
	}

	for n, v := range s.parser {
		prs.RegisterTagParser(n, v)
	}

	node, err := prs.Parse(tp)
	if err != nil {
		return nil, nil, err
	}

	var invalid map[string]support.FullNameField

	if _, found := find(usage, UsePms); found {
		pmsRepository, err := s.getAndConfigurePms()
		if err != nil {
			return nil, nil, err
		}

		invalid, err = pmsRepository.Delete(node, filter)
	}

	if _, found := find(usage, UseAsm); found {
		asmRepository, err := s.getAndConfigureAsm()
		if err != nil {
			return nil, nil, err
		}

		invalid2, err := asmRepository.Delete(node, filter)
		if invalid == nil && len(invalid2) > 0 {
			invalid = map[string]support.FullNameField{}
		}

		// Merge field errors from ASM with PMS errors
		for key, value := range invalid2 {
			invalid[key] = value
		}
	}

	return invalid, node, err
}
