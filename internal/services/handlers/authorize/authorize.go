package authorize

import (
	"fmt"
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
			contracts_handler.GET,
		},
		wellknown.AuthorizationPath,
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
// @Router /authorization [get]
func (s *service) Do(c echo.Context) error {
	log := zerolog.Ctx(c.Request().Context()).With().Logger()
	baseUrl := "http://" + c.Request().Host

	// http://localhost:9044/authorization?client_id=1096301616546-fibmepo6cbhrmtc7ujd87v9mntbsn523.apps.googleusercontent.com&prompt=login&redirect_uri=http%3A%2F%2Flocalhost%3A5556%2Fauth%2Fcallback&response_type=code&scope=openid+profile+email&state=d279b227-4a0d-4ec9-8d2c-9e901dce6999
	r := c.Request()

	type (
		MyRequest struct {
			Headers http.Header
			Args    map[string][]string
			Body    interface{}
		}
	)
	myRequest := MyRequest{
		Headers: r.Header,
		Body:    r.Body,
		Args:    r.URL.Query(),
	}
	log.Info().Interface("myRequest", myRequest).Msg("Do")
	myState := r.URL.Query().Get("state")
	discoveryDocument, err := s.downstreamService.GetDiscoveryDocument()
	if err != nil {
		return err
	}
	clientId := r.URL.Query().Get("client_id")
	myRedirectUri := fmt.Sprintf("%s%s", baseUrl, wellknown.SigninGooglePath)

	authorizationEndpoint := discoveryDocument.AuthorizationEndpoint + "?client_id=" + clientId + "&response_type=code&scope=openid+profile+email&state=" + string(myState) + "&redirect_uri=" + myRedirectUri
	return c.Redirect(http.StatusFound, authorizationEndpoint)
	// return c.JSON(http.StatusOK, myRequest)
}
