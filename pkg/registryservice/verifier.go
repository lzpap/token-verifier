package registryservice

import (
	"context"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/lzpap/token-verifier/pkg/registry"
	"github.com/pkg/errors"
	"net/url"
	"time"
)

type Verifier struct {
	client *nodeclient.Client
}

// NewVerifier creates a new token verifier.
func NewVerifier(baseUrl string) *Verifier {
	return &Verifier{
		client: nodeclient.New(baseUrl),
	}
}

func (v *Verifier) Verify(token *registry.IRC30Token) error {
	// Can it be parsed to bytes
	tokenIdBytes, err := iotago.DecodeHex(token.ID)
	if err != nil {
		return errors.Wrap(err, "failed to parse tokenId")
	}
	// is it the correct length?
	if len(tokenIdBytes) != iotago.FoundryIDLength {
		return errors.New("tokenId is not valid, wrong length")
	}
	// the first byte is always an alias address type byte
	if tokenIdBytes[0] != byte(iotago.AddressAlias) {
		return errors.New("tokenId does not start with an alias address type byte")
	}
	// tokenScheme is simple, meaning the last byt us 0
	if tokenIdBytes[iotago.FoundryIDLength-1] != 0 {
		return errors.New("tokenId does not end with a 0 byte")
	}

	var foundryId iotago.FoundryID
	copy(foundryId[:], tokenIdBytes)

	if token.Decimals == 0 {
		return errors.New("tokenDecimals is 0")
	}

	// validate token url if present
	if len(token.URL) > 0 {
		_, err = url.ParseRequestURI(token.URL)
		if err != nil {
			return errors.Wrap(err, "failed to validate tokenURL")
		}
	}

	// validate logo url if present
	if len(token.LogoURL) > 0 {
		_, err = url.ParseRequestURI(token.LogoURL)
		if err != nil {
			return errors.Wrap(err, "failed to validate logoURL")
		}
	}

	// fetch the foundry output
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	indexerClient, err := v.client.Indexer(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get indexer client")
	}

	_, fOutput, err := indexerClient.Foundry(ctx, foundryId)
	if err != nil {
		return errors.Wrap(err, "failed to get foundry output")
	}

	supplyInfo, ok := fOutput.TokenScheme.(*iotago.SimpleTokenScheme)
	if !ok {
		return errors.New("foundry output is not a simple token scheme")
	}
	if supplyInfo.MaximumSupply.String() != token.MaxSupply {
		return errors.New("mismatch in maximum supply")
	}

	return nil
}
