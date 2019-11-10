// example CLI for executing a build pipeline
package main

import (
	"fmt"
	"io/ioutil"

	"github.com/julian7/magelib/pipeline"
)

func main() {
	content, err := ioutil.ReadFile("./.pipeline.yml")
	if err != nil {
		panic(err)
	}

	pipe, err := pipeline.LoadBuildPipeline(content)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Result: %v\n", pipe)

	if err := pipe.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
