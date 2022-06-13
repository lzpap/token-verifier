package registryservice

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/capossele/swearfilter"
	"github.com/cockroachdb/errors"
	"github.com/labstack/echo"
	"github.com/lzpap/token-verifier/pkg/registry"
	"github.com/lzpap/token-verifier/pkg/registry/registryhttp"
	"go.uber.org/zap"
)

var (
	Networks = map[string]bool{
		"alphanet": true,
		"betanet":  false,
		"shimmer":  false,
	}
)

type HTTPHandler struct {
	service  registry.Service
	logger   *zap.SugaredLogger
	verifier *Verifier
	filter   *swearfilter.SwearFilter
}

func NewHTTPHandler(service registry.Service, logger *zap.SugaredLogger, verifier *Verifier) *HTTPHandler {
	return &HTTPHandler{service: service, logger: logger, filter: swearfilter.NewSwearFilter(true, badWords...), verifier: verifier}
}

func networkAllowed(network string) bool {
	return Networks[network]
}

// SaveToken saves a token to the registry
// Preform the following checks:
// 1. Perform token name length check
// 2. Perform token symbol length check
// 3. Perform token name swear check
// 4. Perform token symbol swear check
// 5. Check if tokenId is legit
func (h *HTTPHandler) SaveToken(c echo.Context) error {
	ctx := c.Request().Context()
	network := c.Param("network")
	if !networkAllowed(network) {
		return c.JSON(http.StatusForbidden, "network not allowed")
	}
	var token *registry.IRC30Token
	if err := json.NewDecoder(c.Request().Body).Decode(&token); err != nil {
		err = errors.Wrap(err, "failed to parse request body as JSON into an token")
		h.logger.Infow("Invalid http request", "error", err)
		return c.JSON(http.StatusBadRequest, registryhttp.NewErrorResponse(err))
	}

	// length checks
	if len(token.Name) > 20 {
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("IRC30Token name too long"))
	}

	if len(token.Symbol) > 4 {
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("IRC30Token symbol too long"))
	}

	// filter
	match, _ := h.filter.Check(token.Name)
	if len(match) > 0 {
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("%s is forbidden, as contains %v", token.Name, match))
	}
	match, _ = h.filter.Check(token.ID)
	if len(match) > 0 {
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("%s is forbidden, as contains %v", token.ID, match))
	}
	match, _ = h.filter.Check(token.Symbol)
	if len(match) > 0 {
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("%s is forbidden, as contains %v", token.Symbol, match))
	}

	match, _ = h.filter.Check(token.Description)
	if len(match) > 0 {
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("%s is forbidden, as contains %v", token.Description, match))
	}

	// semantic checks
	// has a unique name in the registry
	if _, err := h.service.FindTokenByName(ctx, network, token.Name); err == nil {
		return c.JSON(http.StatusBadRequest, registryhttp.NewErrorResponse(errors.New("token name already taken")))
	}
	// has a unique symbol in the registry
	if _, err := h.service.FindTokenBySymbol(ctx, network, token.Symbol); err == nil {
		return c.JSON(http.StatusBadRequest, registryhttp.NewErrorResponse(errors.New("token symbol already taken")))
	}
	// has a unique tokenId in the registry
	if _, err := h.service.LoadToken(ctx, network, token.ID); err == nil {
		return c.JSON(http.StatusBadRequest, registryhttp.NewErrorResponse(errors.New("token ID already registered")))
	}
	// token actually exists in the tangle, maxSupply matches the one in the foundry
	if err := h.verifier.Verify(token); err != nil {
		return c.JSON(http.StatusBadRequest, registryhttp.NewErrorResponse(errors.Wrap(err, "token verification failed")))
	}

	if err := h.service.SaveToken(ctx, network, token); err != nil {
		return c.JSON(http.StatusBadRequest, registryhttp.NewErrorResponse(errors.Wrap(err, "service failed to save Token")))
	}

	return c.JSON(http.StatusCreated, token)
}

func (h *HTTPHandler) LoadToken(c echo.Context) error {
	ctx := c.Request().Context()
	network := c.Param("network")
	if !networkAllowed(network) {
		return c.JSON(http.StatusForbidden, "network not allowed")
	}
	ID := c.Param("ID")
	result, err := h.service.LoadToken(ctx, network, ID)
	if err != nil {
		return c.JSON(http.StatusNotFound, errors.Wrap(err, "service failed to load IRC30Token"))
	}
	return c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) LoadTokens(c echo.Context) error {
	ctx := c.Request().Context()
	network := c.Param("network")
	if !networkAllowed(network) {
		return c.JSON(http.StatusForbidden, "network not allowed")
	}
	result, err := h.service.LoadTokens(ctx, network)
	if err != nil {
		return c.JSON(http.StatusNotFound, errors.Wrap(err, "service failed to load Assets"))
	}
	return c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) DeleteTokensByID(c echo.Context) error {
	ctx := c.Request().Context()
	network := c.Param("network")
	if !networkAllowed(network) {
		return c.JSON(http.StatusForbidden, "network not allowed")
	}
	ID := c.Param("ID")
	err := h.service.DeleteTokenByID(ctx, network, ID)
	if err != nil {
		return c.JSON(http.StatusNotFound, errors.Wrap(err, "service failed to delete the IRC30Token"))
	}
	return c.JSON(http.StatusOK, nil)
}

func (h *HTTPHandler) DeleteTokensByName(c echo.Context) error {
	ctx := c.Request().Context()
	network := c.Param("network")
	if !networkAllowed(network) {
		return c.JSON(http.StatusForbidden, "network not allowed")
	}
	name := c.Param("name")
	err := h.service.DeleteTokenByName(ctx, network, name)
	if err != nil {
		return c.JSON(http.StatusNotFound, errors.Wrap(err, "service failed to delete the IRC30Token"))
	}
	return c.JSON(http.StatusOK, nil)
}

func (h *HTTPHandler) LoadFilter(c echo.Context) error {
	result := h.filter.Load()
	return c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) AddFilter(c echo.Context) error {
	word := c.Param("word")
	if len(word) == 0 {
		return c.JSON(http.StatusBadRequest, "invalid empty-string as filter")
	}
	h.filter.Add(word)
	return c.JSON(http.StatusOK, word)
}

func (h *HTTPHandler) DeleteFilter(c echo.Context) error {
	word := c.Param("word")
	if len(word) == 0 {
		return c.JSON(http.StatusBadRequest, "invalid empty-string as filter")
	}
	h.filter.Delete(word)
	return c.JSON(http.StatusOK, word)
}
