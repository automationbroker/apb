package jsval_test

import (
	"strings"
	"testing"

	"github.com/lestrrat/go-jsschema"
	"github.com/lestrrat/go-jsval"
	"github.com/lestrrat/go-jsval/builder"
	"github.com/stretchr/testify/assert"
)

func TestStringFromSchema(t *testing.T) {
	const src = `{
  "type": "string",
  "maxLength": 15,
  "minLength": 5,
  "default": "Hello, World!"
}`

	s, err := schema.Read(strings.NewReader(src))
	if !assert.NoError(t, err, "schema.Read should succeed") {
		return
	}

	b := builder.New()
	v, err := b.Build(s)
	if !assert.NoError(t, err, "Builder.Build should succeed") {
		return
	}

	c2 := jsval.String()
	c2.Default("Hello, World!").MaxLength(15).MinLength(5)
	if !assert.Equal(t, c2, v.Root(), "constraints are equal") {
		return
	}
}

func TestString(t *testing.T) {
	var s string
	c := jsval.String()
	c.Default("Hello, World!").MaxLength(15)

	if !assert.True(t, c.HasDefault(), "HasDefault is true") {
		return
	}

	if !assert.Equal(t, c.DefaultValue(), "Hello, World!", "DefaultValue returns expected value") {
		return
	}

	if !assert.NoError(t, c.Validate(s), "validate should succeed") {
		return
	}

	c.MinLength(5)
	if !assert.Error(t, c.Validate(s), "validate should fail") {
		return
	}

	s = "Hello"
	if !assert.NoError(t, c.Validate(s), "validate should succeed") {
		return
	}
}