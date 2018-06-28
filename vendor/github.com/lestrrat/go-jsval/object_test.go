package jsval_test

import (
	"strings"
	"testing"

	"github.com/lestrrat/go-jsschema"
	"github.com/lestrrat/go-jsval/builder"
	"github.com/stretchr/testify/assert"
)

func TestObject(t *testing.T) {
	const src = `{
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "name": {
      "type": "string",
      "maxLength": 20,
      "pattern": "^[a-z ]+$"
    },
	  "age": {
		  "type": "integer",
	    "minimum": 0
	  },
	  "tags": {
      "type": "array",
	    "items": {
        "type": "string"
      }
    }
  },
	"patternProperties": {
		"name#[a-z]+": {
			"type": "string"
		}
	}
}`

	s, err := schema.Read(strings.NewReader(src))
	if !assert.NoError(t, err, "reading schema should succeed") {
		return
	}

	b := builder.New()
	v, err := b.Build(s)
	if !assert.NoError(t, err, "Builder.Build should succeed") {
		return
	}

	data := []interface{}{
		map[string]interface{}{"Name": "World"},
		map[string]interface{}{"name": "World"},
		map[string]interface{}{"name": "wooooooooooooooooooooooooooooooorld"},
		map[string]interface{}{
			"tags": []interface{}{ 1, "foo", false },
		},
		map[string]interface{}{"name": "ハロー、ワールド"},
		map[string]interface{}{"foo#ja": "フー！"},
	}
	for _, input := range data {
		t.Logf("Testing %#v (should FAIL)", input)
		if !assert.Error(t, v.Validate(input), "validation fails") {
			return
		}
	}

	data = []interface{}{
		map[string]interface{}{"name": "world"},
		map[string]interface{}{"tags": []interface{}{"foo", "bar", "baz"}},
		map[string]interface{}{"name#ja": "ハロー、ワールド"},
	}
	for _, input := range data {
		t.Logf("Testing %#v (should PASS)", input)
		if !assert.NoError(t, v.Validate(input), "validation passes") {
			return
		}
	}
}

func TestObjectDependency(t *testing.T) {
	const src = `{
  "type": "object",
  "additionalProperties": false,
  "properties": {
	  "foo": { "type": "string" },
	  "bar": { "type": "string" }
  },
  "dependencies": {
	  "foo": ["bar"]
  }
}`

	s, err := schema.Read(strings.NewReader(src))
	if !assert.NoError(t, err, "reading schema should succeed") {
		return
	}

	b := builder.New()
	v, err := b.Build(s)
	if !assert.NoError(t, err, "Builder.Build should succeed") {
		return
	}

	data := []interface{}{
		map[string]interface{}{"foo": "foo"},
	}
	for _, input := range data {
		t.Logf("Testing %#v (should FAIL)", input)
		if !assert.Error(t, v.Validate(input), "validation fails") {
			return
		}
	}

	data = []interface{}{
		map[string]interface{}{"foo": "foo", "bar": "bar"},
	}
	for _, input := range data {
		t.Logf("Testing %#v (should PASS)", input)
		if !assert.NoError(t, v.Validate(input), "validation passes") {
			return
		}
	}
}

func TestObjectSchemaDependency(t *testing.T) {
	const src =`{
  "type": "object",

  "properties": {
    "name": { "type": "string" },
    "credit_card": { "type": "string", "pattern": "^[0-9]+$" }
  },

  "required": ["name"],

  "dependencies": {
    "credit_card": {
      "properties": {
        "billing_address": { "type": "string" }
      },
      "required": ["billing_address"]
    }
  }
}`

	s, err := schema.Read(strings.NewReader(src))
	if !assert.NoError(t, err, "reading schema should succeed") {
		return
	}

	b := builder.New()
	v, err := b.Build(s)
	if !assert.NoError(t, err, "Builder.Build should succeed") {
		return
	}

  data := []interface{}{
    map[string]interface{}{
			"name": "John Doe",
		  "credit_card": "5555555555555555",
		},
  }
  for _, input := range data {
    t.Logf("Testing %#v (should FAIL)", input)
    if !assert.Error(t, v.Validate(input), "validation fails") {
      return
    }
  }

	data = []interface{}{
		map[string]interface{}{
		  "name": "John Doe",
		  "credit_card": "5555555555555555",
		  "billing_address": "555 Debtor's Lane",
		},
		map[string]interface{}{
		  "name": "John Doe",
		  "billing_address": "555 Debtor's Lane",
		},
	}
	for _, input := range data {
		t.Logf("Testing %#v (should PASS)", input)
		if !assert.NoError(t, v.Validate(input), "validation passes") {
			return
		}
	}
}

