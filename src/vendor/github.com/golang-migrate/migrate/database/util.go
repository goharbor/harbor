package database

import (
	"fmt"
	"hash/crc32"
)

const advisoryLockIdSalt uint = 1486364155

// inspired by rails migrations, see https://goo.gl/8o9bCT
func GenerateAdvisoryLockId(databaseName string) (string, error) {
	sum := crc32.ChecksumIEEE([]byte(databaseName))
	sum = sum * uint32(advisoryLockIdSalt)
	return fmt.Sprintf("%v", sum), nil
}
