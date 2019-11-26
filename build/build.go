// build command for running goshipdone for testing purposes
package main

import (
	"flag"
	"log"
	"os"

	"github.com/julian7/goshipdone"
)

func main() {
	publish := flag.Bool("publish", false, "run publish phase (default: false)")
	flag.Parse()

	if *publish {
		os.Setenv("SKIP_PUBLISH", "false")
	}

	if err := goshipdone.Run(""); err != nil {
		log.Fatalln(err)
	}
}
