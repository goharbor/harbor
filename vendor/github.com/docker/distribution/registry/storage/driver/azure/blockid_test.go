package azure

import (
	"math"
	"testing"

	azure "github.com/Azure/azure-sdk-for-go/storage"
)

func Test_blockIdGenerator(t *testing.T) {
	r := newBlockIDGenerator()

	for i := 1; i <= 10; i++ {
		if expected := i - 1; len(r.pool) != expected {
			t.Fatalf("rand pool had wrong number of items: %d, expected:%d", len(r.pool), expected)
		}
		if id := r.Generate(); id == "" {
			t.Fatal("returned empty id")
		}
		if expected := i; len(r.pool) != expected {
			t.Fatalf("rand pool has wrong number of items: %d, expected:%d", len(r.pool), expected)
		}
	}
}

func Test_blockIdGenerator_Feed(t *testing.T) {
	r := newBlockIDGenerator()
	if expected := 0; len(r.pool) != expected {
		t.Fatalf("rand pool had wrong number of items: %d, expected:%d", len(r.pool), expected)
	}

	// feed empty list
	blocks := azure.BlockListResponse{}
	r.Feed(blocks)
	if expected := 0; len(r.pool) != expected {
		t.Fatalf("rand pool had wrong number of items: %d, expected:%d", len(r.pool), expected)
	}

	// feed blocks
	blocks = azure.BlockListResponse{
		CommittedBlocks: []azure.BlockResponse{
			{"1", 1},
			{"2", 2},
		},
		UncommittedBlocks: []azure.BlockResponse{
			{"3", 3},
		}}
	r.Feed(blocks)
	if expected := 3; len(r.pool) != expected {
		t.Fatalf("rand pool had wrong number of items: %d, expected:%d", len(r.pool), expected)
	}

	// feed same block IDs with committed/uncommitted place changed
	blocks = azure.BlockListResponse{
		CommittedBlocks: []azure.BlockResponse{
			{"3", 3},
		},
		UncommittedBlocks: []azure.BlockResponse{
			{"1", 1},
		}}
	r.Feed(blocks)
	if expected := 3; len(r.pool) != expected {
		t.Fatalf("rand pool had wrong number of items: %d, expected:%d", len(r.pool), expected)
	}
}

func Test_toBlockId(t *testing.T) {
	min := 0
	max := math.MaxInt64

	if len(toBlockID(min)) != len(toBlockID(max)) {
		t.Fatalf("different-sized blockIDs are returned")
	}
}
