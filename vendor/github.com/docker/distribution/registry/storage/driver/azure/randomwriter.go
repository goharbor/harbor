package azure

import (
	"fmt"
	"io"
	"io/ioutil"

	azure "github.com/Azure/azure-sdk-for-go/storage"
)

// blockStorage is the interface required from a block storage service
// client implementation
type blockStorage interface {
	CreateBlockBlob(container, blob string) error
	GetBlob(container, blob string) (io.ReadCloser, error)
	GetSectionReader(container, blob string, start, length int64) (io.ReadCloser, error)
	PutBlock(container, blob, blockID string, chunk []byte) error
	GetBlockList(container, blob string, blockType azure.BlockListType) (azure.BlockListResponse, error)
	PutBlockList(container, blob string, blocks []azure.Block) error
}

// randomBlobWriter enables random access semantics on Azure block blobs
// by enabling writing arbitrary length of chunks to arbitrary write offsets
// within the blob. Normally, Azure Blob Storage does not support random
// access semantics on block blobs; however, this writer can download, split and
// reupload the overlapping blocks and discards those being overwritten entirely.
type randomBlobWriter struct {
	bs        blockStorage
	blockSize int
}

func newRandomBlobWriter(bs blockStorage, blockSize int) randomBlobWriter {
	return randomBlobWriter{bs: bs, blockSize: blockSize}
}

// WriteBlobAt writes the given chunk to the specified position of an existing blob.
// The offset must be equals to size of the blob or smaller than it.
func (r *randomBlobWriter) WriteBlobAt(container, blob string, offset int64, chunk io.Reader) (int64, error) {
	rand := newBlockIDGenerator()

	blocks, err := r.bs.GetBlockList(container, blob, azure.BlockListTypeCommitted)
	if err != nil {
		return 0, err
	}
	rand.Feed(blocks) // load existing block IDs

	// Check for write offset for existing blob
	size := getBlobSize(blocks)
	if offset < 0 || offset > size {
		return 0, fmt.Errorf("wrong offset for Write: %v", offset)
	}

	// Upload the new chunk as blocks
	blockList, nn, err := r.writeChunkToBlocks(container, blob, chunk, rand)
	if err != nil {
		return 0, err
	}

	// For non-append operations, existing blocks may need to be splitted
	if offset != size {
		// Split the block on the left end (if any)
		leftBlocks, err := r.blocksLeftSide(container, blob, offset, rand)
		if err != nil {
			return 0, err
		}
		blockList = append(leftBlocks, blockList...)

		// Split the block on the right end (if any)
		rightBlocks, err := r.blocksRightSide(container, blob, offset, nn, rand)
		if err != nil {
			return 0, err
		}
		blockList = append(blockList, rightBlocks...)
	} else {
		// Use existing block list
		var existingBlocks []azure.Block
		for _, v := range blocks.CommittedBlocks {
			existingBlocks = append(existingBlocks, azure.Block{ID: v.Name, Status: azure.BlockStatusCommitted})
		}
		blockList = append(existingBlocks, blockList...)
	}
	// Put block list
	return nn, r.bs.PutBlockList(container, blob, blockList)
}

func (r *randomBlobWriter) GetSize(container, blob string) (int64, error) {
	blocks, err := r.bs.GetBlockList(container, blob, azure.BlockListTypeCommitted)
	if err != nil {
		return 0, err
	}
	return getBlobSize(blocks), nil
}

// writeChunkToBlocks writes given chunk to one or multiple blocks within specified
// blob and returns their block representations. Those blocks are not committed, yet
func (r *randomBlobWriter) writeChunkToBlocks(container, blob string, chunk io.Reader, rand *blockIDGenerator) ([]azure.Block, int64, error) {
	var newBlocks []azure.Block
	var nn int64

	// Read chunks of at most size N except the last chunk to
	// maximize block size and minimize block count.
	buf := make([]byte, r.blockSize)
	for {
		n, err := io.ReadFull(chunk, buf)
		if err == io.EOF {
			break
		}
		nn += int64(n)
		data := buf[:n]
		blockID := rand.Generate()
		if err := r.bs.PutBlock(container, blob, blockID, data); err != nil {
			return newBlocks, nn, err
		}
		newBlocks = append(newBlocks, azure.Block{ID: blockID, Status: azure.BlockStatusUncommitted})
	}
	return newBlocks, nn, nil
}

// blocksLeftSide returns the blocks that are going to be at the left side of
// the writeOffset: [0, writeOffset) by identifying blocks that will remain
// the same and splitting blocks and reuploading them as needed.
func (r *randomBlobWriter) blocksLeftSide(container, blob string, writeOffset int64, rand *blockIDGenerator) ([]azure.Block, error) {
	var left []azure.Block
	bx, err := r.bs.GetBlockList(container, blob, azure.BlockListTypeAll)
	if err != nil {
		return left, err
	}

	o := writeOffset
	elapsed := int64(0)
	for _, v := range bx.CommittedBlocks {
		blkSize := int64(v.Size)
		if o >= blkSize { // use existing block
			left = append(left, azure.Block{ID: v.Name, Status: azure.BlockStatusCommitted})
			o -= blkSize
			elapsed += blkSize
		} else if o > 0 { // current block needs to be splitted
			start := elapsed
			size := o
			part, err := r.bs.GetSectionReader(container, blob, start, size)
			if err != nil {
				return left, err
			}
			newBlockID := rand.Generate()

			data, err := ioutil.ReadAll(part)
			if err != nil {
				return left, err
			}
			if err = r.bs.PutBlock(container, blob, newBlockID, data); err != nil {
				return left, err
			}
			left = append(left, azure.Block{ID: newBlockID, Status: azure.BlockStatusUncommitted})
			break
		}
	}
	return left, nil
}

// blocksRightSide returns the blocks that are going to be at the right side of
// the written chunk: [writeOffset+size, +inf) by identifying blocks that will remain
// the same and splitting blocks and reuploading them as needed.
func (r *randomBlobWriter) blocksRightSide(container, blob string, writeOffset int64, chunkSize int64, rand *blockIDGenerator) ([]azure.Block, error) {
	var right []azure.Block

	bx, err := r.bs.GetBlockList(container, blob, azure.BlockListTypeAll)
	if err != nil {
		return nil, err
	}

	re := writeOffset + chunkSize - 1 // right end of written chunk
	var elapsed int64
	for _, v := range bx.CommittedBlocks {
		var (
			bs = elapsed                     // left end of current block
			be = elapsed + int64(v.Size) - 1 // right end of current block
		)

		if bs > re { // take the block as is
			right = append(right, azure.Block{ID: v.Name, Status: azure.BlockStatusCommitted})
		} else if be > re { // current block needs to be splitted
			part, err := r.bs.GetSectionReader(container, blob, re+1, be-(re+1)+1)
			if err != nil {
				return right, err
			}
			newBlockID := rand.Generate()

			data, err := ioutil.ReadAll(part)
			if err != nil {
				return right, err
			}
			if err = r.bs.PutBlock(container, blob, newBlockID, data); err != nil {
				return right, err
			}
			right = append(right, azure.Block{ID: newBlockID, Status: azure.BlockStatusUncommitted})
		}
		elapsed += int64(v.Size)
	}
	return right, nil
}

func getBlobSize(blocks azure.BlockListResponse) int64 {
	var n int64
	for _, v := range blocks.CommittedBlocks {
		n += int64(v.Size)
	}
	return n
}
