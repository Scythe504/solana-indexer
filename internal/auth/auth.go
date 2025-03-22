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
	err := godotenv.Load()
	var (
		key = os.Getenv("SECRET_KEY")
		maxAge = 86400 * 30
		isProd = os.Getenv("APP_ENV") == "local"
	)
	if err != nil {
		log.Fatalf("Failed to load env variables\n %s", err)
	}

	var (
		googleClientId = os.Getenv("GOOGLE_CLIENT_ID")
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