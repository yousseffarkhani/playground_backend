package authentication

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	gothic.Store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	goth.UseProviders(
		// TODO : change callback URLs
		facebook.New(os.Getenv("FACEBOOK_ID"), os.Getenv("FACEBOOK_SECRET"), "http://localhost:5000/auth/callback/facebook"),
		google.New(os.Getenv("GOOGLE_ID"), os.Getenv("GOOGLE_SECRET"), "http://localhost:5000/auth/callback/google"),
		github.New(os.Getenv("GITHUB_ID"), os.Getenv("GITHUB_SECRET"), "http://localhost:5000/auth/callback/github"),
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

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

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
