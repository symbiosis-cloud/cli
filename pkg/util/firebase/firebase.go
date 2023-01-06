package firebase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
)

type Token struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    string `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	UserID       string `json:"user_id"`
	ProjectID    string `json:"project_id"`
}

type RefreshTokenInput struct {
	GrantType    string `json:"grantType"`
	RefreshToken string `json:"refreshToken"`
}

var (
	FirebaseToken string
)

func ValidateToken(refreshToken string) error {
	if viper.GetString("auth.method") == "token" {

		token := viper.GetString("auth.token")
		tokenExpired := false

		if token != "" {
			claims := jwt.MapClaims{}
			jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
				return nil, nil
			})

			if claims["exp"] == nil {
				return fmt.Errorf("Provided token is invalid")
			}

			timeSec := time.Unix(int64(claims["exp"].(float64)), 0)
			tokenExpired = time.Now().UTC().After(timeSec.UTC())
		}

		apiEndpoint := viper.GetString("api_url")

		// if the access token has expired we have to renew it
		if tokenExpired || token == "" {

			payload, err := json.Marshal(RefreshTokenInput{"refresh_token", refreshToken})

			if err != nil {
				return err
			}

			res, err := http.Post(fmt.Sprintf("%s/rest/v1/firebase/refresh-token", apiEndpoint), "application/json", bytes.NewBuffer(payload))

			if err != nil {
				return err
			}

			if res.StatusCode != http.StatusOK {
				return fmt.Errorf("Unexpected error trying to refreshing authentication token")
			}

			response, err := io.ReadAll(res.Body)

			var newToken *Token

			if err != nil {
				return err
			}

			err = json.Unmarshal(response, &newToken)

			viper.Set("auth.refresh_token", newToken.RefreshToken)
			viper.Set("auth.token", newToken.AccessToken)
			viper.WriteConfig()

		}
	}

	return nil
}
