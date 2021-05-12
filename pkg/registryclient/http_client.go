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

func (c *HTTPClient) SaveAssets(ctx context.Context, network string, assets []*registry.Asset) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(assets).
		SetError(&registryhttp.ErrorResponse{}).
		Post(registryhttp.RegistriesEndpoint + "/" + network + registryhttp.AssetsEndpoint)
	if err != nil {
		return errors.Wrap(err, "failed to execute saveAssets HTTP call")
	}
	if resp.IsSuccess() {
		return nil
	}
	errorResp := resp.Error().(*registryhttp.ErrorResponse)
	return errors.Newf("saveAssets HTTP call returns an error: %s", errorResp.Error)
}

func (c *HTTPClient) LoadAssets(ctx context.Context, network string, assets ...string) error {
	assetID := ""
	if len(assets) > 0 {
		assetID = "/" + assets[0]
	}
	resp, err := c.client.R().
		SetContext(ctx).
		SetError(&registryhttp.ErrorResponse{}).
		Get(registryhttp.RegistriesEndpoint + "/" + network + registryhttp.AssetsEndpoint + assetID)
	if err != nil {
		return errors.Wrap(err, "failed to execute loadAssets HTTP call")
	}
	if resp.IsSuccess() {
		return nil
	}
	errorResp := resp.Error().(*registryhttp.ErrorResponse)
	return errors.Newf("loadAssets HTTP call returns an error: %s", errorResp.Error)
}
