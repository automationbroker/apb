package jsval_test

import (
	"testing"

	"github.com/lestrrat/go-jsval"
)

func TestSanity(t *testing.T) {
	t.Run("MaybeBool", func(t *testing.T) {
		var v jsval.Maybe
		v = &jsval.MaybeBool{}
		_ = v
	})
	t.Run("MaybeFloat", func(t *testing.T) {
		var v jsval.Maybe
		v = &jsval.MaybeFloat{}
		_ = v
	})
	t.Run("MaybeInt", func(t *testing.T) {
		var v jsval.Maybe
		v = &jsval.MaybeInt{}
		_ = v
	})
	t.Run("MaybeString", func(t *testing.T) {
		var v jsval.Maybe
		v = &jsval.MaybeString{}
		_ = v
	})
	t.Run("MaybeTime", func(t *testing.T) {
		var v jsval.Maybe
		v = &jsval.MaybeTime{}
		_ = v
	})
	t.Run("MaybeUint", func(t *testing.T) {
		var v jsval.Maybe
		v = &jsval.MaybeUint{}
		_ = v
	})
}
