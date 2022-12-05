package frontend

import (
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
)

type MyCustomClaims struct {
	Name string `json:"name"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func newToken(name, role string) (string, error) {
	t := time.Now().Add(time.Duration(tokenEXP) * time.Second)
	claims := MyCustomClaims{
		name,
		role,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(t),
			Issuer:    tokenISS,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(cookieSecret)
	return tokenString, err
}

func verifyToken(tokenString string) (claims *MyCustomClaims, err error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return cookieSecret, nil
	})
	if err != nil {
		return
	}
	claims = token.Claims.(*MyCustomClaims)
	return
}
