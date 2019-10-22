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
		log.Print("No .env file found")
	}

	gothic.Store = sessions.NewCookieStore([]byte(getEnv("SESSION_SECRET", "")))
	goth.UseProviders(
		// TODO : change callback URLs
		facebook.New(getEnv("FACEBOOK_ID", ""), getEnv("FACEBOOK_SECRET", ""), "http://localhost:5000/auth/callback/facebook"),
		google.New(getEnv("GOOGLE_ID", ""), getEnv("GOOGLE_SECRET", ""), "http://localhost:5000/auth/callback/google"),
		github.New(getEnv("GITHUB_ID", ""), getEnv("GITHUB_SECRET", ""), "http://localhost:5000/auth/callback/github"),
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

var jwtKey = []byte(getEnv("JWT_SECRET", ""))

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

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
