package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

func main() {
	file, err := readFile()
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	m := make(map[interface{}]interface{})
	err = yaml.UnmarshalStrict(file, &m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	matrixData := make(map[interface{}]interface{})
	content, matrix := recursiveMapChecker(&m, &matrixData)
	contents := allocateMatrixContent(&content, matrix, &matrixData)

	d, err := yaml.Marshal(&contents)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("\n%s\n\n", string(d))
}

func readFile() ([]byte, error) {
	var file []byte
	if len(os.Args) == 2 {
		var err error
		file, err = ioutil.ReadFile(os.Args[1])
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("Please pass a single file name to os.Args parameter")
	}

	return file, nil
}

// Create a recursive function that gets matrix map content.
func recursiveMapChecker(data *map[interface{}]interface{}, pointedData *map[interface{}]interface{}) (map[interface{}]interface{}, interface{}) {
	config := make(map[interface{}]interface{})
	var matrix interface{}

	for key, value := range *data {
		if key, ok := key.(string); ok && key == "matrix" {
			// Get pointer to map interface.
			// Note: this is assuming index of matrix i.e [key]matrix
			// is not an array.
			config = *pointedData
			matrix = value
		} else if cvalue, ok := value.(map[interface{}]interface{}); ok {
			config[key], matrix = recursiveMapChecker(&cvalue, pointedData)
		} else {
			config[key] = value
		}
	}

	return config, matrix
}

func allocateMatrixContent(content *map[interface{}]interface{}, matrix interface{}, pointedData *map[interface{}]interface{}) []map[interface{}]interface{} {
	if data, ok := matrix.([]interface{}); ok {
		contents := make([]map[interface{}]interface{}, len(data))

		for i, values := range data {

			values, ok := values.(map[interface{}]interface{})
			if ok {
				for key, value := range values {
					contents[i] = make(map[interface{}]interface{})
					for k := range *pointedData {
						delete(*pointedData, k)
					}
					(*pointedData)[key] = value
					contents[i] = CopyMap(*content)
					fmt.Println(contents)
				}
			}
		}
		return contents

	} else if data, ok := matrix.(map[interface{}]interface{}); ok {
		contents := make([]map[interface{}]interface{}, len(data))

		var counter int
		for _, values := range data {

			values, ok := values.(map[interface{}]interface{})
			if ok {
				for key, value := range values {
					contents[counter] = make(map[interface{}]interface{})
					for k := range *pointedData {
						delete(*pointedData, k)
					}

					(*pointedData)[key] = value
					contents[counter] = CopyMap(*content)
					counter++
				}
			}
		}
	}

	return nil
}

func CopyMap(m map[interface{}]interface{}) map[interface{}]interface{} {
	content := make(map[interface{}]interface{})
	for k, value := range m {
		vm, ok := value.(map[interface{}]interface{})
		if ok {
			content[k] = CopyMap(vm)
		} else {
			content[k] = value
		}
	}

	return content
}
