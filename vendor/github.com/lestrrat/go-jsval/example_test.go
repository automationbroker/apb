package jsval_test

import (
	"log"

	"github.com/lestrrat/go-jsschema"
	"github.com/lestrrat/go-jsval"
	"github.com/lestrrat/go-jsval/builder"
)

func ExampleBuild() {
	s, err := schema.ReadFile(`/path/to/schema.json`)
	if err != nil {
		log.Printf("failed to open schema: %s", err)
		return
	}

	b := builder.New()
	v, err := b.Build(s)
	if err != nil {
		log.Printf("failed to build validator: %s", err)
		return
	}

	var input interface{}
	if err := v.Validate(input); err != nil {
		log.Printf("validation failed: %s", err)
		return
	}
}

func ExampleManual() {
	v := jsval.Object().
		AddProp(`zip`, jsval.String().RegexpString(`^\d{5}$`)).
		AddProp(`address`, jsval.String()).
		AddProp(`name`, jsval.String()).
		AddProp(`phone_number`, jsval.String().RegexpString(`^[\d-]+$`)).
		Required(`zip`, `address`, `name`)

	var input interface{}
	if err := v.Validate(input); err != nil {
		log.Printf("validation failed: %s", err)
		return
	}
}