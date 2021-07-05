package configuration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBool(t *testing.T) {
	t.Run("when value is nil use default value", func(t *testing.T) {
		// when
		res := GetBool(nil, true)

		// then
		assert.True(t, res)

		t.Run("different default value", func(t *testing.T) {
			// when
			res := GetBool(nil, false)

			// then
			assert.False(t, res)
		})
	})

	t.Run("use value when provided", func(t *testing.T) {
		// given
		v := false
		// when
		res := GetBool(&v, true)
		// then
		assert.False(t, res)
	})
}

func TestGetInt(t *testing.T) {
	t.Run("when value is nil use default value", func(t *testing.T) {
		// when
		res := GetInt(nil, 5)
		// then
		assert.Equal(t, res, 5)
	})
	t.Run("use value when provided", func(t *testing.T) {
		// given
		v := 10
		// when
		res := GetInt(&v, 5)
		// then
		assert.Equal(t, res, 10)
	})
}

func TestGetString(t *testing.T) {
	t.Run("when value is nil use default value", func(t *testing.T) {
		// when
		res := GetString(nil, "defValue")
		// then
		assert.Equal(t, res, "defValue")
	})
	t.Run("use value when provided", func(t *testing.T) {
		// given
		v := "providedValue"
		// when
		res := GetString(&v, "defValue")
		// then
		assert.Equal(t, res, "providedValue")

	})
}

func TestCopyOfMap(t *testing.T) {
	// given
	originalMap := map[string]map[string]string{
		"Canada": {
			"Toronto":   "Ontario",
			"Vancouver": "British Columbia",
			"Winniped":  "Manitoba",
		},
		"USA": {
			"New York":      "New York",
			"Chicago":       "Illinois",
			"San Francisco": "California",
			"Seattla":       "Washington",
		},
	}

	// when
	res := CopyOf(originalMap)

	// then
	assert.Equal(t, res, originalMap)
}
