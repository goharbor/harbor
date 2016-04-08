package utils

var EPSILON float32 = 0.00000001

func FloatEquals(a, b float32) bool {
	if (a-b) < EPSILON && (b-a) < EPSILON {
		return true
	}
	return false
}
