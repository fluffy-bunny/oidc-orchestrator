package home

import (
	"net/http"

	di "github.com/fluffy-bunny/fluffy-dozm-di"
	contracts_handler "github.com/fluffy-bunny/fluffycore/echo/contracts/handler"
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

func init() {
	var _ contracts_handler.IHandler = (*service)(nil)
}

// AddScopedIHandler registers the *service as a singleton.
func AddScopedIHandler(builder di.ContainerBuilder) {
	contracts_handler.AddScopedIHandleWithMetadata[*service](builder,
		ctor,
		[]contracts_handler.HTTPVERB{
			contracts_handler.GET,
		},
		wellknown.HomePath,
	)

}
func ctor(config *contracts_config.Config) (*service, error) {
	return &service{
		config: config,
	}, nil
}
func (s *service) GetMiddleware() []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{}
}

func (s *service) Do(c echo.Context) error {
	// redirect temporarily to swagger
	return c.Redirect(http.StatusFound, "/swagger/")
}
