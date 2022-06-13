package registry

import (
	"context"
)

// IRC30Token defines and IRC30 native token and its metadata to be stored into a mongoDB.
type IRC30Token struct {
	// ID defines tokenId.
	ID string `json:"ID" bson:"ID"`
	// Name defines name of the token.
	Name string `json:"name" bson:"name"`
	// Description defines description of the token.
	Description string `json:"description" bson:"description"`
	// Symbol defines the symbol of the token.
	Symbol string `json:"symbol" bson:"symbol"`
	// Decimals defines the number of decimals of the token.
	Decimals uint `json:"decimals" bson:"decimals"`
	// URL defines the URL of the token.
	URL string `json:"url" bson:"url"`
	// LogoURL defines the url of the token logo.
	LogoURL string `json:"logoUrl" bson:"logoUrl"`
	// Logo defines the svg logo of the token encoded as a hex byte string.
	Logo string `json:"logo" bson:"logo"`
	// MaxSupply defines the possible maximum supply of the token.
	MaxSupply string `json:"maxSupply" bson:"maxSupply"`
}

type Service interface {
	FindTokenBySymbol(ctx context.Context, network string, symbol string) (*IRC30Token, error)
	FindTokenByName(ctx context.Context, network string, name string) (*IRC30Token, error)
	SaveToken(ctx context.Context, network string, record *IRC30Token) error
	LoadTokens(ctx context.Context, network string, ID ...string) ([]*IRC30Token, error)
	LoadToken(ctx context.Context, network string, ID string) (*IRC30Token, error)
	DeleteTokenByID(ctx context.Context, network string, ID string) error
	DeleteTokenByName(ctx context.Context, network string, name string) error
}
