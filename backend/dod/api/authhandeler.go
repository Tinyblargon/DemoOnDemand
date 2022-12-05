package frontend

import (
	"fmt"
	"net/http"

	"github.com/Tinyblargon/DemoOnDemand/dod/authentication"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/api"
)

var rootUser string
var rootPassword string
var cookieSecret []byte
var tokenISS string //"iss" (Issuer) Claim
var tokenEXP uint   //"exp" (Expiration Time) Claim

type Auth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Data struct {
	Token string `json:"token"`
}

func Initialize(user, password, CookieSecret, tokenIssuer string, tokenExpirationTime uint) error {
	if rootUser != "" {
		return fmt.Errorf("user can only be set once")
	}
	if rootPassword != "" {
		return fmt.Errorf("password can only be set once")
	}
	if len(cookieSecret) != 0 {
		return fmt.Errorf("cookieSecret can only be set once")
	}
	if tokenISS != "" {
		return fmt.Errorf("tokenIssuer can only be set once")
	}
	if tokenEXP != 0 {
		return fmt.Errorf("tokenExpirationTime can only be set once")
	}
	rootUser = user
	rootPassword = password
	if CookieSecret == "" {
		return fmt.Errorf("cookieSecret may not be empty")
	}
	cookieSecret = []byte(CookieSecret)
	if tokenIssuer == "" {
		return fmt.Errorf("cookieSecret may not be empty")
	}
	tokenISS = tokenIssuer
	if tokenExpirationTime < 60 {
		return fmt.Errorf("tokenExpirationTime must be greater than 60")
	}
	tokenEXP = tokenExpirationTime
	return nil
}

func authenticate(w http.ResponseWriter, r *http.Request) {
	auth := Auth{}
	err := api.GetBody(r, &auth)
	if err != nil {
		api.OutputUserInputError(w, err.Error())
		return
	}

	if len(auth.Username) == 0 || len(auth.Password) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please provide username and password to obtain the token"))
		return
	}
	var successfulLogin bool
	var role string

	if auth.Username == rootUser && rootUser != "" {
		if auth.Password == rootPassword && rootPassword != "" {
			successfulLogin = true
			role = "root"
		}
	} else {
		role, successfulLogin, err = authentication.Main.Authenticate(auth.Username, auth.Password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error something went wrong during authentication."))
			return
		}
	}

	if successfulLogin {
		token, err := newToken(auth.Username, role)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error generating JWT token: " + err.Error()))
		} else {
			response := api.JsonResponse{
				Data: Data{
					Token: token,
				},
			}
			response.Output(w)
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Invalid username or password."))
	}
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if len(tokenString) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Missing Authorization Header"))
			return
		}
		claims, err := verifyToken(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			// don't know if this is a good idea might give cryptic and not for user destined information.
			w.Write([]byte(err.Error()))
			return
		}
		r.Header.Add("name", claims.Name)
		r.Header.Add("role", claims.Role)
		next.ServeHTTP(w, r)
	})
}
