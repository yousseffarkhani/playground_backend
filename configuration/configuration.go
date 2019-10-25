package configuration

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var Variables *configVariables

type configVariables struct {
	ProductionMode           bool
	TLS                      TLS
	FacebookOAuth            OAuthLogin
	GoogleOAuth              OAuthLogin
	GithubOAuth              OAuthLogin
	JWT_SECRET               string
	SESSION_SECRET           string
	GOOGLE_MAPS_API_KEY      string
	GOOGLE_GEOCODING_API_KEY string
}

type TLS struct {
	PathToCertFile string
	PathToPrivKey  string
}

type OAuthLogin struct {
	ID     string
	Secret string
}

func LoadEnvVariables() {
	err := godotenv.Load()
	if err != nil {
		log.Print("No .env file found")
	}
	Variables = &configVariables{
		ProductionMode: getEnvAsBool("APP_ENV"),
		TLS: TLS{
			PathToCertFile: os.Getenv("CERTFILE"),
			PathToPrivKey:  os.Getenv("PRIVKEY"),
		},
		FacebookOAuth: OAuthLogin{
			ID:     os.Getenv("FACEBOOK_ID"),
			Secret: os.Getenv("FACEBOOK_SECRET"),
		},
		GoogleOAuth: OAuthLogin{
			ID:     os.Getenv("GOOGLE_ID"),
			Secret: os.Getenv("GOOGLE_SECRET"),
		},
		GithubOAuth: OAuthLogin{
			ID:     os.Getenv("GITHUB_ID"),
			Secret: os.Getenv("GITHUB_SECRET"),
		},
		JWT_SECRET:               os.Getenv("JWT_SECRET"),
		SESSION_SECRET:           os.Getenv("SESSION_SECRET"),
		GOOGLE_MAPS_API_KEY:      os.Getenv("GOOGLE_MAPS_API_KEY"),
		GOOGLE_GEOCODING_API_KEY: os.Getenv("GOOGLE_GEOCODING_API_KEY"),
	}
}

func getEnvAsBool(key string) bool {
	valStr := os.Getenv(key)
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return false
}
