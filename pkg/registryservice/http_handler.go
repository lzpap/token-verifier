package registryservice

import (
	"encoding/json"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/labstack/echo"
	"go.uber.org/zap"

	"github.com/capossele/asset-registry/pkg/registry"
	"github.com/capossele/asset-registry/pkg/registry/registryhttp"
)

var (
	networks = map[string]bool{
		"pollen":   true,
		"nectar":   true,
		"internal": true,
		"test":     true,
	}
)

type HTTPHandler struct {
	service registry.Service
	logger  *zap.SugaredLogger
}

func NewHTTPHandler(service registry.Service, logger *zap.SugaredLogger) *HTTPHandler {
	return &HTTPHandler{service: service, logger: logger}
}

func networkAllowed(network string) bool {
	return networks[network]
}

func (h *HTTPHandler) SaveAsset(c echo.Context) error {
	ctx := c.Request().Context()
	network := c.Param("network")
	if !networkAllowed(network) {
		return c.JSON(http.StatusForbidden, "network not allowed")
	}
	var asset *registry.Asset
	if err := json.NewDecoder(c.Request().Body).Decode(&asset); err != nil {
		err = errors.Wrap(err, "failed to parse request body as JSON into an asset")
		h.logger.Infow("Invalid http request", "error", err)
		return c.JSON(http.StatusBadRequest, registryhttp.NewErrorResponse(err))
	}
	if err := h.service.SaveAsset(ctx, network, asset); err != nil {
		return errors.Wrap(err, "service failed to save Assets")
	}
	return c.JSON(http.StatusCreated, asset)
}

func (h *HTTPHandler) LoadAsset(c echo.Context) error {
	ctx := c.Request().Context()
	network := c.Param("network")
	if !networkAllowed(network) {
		return c.JSON(http.StatusForbidden, "network not allowed")
	}
	ID := c.Param("ID")
	result, err := h.service.LoadAsset(ctx, network, ID)
	if err != nil {
		return c.JSON(http.StatusNotFound, errors.Wrap(err, "service failed to load Asset"))
	}
	return c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) LoadAssets(c echo.Context) error {
	ctx := c.Request().Context()
	network := c.Param("network")
	if !networkAllowed(network) {
		return c.JSON(http.StatusForbidden, "network not allowed")
	}
	result, err := h.service.LoadAssets(ctx, network)
	if err != nil {
		return c.JSON(http.StatusNotFound, errors.Wrap(err, "service failed to load Assets"))
	}
	return c.JSON(http.StatusOK, result)
}
