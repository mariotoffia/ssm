package ssm

import (
	"reflect"

	"github.com/pkg/errors"
)

// https://medium.com/capital-one-tech/learning-to-use-go-reflection-822a0aed74b7
func create(node *ssmNode) (reflect.Value, error) {
	if !node.IsRoot() {
		return reflect.Value{}, errors.Errorf("Only root node may be created from this function")
	}

	for _, n := range node.childs {
		createInternal(&n)
	}

	return node.v, nil
}

func createInternal(node *ssmNode) {
	if node.f.Type.Kind() == reflect.Struct {
	}

	for _, n := range node.childs {
		if n.HasChildren() {
			createInternal(&n)
		}
	}
}
