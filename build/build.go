// build command for running goshipdone for testing purposes
package main

import (
	"log"

	"github.com/julian7/goshipdone"
)

func main() {
	if err := goshipdone.Run(""); err != nil {
		log.Fatalln(err)
	}
}
