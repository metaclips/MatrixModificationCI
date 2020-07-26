package main

import (
	"bytes"
	"encoding/gob"
	"log"

	"gopkg.in/yaml.v3"
)

func init() {
	gob.Register(&yaml.Node{})
}

type matrix struct {
	copiedMatrixBuffer bytes.Buffer

	// task
	// 	matrix:
	//	  - name: Lint
	//      lint_script: yarn run lint
	//
	// Given the yaml as above, toplevel matrix content contains all yaml content from task,
	// while matrix content are yaml content that contains matrix values.
	// Only Toplevel matrix content is copied and saved to buffers, linked matrix are not saved to
	// buffer but re-linked.
	TopLevelMatrixContent *yaml.Node
	matrixContent         *yaml.Node

	topLevelMatrixLocation uint // matrix location at the top level
	matrixContentCount     uint // amount of matrix count

	linkedMatrixCount uint
}

// makeConversion moves matrix content for varies matrix content types and append to top level matrix content.
// No matter the node type, for matrix modification, we are always appending to a map, if for a sequence type given as below
//
// - matrix: contents
//
// At the toplevel the node type is a sequence which contains only a map of matrix key and contents.
// If we are to move contents which is a scalar or any other type, we would append to the toplevel which is a sequence.
// If the toplevel is a map, say,
//
//foo:
//	matrix: content
//
// We are to append "contents" which could be of any type and append to toplevel matrix key-val (foo value).
func (b *matrix) makeConversion(pos uint) {
	kind := b.matrixContent.Kind

	switch kind {

	case yaml.MappingNode:
		b.getMapNodeTypes(b.matrixContent, pos)

	case yaml.SequenceNode:
		b.getSequenceNodeTypes(b.matrixContent, pos)

	case yaml.ScalarNode:
		b.getScalarNodeTypes(b.matrixContent)

	default:
		log.Fatalln("unsupported type")
	}
}

// getSequenceNodeTypes moves sequence node types to toplevel matrix node.
// If nested sequence matrix node which is to be moved is a map type, the map contents for the matrix index is moved.
// else it's assumed to either be scalar node type or a sequence type itself and it's contents is moved.
func (b *matrix) getSequenceNodeTypes(node *yaml.Node, index uint) {
	matrixContent := node.Content[index]

	switch matrixContent.Kind {

	case yaml.MappingNode:
		b.TopLevelMatrixContent.Content = append(b.TopLevelMatrixContent.Content[:b.topLevelMatrixLocation],
			append(matrixContent.Content, b.TopLevelMatrixContent.Content[b.topLevelMatrixLocation+2:]...)...)

	default:
		*b.TopLevelMatrixContent = *matrixContent
	}
}

// getScalarNodeTypes moves scalar node type to top level.
// If a matrix content is a scalar node type, it is assumed that the matrix toplevel content is a sequence type
// or it's the only node content else could give unpredicatable output. e.g.
//
// foo:
//   - matrix:
//     - name
//   - works: true
//
// below wont work.
//
// foo:
//   matrix:
//     - name
//   works: true
//
// as it'll give the below output.
//
// foo:	name
func (b *matrix) getScalarNodeTypes(node *yaml.Node) {
	*b.TopLevelMatrixContent = *node
}

// getMapNodeTypes moves map node type to top level.
// Map contents are moved key by value.
func (b *matrix) getMapNodeTypes(node *yaml.Node, index uint) {
	index *= 2

	b.TopLevelMatrixContent.Content = append(b.TopLevelMatrixContent.Content[:b.topLevelMatrixLocation],
		append([]*yaml.Node{node.Content[index], node.Content[index+1]}, b.TopLevelMatrixContent.Content[b.topLevelMatrixLocation+2:]...)...)
}

func (b *matrix) loadBuffer() {
	newCopy := b.copiedMatrixBuffer

	dec := gob.NewDecoder(&newCopy)

	node := &yaml.Node{}
	if err := dec.Decode(node); err != nil {
		log.Fatalln("unable to decode gob", err)
	}

	*b.TopLevelMatrixContent = *node
}
