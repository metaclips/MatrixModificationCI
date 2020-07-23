package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

func main() {
	node := nodecontent{
		node:       &yaml.Node{},
		matrixNode: make([]*matrix, 0),
	}

	node.loadYaml()
	node.getMatrixNodes(node.node, node.onMatrixFound)
	node.copyMatrixToBuffer()

	data := node.moveMatrixContents()
	writeFile(data)
}

func writeFile(data []yaml.Node) {
	d, err := yaml.Marshal(&data)
	if err != nil {
		log.Fatalf("could not marshal yaml file, error: %v", err)
	}

	if len(os.Args) > 2 {
		if err := ioutil.WriteFile(os.Args[2], d, 0644); err != nil {
			log.Fatalln("error writing to file:", err)
		}
		return
	}

	fmt.Printf("\n%s\n\n", string(d))
}
