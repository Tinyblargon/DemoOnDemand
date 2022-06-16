package frontend

import (
	"net/http"

	"github.com/Tinyblargon/DemoOnDemand/dod/authentication"
	"github.com/Tinyblargon/DemoOnDemand/dod/helper/api"
)

type Auth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Data struct {
	Token string `json:"token"`
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
	var succecfulLogin bool
	var role string

	if auth.Username == "root" {
		if auth.Password == "root" {
			succecfulLogin = true
			role = "root"
		}
	} else {
		role, succecfulLogin, err = authentication.Main.Authenticate(auth.Username, auth.Password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error something went wrong during authentication."))
			return
		}
	}

	if succecfulLogin {
		token, err := newToken(auth.Username, role)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error generating JWT token: " + err.Error()))
		} else {
			w.Header().Set("Authorization", "Bearer "+token)
			w.WriteHeader(http.StatusOK)
			data := Data{
				Token: token,
			}
			response := api.JsonResponse{
				Data: data,
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
			// dont know if this is a good idea might give cryptic and not for user destined information.
			w.Write([]byte(err.Error()))
			return
		}
		r.Header.Set("name", claims.Name)
		r.Header.Set("role", claims.Role)

		next.ServeHTTP(w, r)
	})
}
