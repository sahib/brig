// Utility functions that would not hurt the simplicity of Go
// if they would be in the builtins/stdlib.
package util

// Returns the minimum of a and b.
func Min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

// Returns the maximum of a and b.
func Max(a, b int) int {
	if a < b {
		return b
	} else {
		return a
	}
}

// Clamps x into [lo, hi]
func Clamp(x, lo, hi int) int {
	return Max(lo, Min(x, hi))
}

// Like Min() but for uint
func UMin(a, b uint) uint {
	if a < b {
		return a
	} else {
		return b
	}
}

// Like Max() but for uint
func UMax(a, b uint) uint {
	if a < b {
		return b
	} else {
		return a
	}
}

// Like Clamp() but for uint
func UClamp(x, lo, hi uint) uint {
	return UMax(lo, UMin(x, hi))
}
