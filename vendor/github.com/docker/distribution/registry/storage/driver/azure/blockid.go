package azure

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"sync"
	"time"

	azure "github.com/Azure/azure-sdk-for-go/storage"
)

type blockIDGenerator struct {
	pool map[string]bool
	r    *rand.Rand
	m    sync.Mutex
}

// Generate returns an unused random block id and adds the generated ID
// to list of used IDs so that the same block name is not used again.
func (b *blockIDGenerator) Generate() string {
	b.m.Lock()
	defer b.m.Unlock()

	var id string
	for {
		id = toBlockID(int(b.r.Int()))
		if !b.exists(id) {
			break
		}
	}
	b.pool[id] = true
	return id
}

func (b *blockIDGenerator) exists(id string) bool {
	_, used := b.pool[id]
	return used
}

func (b *blockIDGenerator) Feed(blocks azure.BlockListResponse) {
	b.m.Lock()
	defer b.m.Unlock()

	for _, bl := range append(blocks.CommittedBlocks, blocks.UncommittedBlocks...) {
		b.pool[bl.Name] = true
	}
}

func newBlockIDGenerator() *blockIDGenerator {
	return &blockIDGenerator{
		pool: make(map[string]bool),
		r:    rand.New(rand.NewSource(time.Now().UnixNano()))}
}

// toBlockId converts given integer to base64-encoded block ID of a fixed length.
func toBlockID(i int) string {
	s := fmt.Sprintf("%029d", i) // add zero padding for same length-blobs
	return base64.StdEncoding.EncodeToString([]byte(s))
}
