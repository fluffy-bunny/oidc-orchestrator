package healthz

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
		discoveryDocument *contracts_downstream.DiscoveryDocument
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
		wellknown.DiscoveryPath,
	)

}
func ctor(config *contracts_config.Config, downstreamService contracts_downstream.IDownstreamOIDCService) (*service, error) {
	discoveryDocument, err := downstreamService.GetDiscoveryDocument()
	if err != nil {
		return nil, err
	}
	return &service{
		config:            config,
		discoveryDocument: discoveryDocument,
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
// @Success 200 {object} contracts_downstream.DiscoveryDocument
// @Router /.well-known/openid-configuration [get]
func (s *service) Do(c echo.Context) error {
	baseUrl := "http://" + c.Request().Host

	s.discoveryDocument.JwksURI = baseUrl + wellknown.JWKSPath
	//s.discoveryDocument.AuthorizationEndpoint = baseUrl + wellknown.AuthorizationPath
	s.discoveryDocument.TokenEndpoint = baseUrl + wellknown.TokenPath
	s.discoveryDocument.Issuer = baseUrl
	s.discoveryDocument.UserinfoEndpoint = baseUrl + wellknown.UserInfoPath
	s.discoveryDocument.IDTokenSigningAlgValuesSupported = []string{"ES256"}
	s.discoveryDocument.GrantTypesSupported = []string{"authorization_code", "refresh_token"}
	s.discoveryDocument.ResponseTypesSupported = []string{"code"}
	return c.JSON(http.StatusOK, s.discoveryDocument)
}
