package registry

import (
	"context"
)

// Asset defines the asset to be stored into a mongoDB.
type Asset struct {
	// ID defines the ID of the asset (base58 encoded).
	ID string `json:"ID" bson:"ID"`
	// Name defines name of the asset.
	Name string `json:"name" bson:"name"`
	// Symbol defines the symbol of the asset.
	Symbol string `json:"symbol" bson:"symbol"`
	// Supply defines the original total supply of the asset.
	Supply uint64 `json:"supply" bson:"supply"`
	// TransactionID defines the transaction ID (base58 encoded) that created this asset.
	TransactionID string `json:"transactionID" bson:"transactionID"`
}

type Service interface {
	SaveAsset(ctx context.Context, network string, record *Asset) error
	LoadAssets(ctx context.Context, network string, ID ...string) ([]*Asset, error)
	LoadAsset(ctx context.Context, network string, ID string) (*Asset, error)
}
