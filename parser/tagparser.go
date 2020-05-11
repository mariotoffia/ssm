package parser

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

type tagParser struct {
	named []string
}

// NewTagParser creates a new default tag parser
func NewTagParser(named []string) TagParser {
	return &tagParser{named: named}
}

func (p *tagParser) ParseTagString(tagstring string,
	prefix string,
	env string,
	svc string) (StructTag, error) {

	fmt.Printf("parsing %s\n", tagstring)
	st := &StructTagImpl{
		Named: map[string]string{"prefix": RenderPrefix(prefix, env, svc)},
		Tags:  map[string]string{},
	}

	if len(tagstring) == 0 {
		return st, nil
	}

	commas := strings.Split(tagstring, ",")
	for _, kvs := range commas {
		kv := strings.Split(kvs, "=")
		kv[0] = strings.ToLower(strings.TrimSpace(kv[0]))

		if len(kv) == 1 {
			if _, ok := st.Named["name"]; ok {
				return nil, errors.Errorf("Multiple non key value in tag '%s'", tagstring)
			}
			tmp := kv[0]
			kv = []string{"name", tmp}
		}

		switch kv[0] {
		case "name":
			st.Named["name"] = kv[1]
		case "prefix":
			st.Named["prefix"] = RenderPrefix(kv[1], env, "")
			break
		default:
			if stringInSlice(kv[0], p.named) {
				st.Named[kv[0]] = kv[1]
			} else {
				st.Tags[kv[0]] = kv[1]
			}
		}
	}

	if _, ok := st.Named["name"]; ok {
		st.FullName = fmt.Sprintf("%s/%s", st.Named["prefix"], st.Named["name"])
	}

	return st, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
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
