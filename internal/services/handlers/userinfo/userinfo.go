package userinfo

import (
	"context"
	"net/http"

	di "github.com/fluffy-bunny/fluffy-dozm-di"
	contracts_handler "github.com/fluffy-bunny/fluffycore/echo/contracts/handler"
	mocks_oauth2 "github.com/fluffy-bunny/fluffycore/mocks/oauth2"
	contracts_config "github.com/fluffy-bunny/oidc-orchestrator/internal/contracts/config"
	contracts_downstream "github.com/fluffy-bunny/oidc-orchestrator/internal/contracts/downstream"
	wellknown "github.com/fluffy-bunny/oidc-orchestrator/internal/wellknown"
	echo "github.com/labstack/echo/v4"
	jwxt "github.com/lestrrat-go/jwx/v2/jwt"
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
		wellknown.SigninGooglePath,
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

// Do UserInfo
func (s *service) Do(c echo.Context) error {
	log := zerolog.Ctx(c.Request().Context()).With().Logger()
	//baseUrl := "http://" + c.Request().Host
	//myRedirectUri := fmt.Sprintf("%s%s", baseUrl, wellknown.SigninGooglePath)

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
	basicAuth := r.Header.Get("Authorization")

	claims := mocks_oauth2.NewClaims()
	notTrustedToken, err := jwxt.ParseString(basicAuth,
		jwxt.WithValidate(false),
		jwxt.WithVerify(false))
	if err != nil {
		log.Error().Err(err).Msg("ExchangeCodeForToken")
		return c.JSON(http.StatusBadRequest, "could not parse id_token")
	}
	tokenMap, err := notTrustedToken.AsMap(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("ExchangeCodeForToken")
		return c.JSON(http.StatusBadRequest, "could not parse id_token")
	}
	for k, v := range tokenMap {
		claims.Set(k, v)
	}

	return c.JSON(http.StatusOK, claims)

}
