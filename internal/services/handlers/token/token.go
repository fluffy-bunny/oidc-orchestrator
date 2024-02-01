package token

import (
	"fmt"
	"net/http"
	"time"

	di "github.com/fluffy-bunny/fluffy-dozm-di"
	contracts_handler "github.com/fluffy-bunny/fluffycore/echo/contracts/handler"
	mocks_oauth2 "github.com/fluffy-bunny/fluffycore/mocks/oauth2"
	core_utils "github.com/fluffy-bunny/fluffycore/utils"
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
	fmt.Println("---------------->TOKEN Do")
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
	case "refresh_token":
		return s.handleRefreshTokenRequest(c)
	}
	log.Error().Msgf("grant_type: %s", grantType)
	return c.JSON(http.StatusBadRequest, "unsupported_grant_type")
}
func (s *service) handleRefreshTokenRequest(c echo.Context) error {
	// 1. Pull the wrapped downstream token and use it against the downstream token endpoint.
	// 2. If successfull, create our access_token
	// 2.1 If the refresh token is valid, create a new wrapped refresh_token
	// 3. Return the access_token and refresh_token

	ctx := c.Request().Context()

	log := zerolog.Ctx(ctx).With().Logger()
	r := c.Request()
	baseUrl := "http://" + r.Host

	refreshToken := r.Form.Get("refresh_token")
	if core_utils.IsEmptyOrNil(refreshToken) {
		return c.JSON(http.StatusBadRequest, "refresh_token is missing")
	}
	scope := r.Form.Get("scope")

	log = log.With().Str("refreshToken", refreshToken).Logger()
	claims, err := mocks_oauth2.ValidateToken(ctx, refreshToken)
	if err != nil {
		log.Error().Err(err).Msg("handleRefreshTokenRequest")
		return c.JSON(http.StatusBadRequest, "bad refresh_token")
	}
	drt := claims.Get("downstream_refresh_token").(string)
	baseAccessToken := claims.Get("base_access_token").(mocks_oauth2.IClaims)

	// pull the basic auth from the header
	basicAuth := r.Header.Get("Authorization")
	request := &contracts_downstream.RefreshTokenRequest{
		RefreshToken: drt,
		Scope:        scope,
	}
	response, err := s.downstreamService.RefreshToken(ctx, basicAuth, request)
	if err != nil {
		log.Error().Err(err).Msg("handleRefreshTokenRequest")
		return c.JSON(http.StatusBadRequest, "could not refresh token")
	}
	// if we get here then the downstream token is good.  We don't care about anything it gave us back.
	// we need to mint a new refresh token based upon the access_token we stored in our refresh_token

	now := time.Now()
	myNewAccessToken := mocks_oauth2.NewClaims()
	for k, v := range baseAccessToken.Claims() {
		myNewAccessToken.Set(k, v)
	}
	myNewAccessToken.Set("exp", now.Add(time.Minute*30).Unix())
	myNewAccessToken.Set("iat", now.Unix())
	myAccessToken, _ := mocks_oauth2.MintToken(myNewAccessToken)
	response.AccessToken = myAccessToken

	rtClaims := mocks_oauth2.NewClaims()
	rtClaims.Set("iss", baseUrl)
	rtClaims.Set("downstream_refresh_token", response.RefreshToken)
	rtClaims.Set("base_access_token", baseAccessToken)
	myRefreshToken, _ := mocks_oauth2.MintToken(rtClaims)
	response.RefreshToken = myRefreshToken

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
