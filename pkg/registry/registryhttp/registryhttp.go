package registryhttp

const (
	RegistriesEndpoint = "/registries"
	TokensEndpoint     = "/tokens"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewErrorResponse(err error) *ErrorResponse {
	return &ErrorResponse{Error: err.Error()}
}
