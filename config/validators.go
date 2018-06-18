package config

import "fmt"

// EnumValidator checks if the supplied string value is in the `options` list.
func EnumValidator(options ...string) func(val interface{}) error {
	return func(val interface{}) error {
		s, ok := val.(string)
		if !ok {
			return fmt.Errorf("enum value is not a string: %v", val)
		}

		for _, option := range options {
			if option == s {
				return nil
			}
		}

		return fmt.Errorf("not a valid enum value: %v (allowed: %v)", s, options)
	}
}

// IntRangeValidator checks if the supplied integer value lies in the
// inclusive boundaries of `min` and `max`.
func IntRangeValidator(min, max int64) func(val interface{}) error {
	return func(val interface{}) error {
		i, ok := val.(int64)
		if !ok {
			return fmt.Errorf("value is not an int64: %v", val)
		}

		if i < min {
			return fmt.Errorf("value may not be less than %d", min)
		}

		if i > max {
			return fmt.Errorf("value may not be more than %d", max)
		}

		return nil
	}
}

// FloatRangeValidator checks if the supplied float value lies in the
// inclusive boundaries of `min` and `max`.
func FloatRangeValidator(min, max float64) func(val interface{}) error {
	return func(val interface{}) error {
		i, ok := val.(float64)
		if !ok {
			return fmt.Errorf("value is not a float64: %v", val)
		}

		if i < min {
			return fmt.Errorf("value may not be less than %f", min)
		}

		if i > max {
			return fmt.Errorf("value may not be more than %f", max)
		}

		return nil
	}
}
