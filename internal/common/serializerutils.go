package common

import (
	"github.com/mariotoffia/ssm.git/internal/reflectparser"
	"github.com/mariotoffia/ssm.git/internal/tagparser"
	"github.com/mariotoffia/ssm.git/support"
)

// NodesToParameterMap grabs all FullNames on nodes that do have tag set
// in order to get data fom parameter store. Note that it chcks for the
// tag SsmType = st. The full name is the associated with the node itself.
// This is to gain a more accessable structure to seach for nodes.
func NodesToParameterMap(node *reflectparser.SsmNode,
	paths map[string]*reflectparser.SsmNode, filter *support.FieldFilters, st tagparser.StoreType) bool {
	issecure := false
	if node.HasChildren() {
		for _, n := range node.Children() {
			if NodesToParameterMap(&n, paths, filter, st) {
				issecure = true
			}
		}
	} else {
		if node.Tag().SsmType() == st {
			if filter.IsIncluded(node.FqName()) {
				paths[node.Tag().FullName()] = node
				if node.Tag().Secure() {
					issecure = true
				}
			}
		}
	}

	return issecure
}
