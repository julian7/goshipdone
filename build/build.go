// build command for running goshipdone for testing purposes
package main

import (
	"github.com/julian7/goshipdone"
	"log"
)

func main() {
	if err := goshipdone.Run(""); err != nil {
		log.Fatalln(err)
	}
}
