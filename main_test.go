package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func BenchmarkMap(b *testing.B) {
	b.ReportAllocs()

	test := `
- matrix: foo
- key: value
`

	result := `- - foo
  - key: value
`

	for i := 0; i < b.N; i++ {
		node := nodecontent{
			node:       &yaml.Node{},
			matrixNode: make([]*matrix, 0),
		}

		if err := yaml.Unmarshal([]byte(test), node.node); err != nil {
			log.Fatalln("unable to marshal yaml text", err)
		}

		node.getMatrixNodes(node.node, node.onMatrixFound)
		node.copyMatrixToBuffer()

		data := node.moveMatrixContents()

		generatedResult, err := yaml.Marshal(data)
		if err != nil {
			b.Fatalf("unable to marshal yaml results %s", err)
		}

		assert.Equal(b, result, string(generatedResult))
	}
}

func TestFiles(t *testing.T) {
	const (
		resultPath = "testdata/results/"
		testPath   = "testdata/tests/"
	)

	getMatrix := func(filepath, filename string) error {
		yamlBytes, err := ioutil.ReadFile(filepath)
		if err != nil {
			t.Errorf("could not read yaml test file %s, err: %s", filepath, err)
			return err
		}

		node := nodecontent{
			node:       &yaml.Node{},
			matrixNode: make([]*matrix, 0),
		}

		if err := yaml.Unmarshal(yamlBytes, node.node); err != nil {
			t.Errorf("could not unmarshal yaml test file %s, err: %s", filepath, err)
			return err
		}

		node.getMatrixNodes(node.node, node.onMatrixFound)
		node.copyMatrixToBuffer()

		data := node.moveMatrixContents()

		generatedBytes, err := yaml.Marshal(data)
		if err != nil {
			t.Errorf("unable to marshal yaml results, err: %s", err)
			return err
		}

		expectedBytes, err := ioutil.ReadFile(resultPath + filename)
		if err != nil {
			t.Errorf("could not read yaml file result in results directory, err: %s", err)
			return err
		}

		if ok := assert.Equal(t, string(expectedBytes), string(generatedBytes)); !ok {
			t.Logf("results are not equal for file: %s", filename)
			fmt.Println()
			fmt.Println(string(generatedBytes))
		}
		return nil
	}

	filepath.Walk(testPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if err != nil {
			t.Errorf("error walking file path")
			return err
		}

		if err = getMatrix(path, info.Name()); err != nil {
			return err
		}

		return nil
	})
}
