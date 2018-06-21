package builder

import (
	"strings"
	"testing"

	"github.com/lestrrat/go-jsschema"
	"github.com/stretchr/testify/assert"
)

func TestGuessType(t *testing.T) {
	data := map[string]schema.PrimitiveTypes{
		`{ "items": { "type": "number" } }`:                 {schema.ArrayType},
		`{ "properties": { "foo": { "type": "number" } } }`: {schema.ObjectType},
		`{ "additionalProperties": false }`:                 {schema.ObjectType},
		`{ "maxLength": 10 }`:                               {schema.StringType},
		`{ "pattern": "^[a-fA-F0-9]+$" }`:                   {schema.StringType},
		`{ "format": "email" }`:                             {schema.StringType},
		`{ "multipleOf": 1 }`:                               {schema.IntegerType},
		`{ "multipleOf": 1.1 }`:                             {schema.NumberType},
		`{ "minimum": 0, "pattern": "^[a-z]+$" }`:           {schema.StringType, schema.IntegerType},
	}

	for src, expected := range data {
		t.Logf("Testing '%s'", src)
		s, err := schema.Read(strings.NewReader(src))
		if !assert.NoError(t, err, "schema.Read should succeed") {
			return
		}

		if !assert.Equal(t, guessSchemaType(s), expected, "types match") {
			return
		}
	}
}