package config

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnumValidator(t *testing.T) {
	defaults := DefaultMapping{
		"enum-val": DefaultEntry{
			Default:      "a",
			NeedsRestart: false,
			Validator:    EnumValidator("a", "b", "c"),
		},
	}

	// Check initial validation:
	_, err := Open(bytes.NewReader([]byte("enum-val: d")), defaults)
	require.NotNil(t, err)

	cfg, err := Open(bytes.NewReader([]byte("enum-val: c")), defaults)
	require.Nil(t, err)
	require.Equal(t, cfg.String("enum-val"), "c")

	// Set an invalid enum value:
	require.NotNil(t, cfg.SetString("enum-val", "C"))
	require.Nil(t, cfg.SetString("enum-val", "a"))
	require.Equal(t, cfg.String("enum-val"), "a")
}

func TestIntValidator(t *testing.T) {
	vdt := IntRangeValidator(10, 100)
	require.Contains(t, vdt("x").Error(), "is not an int64")
	require.Contains(t, vdt(int64(9)).Error(), "may not be less than 10")
	require.Contains(t, vdt(int64(101)).Error(), "may not be more than 100")

	require.Nil(t, vdt(int64(10)))
	require.Nil(t, vdt(int64(100)))
	require.Nil(t, vdt(int64(50)))
}

func TestFloatValidator(t *testing.T) {
	vdt := FloatRangeValidator(0.5, 1.5)
	require.Contains(t, vdt("x").Error(), "is not a float")
	require.Contains(t, vdt(int64(1)).Error(), "is not a float")
	require.Contains(t, vdt(float64(0.49999999999999)).Error(), "may not be less than 0.5")
	require.Contains(t, vdt(float64(1.50000000000001)).Error(), "may not be more than 1.5")

	require.Nil(t, vdt(float64(0.50)))
	require.Nil(t, vdt(float64(1.50)))
	require.Nil(t, vdt(float64(0.75)))
}
