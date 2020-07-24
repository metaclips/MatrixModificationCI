## Matrix CI Modifier
[![Build Status](https://api.cirrus-ci.com/github/metaclips/MatrixModificationCI.svg)](https://cirrus-ci.com/github/metaclips/MatrixModificationCI)
![Github CI](https://github.com/metaclips/MatrixModificationCI/workflows/Test/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/metaclips/MatrixModificationCI)](https://goreportcard.com/report/github.com/metaclips/MatrixModificationCI)


### Preface

Matrix modifier is a fast CI modifier which splits matrix contents into multiple nodes. [More here](https://cirrus-ci.org/guide/writing-tasks/#matrix-modification)


## Examples

Say we have a task 

```bash

foo:
    matrix:
      - name: bar
      - name: baz

```

This is converted to using matrix modifier.

```bash

- foo:
    name: bar

- foo:
    name: baz

```

```bash

container:
  image: node:latest

task:
  node_modules_cache:
    folder: node_modules
    fingerprint_script: cat yarn.lock
    populate_script: yarn install

  matrix:
    - name: Lint
      lint_script: yarn run lint
    - name: Test
      container:
        matrix:
          - image: node:latest
          - image: node:lts
      test_script: yarn run test
    - name: Publish
      depends_on:
        - Lint
        - Test
      only_if: $BRANCH == "master"
      publish_script: yarn run publish

```

Converted to 

```bash

- container:
    image: node:latest
    task:
        - node_modules_cache:
          folder: node_modules
          fingerprint_script: cat yarn.lock
          populate_script: yarn install
        - name: Lint
          lint_script: yarn run lint
- container:
    image: node:latest
    task:
        - node_modules_cache:
          folder: node_modules
          fingerprint_script: cat yarn.lock
          populate_script: yarn install
        - name: Test
          container:
            - image: node:latest
          test_script: yarn run test
- container:
    image: node:latest
    task:
        - node_modules_cache:
          folder: node_modules
          fingerprint_script: cat yarn.lock
          populate_script: yarn install
        - name: Test
          container:
            - image: node:lts
          test_script: yarn run test
- container:
    image: node:latest
    task:
        - node_modules_cache:
          folder: node_modules
          fingerprint_script: cat yarn.lock
          populate_script: yarn install
        - name: Publish
          depends_on:
            - Lint
            - Test
          only_if: $BRANCH == "master"
          publish_script: yarn run publish


```

## Installation

### build
```bash

git clone https://github.com/metaclips/MatrixModificationCI

cd MatrixModificationCI

go build

```

### Execution

Pass yaml file to be read as first argument. To output to a file, pass result file name as second argument.

If for example you want to read a test.yaml file and pass result to result.yaml.

```bash

MatrixModificationCI test.yaml result.yaml

```