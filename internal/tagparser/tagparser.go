package tagparser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// ParseAsmTagString parses the inparam tag string s into a SsmTag interface
func ParseAsmTagString(s string, prefix string, env string, svc string) (SsmTag, error) {
	if len(s) == 0 {
		return nil, errors.Errorf("tag string cannot be empty")
	}

	tag := AsmTag{prefix: renderPrefix(prefix, env, svc), tags: map[string]string{}}
	commas := strings.Split(s, ",")
	for _, kvs := range commas {
		kv := strings.Split(kvs, "=")
		kv[0] = strings.TrimSpace(kv[0])

		if len(kv) == 1 {
			if tag.name != "" {
				return nil, errors.Errorf("already got a name %s and cannot overwrite it with %s - tag: %s",
					tag.name, kv[0], s)
			}
			tag.name = kv[0]
		} else {
			// key = value
			kv[1] = strings.TrimSpace(kv[1])
			switch kv[0] {
			case "name":
				if tag.name != "" {
					return nil, errors.Errorf("already got a name %s and cannot overwrite it with %s - tag: %s",
						tag.name, kv[0], s)
				}
				tag.name = kv[1]
			case "prefix":
				tag.prefix = renderPrefix(kv[1], env, "")
			case "keyid":
				tag.keyID = kv[1]
			case "vid":
				tag.versionID = kv[1]
			case "vs":
				tag.versionStage = kv[1]
			default:
				tag.tags[kv[0]] = kv[1]
			}
		}
	}

	if tag.name == "" {
		return nil, errors.Errorf("No name specified in tag %s", s)
	}

	return &tag, nil
}

// ParsePmsTagString parses the inparam tag string s into a SsmTag interface
func ParsePmsTagString(s string, prefix string, env string, svc string) (SsmTag, error) {
	if len(s) == 0 {
		return nil, errors.Errorf("tag string cannot be empty")
	}

	tag := PmsTag{prefix: renderPrefix(prefix, env, svc), tags: map[string]string{}, overwrite: true}
	commas := strings.Split(s, ",")
	for _, kvs := range commas {
		kv := strings.Split(kvs, "=")
		kv[0] = strings.TrimSpace(kv[0])

		if len(kv) == 1 {
			if tag.name != "" {
				return nil, errors.Errorf("already got a name %s and cannot overwrite it with %s - tag: %s",
					tag.name, kv[0], s)
			}
			tag.name = kv[0]
		} else {
			// key = value
			kv[1] = strings.TrimSpace(kv[1])
			switch kv[0] {
			case "name":
				if tag.name != "" {
					return nil, errors.Errorf("already got a name %s and cannot overwrite it with %s - tag: %s",
						tag.name, kv[0], s)
				}
				tag.name = kv[1]
			case "keyid":
				tag.keyID = kv[1]
			case "prefix":
				tag.prefix = renderPrefix(kv[1], env, "")
			case "description":
				tag.description = kv[1]
			case "pattern":
				tag.pattern = kv[1]
			case "overwrite":
				tag.overwrite, _ = strconv.ParseBool(kv[1])
			default:
				tag.tags[kv[0]] = kv[1]
			}
		}
	}

	if tag.name == "" {
		return nil, errors.Errorf("No name specified in tag %s", s)
	}

	return &tag, nil
}

func renderPrefix(prefix string, env string, svc string) string {
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
