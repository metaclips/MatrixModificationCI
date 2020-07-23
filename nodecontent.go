package main

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"log"
	"os"
	"reflect"

	"gopkg.in/yaml.v3"
)

type nodecontent struct {
	node       *yaml.Node
	matrixNode []*matrix
}

func (b *nodecontent) loadYaml() {
	if len(os.Args) > 1 {
		bytes, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			log.Fatalln("could not read file content, err:", err)
		}

		if err := yaml.Unmarshal(bytes, b.node); err != nil {
			log.Fatalln("could not unmarshal yaml file, err:", err)
		}

		return
	}

	log.Fatalln("yaml file path was not passed in parameter.")
}

// getMatrixNodes loop through nodes to get matrixes.
func (b *nodecontent) getMatrixNodes(node *yaml.Node, onMatrixFound func(*yaml.Node, int)) {
	switch node.Kind {

	case yaml.DocumentNode:
		b.loopDocumentNode(node, onMatrixFound)

	case yaml.MappingNode:
		b.loopMapNode(node, onMatrixFound)

	case yaml.SequenceNode:
		b.loopSeqNode(node, onMatrixFound)

	case yaml.ScalarNode:

	default:
		log.Fatalln("This type is not yet supported", node.Kind, node.Value)
	}
}

func (b *nodecontent) loopDocumentNode(node *yaml.Node, onMatrixFound func(*yaml.Node, int)) {
	for _, nodeContent := range node.Content {
		b.getMatrixNodes(nodeContent, onMatrixFound)
	}
}

// loopSeqNode loop nodes of type sequence.
// Sequence nodes are loop'd contents by contents unlike map where key is stored on one index and value in another.
// SeqNode contents individually be of another node type except a Scalar node.
func (b *nodecontent) loopSeqNode(node *yaml.Node, onMatrixFound func(*yaml.Node, int)) {
	for _, nodeContent := range node.Content {
		b.getMatrixNodes(nodeContent, onMatrixFound)
	}
}

// loopMapNode loop nodes with kind Map.
// Map Nodes are in array key, value pairs where key is divisble by 2.
// We first search all keys if a matrix is present before searching children nodes.
func (b *nodecontent) loopMapNode(yamlNode *yaml.Node, onMatrixFound func(*yaml.Node, int)) {
	for i := 0; i < len(yamlNode.Content); i += 2 {
		if key := yamlNode.Content[i].Value; key == "matrix" {
			onMatrixFound(yamlNode, i)
		}
	}

	for i := 0; i < len(yamlNode.Content); i += 2 {
		b.getMatrixNodes(yamlNode.Content[i+1], onMatrixFound)
	}
}

// onMatrixFound copies matrix contents and top level nodes to a matrix struct.
func (b *nodecontent) onMatrixFound(toplevelNode *yaml.Node, index int) {
	matrixContent := toplevelNode.Content[index+1]

	var matrixContentCount int

	if matrixContent.Kind == yaml.MappingNode {
		matrixContentCount = (len(matrixContent.Content) + 1) / 2
	} else {
		matrixContentCount = len(matrixContent.Content)
	}

	// matrixContentCount is zero indexed.
	matrixContentCount--
	if matrixContentCount < 0 {
		matrixContentCount = 0
	}

	matrixNode := &matrix{
		TopLevelMatrixContent: toplevelNode,
		matrixContent:         matrixContent,

		topLevelMatrixLocation: uint(index),
		matrixContentCount:     uint(matrixContentCount), // Amount of matrix content
	}

	b.matrixNode = append(b.matrixNode, matrixNode)
}

// copyMatrixToBuffer copies all matrix contents to their buffer.
// Note: most buffers might be redundant as nested matrixes when recopied are no longer associated.
// Given the matrix below, since nodes are connected using matrixes, if we copy the stored buffers for the two matrixes,
// the two nodes will no longer be connected. recopyMatrixContents fixes this by ignoring the copues.
//
//  matrix:
// 	  - image: node:latest
//	  - matrix:
func (b *nodecontent) copyMatrixToBuffer() {
	for i := 0; i < len(b.matrixNode); i++ {
		matrixNode := b.matrixNode[i]

		var matrixBuf bytes.Buffer
		enc := gob.NewEncoder(&matrixBuf)
		if err := enc.Encode(matrixNode.TopLevelMatrixContent); err != nil {
			log.Fatalln("failed to encode", err)
		}

		matrixNode.copiedMatrixBuffer = matrixBuf

		i += int(matrixNode.linkedMatrixCount)
	}
}

// moveMatrixContents move matrix contents to its top level node.
// Matrix contents are picked in increasing order and moved in backward transition.
// Given the below matrix
//
//  matrix:
// 	  image: node:latest
//	  matrix:
//		work-dir: path
//
// if work-dir map need to be on the first node, it is first moved to its child matrix,
// then moved to the parent node.
func (b *nodecontent) moveMatrixContents() []yaml.Node {
	parsedYAMLs := make([]yaml.Node, 0)
	parsedData := make([][]byte, 0)

	possibleCombinations := make([]uint, len(b.matrixNode))

	shouldIter := true
	for shouldIter {
		// matrix contents are picked and moved in backward transition.
		for i := len(possibleCombinations) - 1; i >= 0; i-- {
			b.matrixNode[i].makeConversion(possibleCombinations[i])
		}

		bytes, err := yaml.Marshal(&b.node.Content[0])
		if err != nil {
			log.Fatalln("could not marshal file", err)
		}

		b.recopyMatrixContents()

		equal := false
		for _, g := range parsedData {
			if equal = reflect.DeepEqual(g, bytes); equal {
				break
			}
		}

		if !equal {
			parsedData = append(parsedData, bytes)
		}

		shouldIter = b.shouldIter(possibleCombinations)
	}

	for _, bytes := range parsedData {
		var node yaml.Node
		if err := yaml.Unmarshal(bytes, &node); err != nil {
			log.Fatalln("unable to unmarshal file")
		}

		parsedYAMLs = append(parsedYAMLs, *node.Content[0])
	}

	return parsedYAMLs
}

// recopyMatrixContents copies matrix contents stored in buffer back to toplevel matrix.
// Most buffers are neglected if a certain buffer stores nested matrix content.
func (b *nodecontent) recopyMatrixContents() {
	for i := 0; i < len(b.matrixNode); i++ {
		if len(b.matrixNode[i].copiedMatrixBuffer.Bytes()) == 0 {
			log.Fatalln("matrix node was not copied")
		}

		b.matrixNode[i].loadBuffer()

		resetLinkedMatrixPointers := func(toplevelNode *yaml.Node, index int) {
			matrixContent := toplevelNode.Content[index+1]
			b.matrixNode[i].TopLevelMatrixContent = toplevelNode
			b.matrixNode[i].matrixContent = matrixContent
			// iterate loop if more matrix nodes are found, also copying the new pointers.
			i++
		}

		b.getMatrixNodes(b.matrixNode[i].TopLevelMatrixContent, resetLinkedMatrixPointers)
	}
}

// shouldIter iterates position till possible combinations are exhausted. Iteration is done in increasing order.
func (b *nodecontent) shouldIter(pos []uint) bool {
	for i := len(pos) - 1; i >= 0; i-- {
		// Iterate from most significant bit side.
		if pos[i] == b.matrixNode[i].matrixContentCount {
			return b.getNextShiftPos(pos, uint(i))
		}

		pos[i]++
		return true
	}

	return false
}

func (b *nodecontent) getNextShiftPos(pos []uint, index uint) bool {
	for i := len(pos) - 1; i >= 0; i-- {
		if pos[i] == b.matrixNode[i].matrixContentCount {
			pos[i] = 0
			continue
		}

		pos[i]++
		return true
	}

	return false
}
