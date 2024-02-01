package token

import (
	"context"
	"net/http"
	"time"

	mocks_oauth2 "github.com/fluffy-bunny/fluffycore/mocks/oauth2"
	fluffycore_utils "github.com/fluffy-bunny/fluffycore/utils"
	echo "github.com/labstack/echo/v4"
	jwxt "github.com/lestrrat-go/jwx/v2/jwt"
	zerolog "github.com/rs/zerolog"
)

func (s *service) handleAuthorizationCodeRequest(c echo.Context) error {
	log := zerolog.Ctx(c.Request().Context()).With().Logger()
	ctx := c.Request().Context()
	r := c.Request()
	baseUrl := "http://" + c.Request().Host

	redirectURI := r.Form.Get("redirect_uri")
	code := r.Form.Get("code")

	// pull the basic auth from the header
	basicAuth := r.Header.Get("Authorization")
	log.Info().Msgf("calling ExchangeCodeForToken")
	response, err := s.downstreamService.ExchangeCodeForToken(context.Background(), basicAuth, code, redirectURI)
	if err != nil {
		log.Error().Err(err).Msg("ExchangeCodeForToken")
		return c.JSON(http.StatusBadRequest, "could not exchange code for token")
	}
	// crack open hte id_token
	claims := mocks_oauth2.NewClaims()
	notTrustedToken, err := jwxt.ParseString(response.IDToken,
		jwxt.WithValidate(false),
		jwxt.WithVerify(false))

	if err != nil {
		log.Error().Err(err).Msg("ExchangeCodeForToken")
		return c.JSON(http.StatusBadRequest, "could not parse id_token")
	}
	tokenMap, err := notTrustedToken.AsMap(ctx)
	if err != nil {
		log.Error().Err(err).Msg("ExchangeCodeForToken")
		return c.JSON(http.StatusBadRequest, "could not parse id_token")
	}
	iat := tokenMap["iat"].(time.Time)
	exp := tokenMap["exp"].(time.Time)
	_, ok := tokenMap["nbf"]
	if ok {
		nbf := tokenMap["nbf"].(time.Time)
		tokenMap["nbf"] = nbf.Unix()
	}
	tokenMap["iat"] = iat.Unix()
	tokenMap["exp"] = exp.Unix()
	for k, v := range tokenMap {
		claims.Set(k, v)
	}
	claims.Set("iss", baseUrl)

	log.Info().Interface("claims", claims).Msg("ExchangeCodeForToken")
	myIdToken, _ := mocks_oauth2.MintToken(claims)
	response.IDToken = myIdToken

	// build out the access_token
	// here we transfer over some minimal claims so that we just echo them back in our user_info api
	// this is also where you would do a token exchange and get the full claims of what the user needs.
	claims = mocks_oauth2.NewClaims()
	claims.Set("iss", baseUrl)
	claims.Set("sub", tokenMap["sub"])
	claims.Set("email", tokenMap["email"])
	claims.Set("aud", "myaud")
	claims.Set("permissions", []string{
		"permission.one",
		"permission.two",
		"permission.three",
	})
	baseAccessToken := claims.Claims()
	now := time.Now()
	claims.Set("exp", now.Add(time.Minute*30).Unix())
	claims.Set("iat", now.Unix())
	myAccessToken, _ := mocks_oauth2.MintToken(claims)
	response.AccessToken = myAccessToken
	if !fluffycore_utils.IsEmptyOrNil(response.RefreshToken) {
		// wrap the refresh_token in ours.
		claims = mocks_oauth2.NewClaims()
		claims.Set("downstream_refresh_token", response.RefreshToken)
		claims.Set("iss", baseUrl)
		claims.Set("base_access_token", baseAccessToken)
		myRefreshToken, _ := mocks_oauth2.MintToken(claims)
		response.RefreshToken = myRefreshToken
	}

	log.Info().Interface("response", response).Msg("ExchangeCodeForToken")
	return c.JSON(http.StatusOK, response)

}
