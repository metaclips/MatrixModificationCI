package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"

	"gopkg.in/yaml.v3"
)

type matrix struct {
	matrix                map[string]interface{}
	topLevelMatrixContent []map[string]interface{}
	matrixContent         []interface{}
}

func main() {
	file, err := readFile()
	if err != nil {
		log.Fatalf("could not read file, error: %v", err)
	}

	matrix := matrix{
		topLevelMatrixContent: make([]map[string]interface{}, 0),
		matrix:                make(map[string]interface{})}

	err = yaml.Unmarshal(file, &matrix.matrix)
	if err != nil {
		log.Fatalf("could not unmarshal yaml, error: %v", err)
	}

	matrix.topLevelMatrixMapChecker(&matrix.matrix)
	matrix.moveMatrixContent()
	matrix.convertMapToArrayInterface()

	data := matrix.performMatrixManipulation()
	writeFile(data)
}

func readFile() ([]byte, error) {
	if len(os.Args) > 1 {
		return ioutil.ReadFile(os.Args[1])
	}

	return nil, errors.New("please pass a single file name to os.Args parameter")
}

func writeFile(data []map[string]interface{}) {
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

// Create a recursive function that gets toplevel matrix map content.
func (b *matrix) topLevelMatrixMapChecker(data *map[string]interface{}) {
	var checkInterfaceArray func([]interface{})

	checkInterfaceArray = func(v []interface{}) {
		for _, e := range v {
			switch content := e.(type) {
			case map[string]interface{}:
				b.topLevelMatrixMapChecker(&content)
			case []interface{}:
				checkInterfaceArray(content)
			}
		}
	}

	for key, value := range *data {
		switch e := value.(type) {
		case map[string]interface{}:
			b.topLevelMatrixMapChecker(&e)
		case []interface{}:
			checkInterfaceArray(e)
		}

		if key == "matrix" {
			b.topLevelMatrixContent = append(b.topLevelMatrixContent, *data)
		}
	}
}

func (b *matrix) moveMatrixContent() {
	b.matrixContent = make([]interface{}, len(b.topLevelMatrixContent))

	for i, value := range b.topLevelMatrixContent {
		b.matrixContent[i] = value["matrix"]
		delete(b.topLevelMatrixContent[i], "matrix")
	}
}

func (b *matrix) convertMapToArrayInterface() {
	for i, value := range b.matrixContent {
		switch e := value.(type) {
		case []interface{}:
			continue
		case map[string]interface{}:
			b.matrixContent[i] = convertMapToArrayInterface(e)
		default:
			log.Fatalln("unsupported format. Exiting now.")
		}
	}
}

func (b *matrix) performMatrixManipulation() []map[string]interface{} {
	nos := make([]int, len(b.matrixContent))
	mulNo := 1

	for i, value := range b.matrixContent {
		if b, ok := value.([]interface{}); ok {
			nos[i] = len(b)
			if nos[i] == 0 {
				log.Fatalln("empty interface found!!!!")
			}

			mulNo *= nos[i]
			continue
		}

		log.Fatalln("could not get matrix content interface array. Did you call convertMapToArrayInterface?.")
	}

	values := make([]int, len(nos))
	data := make([]map[string]interface{}, 0)

	for i := 0; i < mulNo; i++ {
		content := b.makeCopy(values)

		// Take for example test3.yaml where there are matrix
		// with inner matrix, this could lead to a double copy.
		// lenData := len(data)
		foundEqual := false
		for j := 0; j < len(data); j++ {
			if reflect.DeepEqual(content, data[j]) {
				foundEqual = true
				break
			}
		}

		if !foundEqual {
			data = append(data, content)
		}
		getNextIter(nos, values, 0)
	}

	return data
}

func (b matrix) makeCopy(positions []int) map[string]interface{} {
	matrixContents := make([]*map[string]interface{}, len(b.matrixContent))

	for i, pos := range positions {
		if val, ok := b.matrixContent[i].([]interface{}); ok {
			mapContent, ok := val[pos].(map[string]interface{})
			if !ok {
				log.Fatalln("error copying to top level matrix")
			}

			matrixContents[i] = &mapContent
			for key, value := range mapContent {
				b.topLevelMatrixContent[i][key] = value
			}
		}
	}

	jsonBytes, _ := json.Marshal(b.matrix)

	data := map[string]interface{}{}
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		log.Fatalln("could not covert back to json, err:", err)
	}

	for i, matrixContent := range matrixContents {
		for key := range *matrixContent {
			delete(b.topLevelMatrixContent[i], key)
		}
	}

	return data
}

func convertMapToArrayInterface(mapContent map[string]interface{}) []interface{} {
	data := make([]interface{}, 0)
	for key, value := range mapContent {
		b := map[string]interface{}{
			key: value,
		}
		data = append(data, b)
	}

	return data
}

func getNextIter(no, values []int, topLevel int) {
	if topLevel >= len(no) {
		return
	}

	if values[topLevel]+1 >= no[topLevel] {
		values[topLevel] = 0
		topLevel++
		getNextIter(no, values, topLevel)
		return
	}

	values[topLevel]++
}
