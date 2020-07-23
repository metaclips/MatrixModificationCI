## Matrix CI Modifier

### Preface

Matrix modifier is a CI modifier. Sometimes it's useful to run the same task against different software versions. Or run different batches of tests based on an environment variable. For cases like these, the matrix modifier comes very handy. It's possible to use matrix keyword only inside of a particular task to have multiple tasks based on the original one. Each new task will be created from the original task by replacing the whole matrix YAML node with each matrix's children separately.

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

./MatrixModificationCI test.yaml result.yaml

```