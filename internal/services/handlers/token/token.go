package token

import (
	"net/http"

	di "github.com/fluffy-bunny/fluffy-dozm-di"
	contracts_handler "github.com/fluffy-bunny/fluffycore/echo/contracts/handler"
	contracts_config "github.com/fluffy-bunny/oidc-orchestrator/internal/contracts/config"
	contracts_downstream "github.com/fluffy-bunny/oidc-orchestrator/internal/contracts/downstream"
	wellknown "github.com/fluffy-bunny/oidc-orchestrator/internal/wellknown"
	echo "github.com/labstack/echo/v4"
	zerolog "github.com/rs/zerolog"
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
			contracts_handler.POST,
		},
		wellknown.TokenPath,
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
// @Router /token [get]
func (s *service) Do(c echo.Context) error {
	log := zerolog.Ctx(c.Request().Context()).With().Logger()
	r := c.Request()
	type (
		MyRequest struct {
			Headers http.Header
			Args    map[string][]string
			Form    interface{}
		}
	)
	// decode the form
	err := r.ParseForm()
	if err != nil {
		log.Error().Err(err).Msg("Do")
		return err
	}

	myRequest := MyRequest{
		Headers: r.Header,
		Form:    r.Form,
		Args:    r.URL.Query(),
	}
	log.Info().Interface("myRequest", myRequest).Msg("Do")
	grantType := r.Form.Get("grant_type")
	switch grantType {
	case "authorization_code":
		return s.handleAuthorizationCodeRequest(c)
	}
	log.Error().Msgf("grant_type: %s", grantType)
	return c.JSON(http.StatusBadRequest, "unsupported_grant_type")
}

func (s *service) handleAuthorizationCodeRequest(c echo.Context) error {
	log := zerolog.Ctx(c.Request().Context()).With().Logger()

	r := c.Request()
	redirectURI := r.Form.Get("redirect_uri")
	code := r.Form.Get("code")
	if redirectURI != s.config.AuthorizedRedirectUrl {
		log.Error().Msgf("redirect_uri: %s", redirectURI)
		return c.JSON(http.StatusBadRequest, "invalid_grant")
	}
	// pull the basic auth from the header
	basicAuth := r.Header.Get("Authorization")

	response, err := s.downstreamService.ExchangeCodeForToken(basicAuth, code, redirectURI)
	if err != nil {
		log.Error().Err(err).Msg("ExchangeCodeForToken")
		return c.JSON(http.StatusBadRequest, "could not exchange code for token")
	}
	log.Info().Interface("response", response).Msg("ExchangeCodeForToken")
	return c.JSON(http.StatusOK, response)

}

/*
{
  "Args": {},
  "Form": {
    "code": [
      "4/0AfJohXmvwrYuWtGQHoMbb2xUgnNYI9dqFrVNghH2FIlMT-e5nIUEFxmfPpzQGY_Vqg57iw"
    ],
    "grant_type": ["authorization_code"],
    "redirect_uri": ["http://localhost:5556/auth/callback"]
  },
  "Headers": {
    "Accept-Encoding": ["gzip"],
    "Authorization": [
      "Basic MTA5NjMwMTYxNjU0Ni1maWJtZXBvNmNiaHJtdGM3dWpkODd2OW1udGJzbjUyMy5hcHBzLmdvb2dsZXVzZXJjb250ZW50LmNvbTpnT0t3bU4xODFDZ3NuUVFEV3FUU1pqRnM="
    ],
    "Content-Length": ["171"],
    "Content-Type": ["application/x-www-form-urlencoded"],
    "User-Agent": ["Go-http-client/1.1"]
  }
}
*/