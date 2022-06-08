package frontend

import (
	"fmt"
	"time"

	"github.com/Tinyblargon/DemoOnDemand/dod/global"
	jwt "github.com/golang-jwt/jwt/v4"
)

type MyCustomClaims struct {
	Name string `json:"name"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func newToken(name, role string) (string, error) {
	signingKey := global.CookieSecret

	t := time.Now().Add(time.Second * 3600)
	claims := MyCustomClaims{
		name,
		role,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(t),
			Issuer:    "dod",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(signingKey)
	return tokenString, err
}

func verifyToken(tokenString string) (claims *MyCustomClaims, err error) {

	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return global.CookieSecret, nil
	})
	if err != nil {
		return
	}
	claims = token.Claims.(*MyCustomClaims)
	return
}
