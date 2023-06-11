package targets

import (
	"fmt"
	"strconv"
	"strings"
)

const MinDelegationHashPrefixBitLen = 1
const MaxDelegationHashPrefixBitLen = 32

// hexEncode formats x as a hex string. The hex string is left padded with
// zeros to padWidth, if necessary.
func hexEncode(x uint64, padWidth int) string {
	// Benchmarked to be more than 10x faster than padding with Sprintf.
	s := strconv.FormatUint(x, 16)
	if len(s) >= padWidth {
		return s
	}
	return strings.Repeat("0", padWidth-len(s)) + s
}

const bitsPerHexDigit = 4

// numHexDigits returns the number of hex digits required to encode the given
// number of bits.
func numHexDigits(numBits int) int {
	// ceil(numBits / bitsPerHexDigit)
	return ((numBits - 1) / bitsPerHexDigit) + 1
}

// HashBins represents an ordered list of hash bin target roles, which together
// partition the space of target path hashes equal-sized buckets based on path
// has prefix.
type HashBins struct {
	rolePrefix  string
	bitLen      int
	hexDigitLen int

	numBins           uint64
	numPrefixesPerBin uint64
}

// NewHashBins creates a HashBins partitioning with 2^bitLen buckets.
func NewHashBins(rolePrefix string, bitLen int) (*HashBins, error) {
	if bitLen < MinDelegationHashPrefixBitLen || bitLen > MaxDelegationHashPrefixBitLen {
		return nil, fmt.Errorf("bitLen is out of bounds, should be between %v and %v inclusive", MinDelegationHashPrefixBitLen, MaxDelegationHashPrefixBitLen)
	}

	hexDigitLen := numHexDigits(bitLen)
	numBins := uint64(1) << bitLen

	numPrefixesTotal := uint64(1) << (bitsPerHexDigit * hexDigitLen)
	numPrefixesPerBin := numPrefixesTotal / numBins

	return &HashBins{
		rolePrefix:        rolePrefix,
		bitLen:            bitLen,
		hexDigitLen:       hexDigitLen,
		numBins:           numBins,
		numPrefixesPerBin: numPrefixesPerBin,
	}, nil
}

// NumBins returns the number of hash bin partitions.
func (hb *HashBins) NumBins() uint64 {
	return hb.numBins
}

// GetBin returns the HashBin at index i, or nil if i is out of bounds.
func (hb *HashBins) GetBin(i uint64) *HashBin {
	if i >= hb.numBins {
		return nil
	}

	return &HashBin{
		rolePrefix:  hb.rolePrefix,
		hexDigitLen: hb.hexDigitLen,
		first:       i * hb.numPrefixesPerBin,
		last:        ((i + 1) * hb.numPrefixesPerBin) - 1,
	}
}

// HashBin represents a hex prefix range. First should be less than Last.
type HashBin struct {
	rolePrefix  string
	hexDigitLen int
	first       uint64
	last        uint64
}

// RoleName returns the name of the role that signs for the HashBin.
func (b *HashBin) RoleName() string {
	if b.first == b.last {
		return b.rolePrefix + hexEncode(b.first, b.hexDigitLen)
	}

	return b.rolePrefix + hexEncode(b.first, b.hexDigitLen) + "-" + hexEncode(b.last, b.hexDigitLen)
}

// HashPrefixes returns a slice of all hash prefixes in the bin.
func (b *HashBin) HashPrefixes() []string {
	n := int(b.last - b.first + 1)
	ret := make([]string, int(n))

	x := b.first
	for i := 0; i < n; i++ {
		ret[i] = hexEncode(x, b.hexDigitLen)
		x++
	}

	return ret
}
