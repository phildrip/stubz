package external

import (
	"context"
	"github.com/phildrip/toe/testdata/input/external/types"
	"io"
)

// ExternalInterface uses type aliases from another package
type ExternalInterface interface {
	Get(ctx context.Context, id types.BlobId) (string, error)
	Set(ctx context.Context, id types.BlobId, content io.Reader) error
	Delete(ctx context.Context, id types.BlobId, collectionId types.CollectionId) error
}

