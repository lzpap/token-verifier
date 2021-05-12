package registryclient

import (
	"context"

	"github.com/capossele/asset-registry/pkg/registry"
	"github.com/capossele/asset-registry/pkg/registry/registryhttp"
	"github.com/cockroachdb/errors"
	"github.com/go-resty/resty/v2"
)

type HTTPClient struct {
	client *resty.Client
}

func NewHTTPClient(restyClient *resty.Client) *HTTPClient {
	return &HTTPClient{client: restyClient}

}

func (c *HTTPClient) SaveAssets(ctx context.Context, assets []*registry.Asset) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(assets).
		SetError(&registryhttp.ErrorResponse{}).
		Post(registryhttp.SaveAssetsEndpoint)
	if err != nil {
		return errors.Wrap(err, "failed to execute saveAssets HTTP call")
	}
	if resp.IsSuccess() {
		return nil
	}
	errorResp := resp.Error().(*registryhttp.ErrorResponse)
	return errors.Newf("saveAssets HTTP call returns an error: %s", errorResp.Error)
}
