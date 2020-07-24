package main

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestMatrixNodeCount(t *testing.T) {
	test := `
- matrix: foo
- key: value
`

	node := nodecontent{
		node:       &yaml.Node{},
		matrixNode: make([]*matrix, 0),
	}

	if err := yaml.Unmarshal([]byte(test), node.node); err != nil {
		log.Fatalln("unable to marshal yaml text", err)
	}

	node.getMatrixNodes(node.node, node.onMatrixFound)

	assert.Equal(t, 1, len(node.matrixNode))
}
