package auth

import (
	"fmt"
	"log"
	"os"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

func NewAuth() {
	envPaths := []string{
		".env",
		"/etc/secrets/.env",
	}
	envLoaded := false

	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			log.Printf("Loaded env from %s", path)
			envLoaded = true
			break
		}
	}

	if !envLoaded {
		log.Fatal("No .env files exist")
		return
	}

	var (
		key    = os.Getenv("SECRET_KEY")
		maxAge = 86400 * 30
		isProd = os.Getenv("APP_ENV") == "local"
	)

	var (
		googleClientId     = os.Getenv("GOOGLE_CLIENT_ID")
		googleClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	)

	store := sessions.NewCookieStore([]byte(key))

	store.MaxAge(maxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = isProd

	gothic.Store = store

	goth.UseProviders(
		google.New(googleClientId, googleClientSecret, fmt.Sprintf("%s/auth/callback/google", os.Getenv("PUBLIC_URL"))),
	)
}
