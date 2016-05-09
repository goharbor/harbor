package azure

import (
	"fmt"
	"io"

	azure "github.com/Azure/azure-sdk-for-go/storage"
)

// azureBlockStorage is adaptor between azure.BlobStorageClient and
// blockStorage interface.
type azureBlockStorage struct {
	azure.BlobStorageClient
}

func (b *azureBlockStorage) GetSectionReader(container, blob string, start, length int64) (io.ReadCloser, error) {
	return b.BlobStorageClient.GetBlobRange(container, blob, fmt.Sprintf("%v-%v", start, start+length-1))
}

func newAzureBlockStorage(b azure.BlobStorageClient) azureBlockStorage {
	a := azureBlockStorage{}
	a.BlobStorageClient = b
	return a
}
