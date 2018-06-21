package jsval_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lestrrat/go-jsschema"
	"github.com/lestrrat/go-jsval"
	"github.com/lestrrat/go-jsval/builder"
	"github.com/stretchr/testify/assert"
)

// test against github#2
func TestArrayItemsReference(t *testing.T) {
	const src = `{
	"definitions": {
	  "uint": {
			"type": "integer",
			"minimum": 0
		}
	},
	"type": "object",
	"properties": {
		"numbers": {
			"type": "array",
			"items": { "$ref": "#/definitions/uint" }
		},
		"tuple": {
			"items": [ { "type": "string" }, { "type": "boolean" }, { "type": "number" } ]
		}
	}
}`
	s, err := schema.Read(strings.NewReader(src))
	if !assert.NoError(t, err, "schema.Reader should succeed") {
		return
	}

	b := builder.New()
	v, err := b.Build(s)
	if !assert.NoError(t, err, "builder.Build should succeed") {
		return
	}

	buf := bytes.Buffer{}
	g := jsval.NewGenerator()
	if !assert.NoError(t, g.Process(&buf, v), "Generator.Process should succeed") {
		return
	}

	code := buf.String()
	if !assert.True(t, strings.Contains(code, "\tItems("), "Generated code chould contain `.Items()`") {
		return
	}

	if !assert.True(t, strings.Contains(code, "\tPositionalItems("), "Generated code should contain `.PositionalItems()`") {
		return
	}
}