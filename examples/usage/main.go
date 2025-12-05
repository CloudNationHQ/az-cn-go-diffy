package main

import (
	"log"

	"github.com/cloudnationhq/az-cn-go-diffy"
)

func main() {
	findings, err := diffy.ValidateSchema(
		diffy.WithTerraformRoot("../module"),
		func(opts *diffy.SchemaValidatorOptions) {
			opts.Silent = false
		},
	)
	if err != nil {
		log.Fatalf("validation failed: %v", err)
	}

	if len(findings) == 0 {
		log.Println("No validation findings.")
		return
	}

	for _, finding := range findings {
		log.Println(diffy.FormatFinding(finding))
	}
}
