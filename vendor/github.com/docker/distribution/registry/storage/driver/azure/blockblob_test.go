package azure

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	azure "github.com/Azure/azure-sdk-for-go/storage"
)

type StorageSimulator struct {
	blobs map[string]*BlockBlob
}

type BlockBlob struct {
	blocks    map[string]*DataBlock
	blockList []string
}

type DataBlock struct {
	data      []byte
	committed bool
}

func (s *StorageSimulator) path(container, blob string) string {
	return fmt.Sprintf("%s/%s", container, blob)
}

func (s *StorageSimulator) BlobExists(container, blob string) (bool, error) {
	_, ok := s.blobs[s.path(container, blob)]
	return ok, nil
}

func (s *StorageSimulator) GetBlob(container, blob string) (io.ReadCloser, error) {
	bb, ok := s.blobs[s.path(container, blob)]
	if !ok {
		return nil, fmt.Errorf("blob not found")
	}

	var readers []io.Reader
	for _, bID := range bb.blockList {
		readers = append(readers, bytes.NewReader(bb.blocks[bID].data))
	}
	return ioutil.NopCloser(io.MultiReader(readers...)), nil
}

func (s *StorageSimulator) GetSectionReader(container, blob string, start, length int64) (io.ReadCloser, error) {
	r, err := s.GetBlob(container, blob)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(b[start : start+length])), nil
}

func (s *StorageSimulator) CreateBlockBlob(container, blob string) error {
	path := s.path(container, blob)
	bb := &BlockBlob{
		blocks:    make(map[string]*DataBlock),
		blockList: []string{},
	}
	s.blobs[path] = bb
	return nil
}

func (s *StorageSimulator) PutBlock(container, blob, blockID string, chunk []byte) error {
	path := s.path(container, blob)
	bb, ok := s.blobs[path]
	if !ok {
		return fmt.Errorf("blob not found")
	}
	data := make([]byte, len(chunk))
	copy(data, chunk)
	bb.blocks[blockID] = &DataBlock{data: data, committed: false} // add block to blob
	return nil
}

func (s *StorageSimulator) GetBlockList(container, blob string, blockType azure.BlockListType) (azure.BlockListResponse, error) {
	resp := azure.BlockListResponse{}
	bb, ok := s.blobs[s.path(container, blob)]
	if !ok {
		return resp, fmt.Errorf("blob not found")
	}

	// Iterate committed blocks (in order)
	if blockType == azure.BlockListTypeAll || blockType == azure.BlockListTypeCommitted {
		for _, blockID := range bb.blockList {
			b := bb.blocks[blockID]
			block := azure.BlockResponse{
				Name: blockID,
				Size: int64(len(b.data)),
			}
			resp.CommittedBlocks = append(resp.CommittedBlocks, block)
		}

	}

	// Iterate uncommitted blocks (in no order)
	if blockType == azure.BlockListTypeAll || blockType == azure.BlockListTypeCommitted {
		for blockID, b := range bb.blocks {
			block := azure.BlockResponse{
				Name: blockID,
				Size: int64(len(b.data)),
			}
			if !b.committed {
				resp.UncommittedBlocks = append(resp.UncommittedBlocks, block)
			}
		}
	}
	return resp, nil
}

func (s *StorageSimulator) PutBlockList(container, blob string, blocks []azure.Block) error {
	bb, ok := s.blobs[s.path(container, blob)]
	if !ok {
		return fmt.Errorf("blob not found")
	}

	var blockIDs []string
	for _, v := range blocks {
		bl, ok := bb.blocks[v.ID]
		if !ok { // check if block ID exists
			return fmt.Errorf("Block id '%s' not found", v.ID)
		}
		bl.committed = true
		blockIDs = append(blockIDs, v.ID)
	}

	// Mark all other blocks uncommitted
	for k, b := range bb.blocks {
		inList := false
		for _, v := range blockIDs {
			if k == v {
				inList = true
				break
			}
		}
		if !inList {
			b.committed = false
		}
	}

	bb.blockList = blockIDs
	return nil
}

func NewStorageSimulator() StorageSimulator {
	return StorageSimulator{
		blobs: make(map[string]*BlockBlob),
	}
}
