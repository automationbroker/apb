package jsval_test

import "testing"

func TestGenerated(t *testing.T) {
	err := V0.Validate(map[string]interface{}{
		"minItems": -1,
	})
	if err == nil {
		t.Errorf("Validation failed: %s", err)
	}
}
