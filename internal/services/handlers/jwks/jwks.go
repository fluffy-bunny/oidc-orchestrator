package healthz

import (
	"net/http"

	di "github.com/fluffy-bunny/fluffy-dozm-di"
	contracts_handler "github.com/fluffy-bunny/fluffycore/echo/contracts/handler"
	mocks_oauth2 "github.com/fluffy-bunny/fluffycore/mocks/oauth2"
	contracts_config "github.com/fluffy-bunny/oidc-orchestrator/internal/contracts/config"
	contracts_downstream "github.com/fluffy-bunny/oidc-orchestrator/internal/contracts/downstream"
	wellknown "github.com/fluffy-bunny/oidc-orchestrator/internal/wellknown"
	echo "github.com/labstack/echo/v4"
)

type (
	service struct {
		config            *contracts_config.Config
		downstreamService contracts_downstream.IDownstreamOIDCService
	}
)

var signingKey *mocks_oauth2.SigningKey
var jwksKeys *mocks_oauth2.JWKSKeys

func init() {
	var _ contracts_handler.IHandler = (*service)(nil)
	signingKey, _ = mocks_oauth2.LoadSigningKey()
	jwksKeys = &mocks_oauth2.JWKSKeys{
		Keys: []mocks_oauth2.PublicJwk{
			signingKey.PublicJwk,
		},
	}
}

// AddScopedIHandler registers the *service as a singleton.
func AddScopedIHandler(builder di.ContainerBuilder) {
	contracts_handler.AddScopedIHandleWithMetadata[*service](builder,
		ctor,
		[]contracts_handler.HTTPVERB{
			contracts_handler.GET,
		},
		wellknown.JWKSPath,
	)

}
func ctor(config *contracts_config.Config, downstreamService contracts_downstream.IDownstreamOIDCService) (*service, error) {
	return &service{
		config:            config,
		downstreamService: downstreamService,
	}, nil
}
func (s *service) GetMiddleware() []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{}
}

// HealthCheck godoc
// @Summary Show the status of server.
// @Description get the status of server.
// @Tags root
// @Accept */*
// @Produce json
// @Success 200 {object} interface{}
// @Router /.well-known/jwks [get]
func (s *service) Do(c echo.Context) error {
	return c.JSON(http.StatusOK, jwksKeys)
}
