package sdkio

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Beijing Volcanoengine Technology Ltd.

import "io"

const (
	// Byte is 8 bits
	Byte int64 = 1
	// KbByte (KiB) is 1024 Bytes
	KbByte = Byte * 1024
	// MbByte (MiB) is 1024 KiB
	MbByte = KbByte * 1024
	// GbByte (GiB) is 1024 MiB
	GbByte = MbByte * 1024
)

const (
	SeekStart   = io.SeekStart   // seek relative to the origin of the file
	SeekCurrent = io.SeekCurrent // seek relative to the current offset
	SeekEnd     = io.SeekEnd     // seek relative to the end
)
