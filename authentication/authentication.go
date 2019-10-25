package authentication

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	"github.com/yousseffarkhani/playground/backend2/configuration"
)

func InitAuthentication() {
	jwtKey = []byte(configuration.Variables.JWT_SECRET)
	gothic.Store = sessions.NewCookieStore([]byte(configuration.Variables.SESSION_SECRET))
	setupGothProviders(configuration.Variables.ProductionMode)
}

func setupGothProviders(productionMode bool) {
	var callbackBaseURL string
	if productionMode {
		callbackBaseURL = "https://playground.yousseffarkhani.website"
	} else {
		callbackBaseURL = "http://localhost:5000"
	}
	goth.UseProviders(
		facebook.New(configuration.Variables.FacebookOAuth.ID, configuration.Variables.FacebookOAuth.Secret, fmt.Sprintf("%s/auth/callback/facebook", callbackBaseURL)),
		google.New(configuration.Variables.GoogleOAuth.ID, configuration.Variables.GoogleOAuth.Secret, fmt.Sprintf("%s/auth/callback/google", callbackBaseURL)),
		github.New(configuration.Variables.GithubOAuth.ID, configuration.Variables.GithubOAuth.Secret, fmt.Sprintf("%s/auth/callback/github", callbackBaseURL)),
	)
}

func SetJwtCookie(w http.ResponseWriter, username string) {
	validToken, expirationTime, err := generateJWT(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "Token",
		Value:   validToken,
		Expires: expirationTime,
		Path:    "/",
	})
}

func UnsetJWTCookie(w http.ResponseWriter) {
	expired := time.Now().Add(time.Minute * -5)
	http.SetCookie(w, &http.Cookie{
		Name:    "Token",
		Value:   "",
		Expires: expired,
	})
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

var jwtKey []byte

func generateJWT(username string) (string, time.Time, error) {
	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		fmt.Printf("Something went wrong: %v", err)
		return "", time.Time{}, err
	}
	return tokenString, expirationTime, nil
}

func ParseCookie(c *http.Cookie) (*Claims, *jwt.Token, error) {
	tokenString := c.Value
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, nil, err
	}
	return claims, token, nil
}
