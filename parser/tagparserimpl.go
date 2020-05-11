package parser

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

type structTag struct {
	named  map[string]string
	tags   map[string]string
	fqname string
}

func (s *structTag) Named() map[string]string { return s.named }
func (s *structTag) Tags() map[string]string  { return s.tags }
func (s *structTag) Name() string             { return s.named["name"] }
func (s *structTag) FullName() string         { return s.fqname }

type tagParserImpl struct {
}

func newTagParser() TagParser {
	return &tagParserImpl{}
}

func (p *tagParserImpl) ParseTagString(tagstring string,
	prefix string,
	env string,
	svc string) (StructTag, error) {

	fmt.Printf("parsing %s\n", tagstring)
	st := &structTag{
		named: map[string]string{"prefix": RenderPrefix(prefix, env, svc)},
		tags:  map[string]string{},
	}

	if len(tagstring) == 0 {
		return st, nil
	}

	commas := strings.Split(tagstring, ",")
	for _, kvs := range commas {
		kv := strings.Split(kvs, "=")
		kv[0] = strings.ToLower(strings.TrimSpace(kv[0]))

		if len(kv) == 1 {
			if _, ok := st.named["name"]; ok {
				return nil, errors.Errorf("Multiple non key value in tag '%s'", tagstring)
			}
			kv[1] = kv[0]
			kv[0] = "name"
		}

		switch kv[0] {
		case "name":
			st.named["name"] = kv[1]
		case "prefix":
			st.named["prefix"] = RenderPrefix(kv[1], env, "")
			break
		default:
			st.tags[kv[0]] = kv[1]
		}
	}

	if _, ok := st.named["name"]; ok {
		st.fqname = RenderPrefix(prefix, env, svc) + "/" + st.named["name"]
	}

	return st, nil
}

// RenderPrefix renders a prefix based on the inparam strings
func RenderPrefix(prefix string, env string, svc string) string {
	if strings.HasPrefix(env, "/") {
		env = env[1:]
	}
	if strings.HasSuffix(env, "/") {
		env = env[:1]
	}
	if strings.HasPrefix(svc, "/") {
		svc = svc[1:]
	}
	if strings.HasSuffix(svc, "/") {
		svc = svc[:1]
	}
	if prefix != "" && !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if strings.HasSuffix(prefix, "/") {
		prefix = prefix[:1]
	}

	if prefix == "" {
		return fmt.Sprintf("/%s/%s", env, svc)
	}
	if svc == "" {
		return fmt.Sprintf("/%s%s", env, prefix)
	}
	return fmt.Sprintf("/%s/%s%s", env, svc, prefix)
}
