package sdkrand

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Beijing Volcanoengine Technology Ltd.

import "math/rand"

func Read(r *rand.Rand, p []byte) (int, error) {
	return r.Read(p)
}
