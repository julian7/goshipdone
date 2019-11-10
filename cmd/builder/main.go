// Command builder is an example CLI for executing a build pipeline
package main

import (
	"io/ioutil"
	"log"

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

	log.Printf("Result: %v", pipe)

	if err := pipe.Run(); err != nil {
		log.Printf("Error: %v", err)
	}
}
