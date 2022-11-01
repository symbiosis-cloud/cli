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

func ValidateToken() error {
	if viper.GetString("auth.method") == "token" {
		claims := jwt.MapClaims{}
		jwt.ParseWithClaims(viper.GetString("auth.token"), claims, func(token *jwt.Token) (interface{}, error) {
			return nil, nil
		})

		if claims["exp"] == nil {
			return fmt.Errorf("Provided token is invalid")
		}

		timeSec := time.Unix(int64(claims["exp"].(float64)), 0)
		tokenExpired := time.Now().UTC().After(timeSec.UTC())

		// if the access token has expired we have to renew it
		if tokenExpired {
			payload := bytes.NewBufferString(fmt.Sprintf("grant_type=refresh_token&refresh_token=%s", viper.GetString("auth.refresh_token")))
			res, err := http.Post("https://securetoken.googleapis.com/v1/token?key=AIzaSyBbGgIU15KOodwZXwH_e0OpKWLwt0udAz0", "application/x-www-form-urlencoded", payload)

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
