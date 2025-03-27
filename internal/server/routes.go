package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/markbates/goth/gothic"
	"github.com/scythe504/solana-indexer/internal/database"
	"github.com/scythe504/solana-indexer/internal/kafka"
	"github.com/scythe504/solana-indexer/internal/utils"
)

type JwtClaims struct {
	UserId string `json:"userId"`
	jwt.RegisteredClaims
}

func (s *Server) RegisterRoutes() http.Handler {
	r := mux.NewRouter()

	// Apply CORS middleware
	r.Use(s.corsMiddleware)

	r.HandleFunc("/", s.HelloWorldHandler)

	r.HandleFunc("/health", s.healthHandler)

	r.HandleFunc("/auth/callback/{provider}", s.getAuthHandler)

	r.HandleFunc("/auth/{provider}", s.beginAuthHandler)

	r.HandleFunc("/logout/{provider}", s.logoutHandler)

	r.HandleFunc("/webhook/{receiverName}", s.handleWebhookReceiver)

	r.Use(s.authMiddleWare)

	r.HandleFunc("/create-database", s.createUserDatabase)

	r.HandleFunc("/index-token", s.indexAddress)

	return r
}

// CORS middleware
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS Headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Wildcard allows all origins
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "false") // Credentials not allowed with wildcard origins

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) authMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		tokenStrings := strings.Split(authHeader, " ")
		if authHeader == "" || tokenStrings[0] != "Bearer" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		jwtToken := tokenStrings[1]

		secretKey := []byte(os.Getenv("JWT_SECRET"))
		token, err := jwt.ParseWithClaims(jwtToken, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		}, jwt.WithValidMethods([]string{
			jwt.SigningMethodHS256.Alg(),
		}))

		if err != nil {
			http.Error(w, "Unauthorized, Invalid Token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*JwtClaims)

		if !ok || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Attach user ID to request context
		r = r.WithContext(context.WithValue(r.Context(), "userId", claims.UserId))

		next.ServeHTTP(w, r)
	})
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, err := json.Marshal(s.db.Health())

	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) handleWebhookReceiver(w http.ResponseWriter, r *http.Request) {
	receiverName := mux.Vars(r)["receiverName"]
	r = r.WithContext(context.WithValue(context.Background(), "receiverName", receiverName))

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading the request body: ", err)
		http.Error(w, "Error reading the request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = kafka.ParseBodyAndPushToProducer(s.kafka, body, receiverName)

	if err != nil {
		log.Println("Invalid Json or failed to push the data to kafka producer", err)
		http.Error(w, "Error parsing the body to json: ", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "success"}`))
}

func (s *Server) createUserDatabase(w http.ResponseWriter, r *http.Request) {
	var dbCredential database.UserDatabaseCredential
	userId := mux.Vars(r)["userId"]

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading the request body: ", err)
		http.Error(w, "Error reading the request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &dbCredential); err != nil {
		log.Println("Error occured while trying to parse body", err)
		http.Error(w, "Invalid Json Payload", http.StatusBadRequest)
		return
	}
	dbExists, _ := s.db.GetDatabaseConfig(dbCredential.UserId)

	if dbExists != nil {
		log.Println("Database config found for user with userId: ", dbCredential.UserId)
		http.Error(w, "Database config already exists", http.StatusBadRequest)
		return
	}

	err = s.db.CreateDatabaseForUser(userId, dbCredential)

	if err != nil {
		log.Println("Invalid database credentials for userId: ", dbCredential.UserId)
		http.Error(w, "Please check you database credentials or if your database is running", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": "Database credentials stored securely"}`))
}

func (s *Server) getAuthHandler(w http.ResponseWriter, r *http.Request) {

	providerName := mux.Vars(r)["provider"]

	r = r.WithContext(context.WithValue(context.Background(), "provider", providerName))

	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		fmt.Fprintln(w, r)
		return
	}

	// Let the database assign a UUID for the user (don't set ID here)
	dbUser := &database.User{
		Name:          &user.Name,
		Email:         &user.Email,
		EmailVerified: false,
		Image:         &user.AvatarURL,
	}

	existingUser, err := s.db.GetUserByEmail(*dbUser.Email)

	if err != nil {
		// CreateUser will generate a UUID if ID is empty
		err := s.db.CreateUser(dbUser)
		if err != nil {
			log.Printf("Error creating user: %v", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
	} else {
		dbUser.ID = existingUser.ID
	}

	existingAccount, _ := s.db.GetUserByProviderId(user.UserID)

	if existingAccount != nil {
		fmt.Println("Account Already Exists in Db")
		http.Redirect(w, r, os.Getenv("FRONTEND_URL"), http.StatusFound)
		return
	}
	now := time.Now()

	account := database.Account{
		ID:                 utils.GenerateUUID(),
		UserID:             dbUser.ID, // Use the database user ID here
		ProviderType:       "oauth2",
		ProviderID:         "google.com",
		ProviderAccountID:  user.UserID,
		RefreshToken:       &user.RefreshToken,
		AccessToken:        &user.AccessToken,
		AccessTokenExpires: &user.ExpiresAt,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	err = s.db.CreateAccount(&account)

	if err != nil {
		log.Printf("Error creating account: %v", err)
		http.Error(w, "Failed to create account", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s/onboarding/database", os.Getenv("FRONTEND_URL")), http.StatusFound)
}

func (s *Server) beginAuthHandler(w http.ResponseWriter, r *http.Request) {
	// try to get the user without re-authenticating
	if gothUser, err := gothic.CompleteUserAuth(w, r); err == nil {
		t, _ := template.New("foo").Parse(userTemplate)
		t.Execute(w, gothUser)
	} else {
		gothic.BeginAuthHandler(w, r)
	}
}

func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {
	gothic.Logout(w, r)
	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (s *Server) indexAddress(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Failed to ready body")
		http.Error(w, "Invalid Body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var addressData database.Subscription
	if err = json.Unmarshal(body, &addressData); err != nil {
		log.Println("Failed to Unmarshal ")
		http.Error(w, "Unable to parse body: ", http.StatusInternalServerError)
		return
	}

	userId := mux.Vars(r)["userId"]

	if err = s.db.CreateSubscription(
		addressData.TokenAddress,
		addressData.Strategies,
		userId,
	); err != nil {
		log.Println("Error occured while creating subscriptions, err: ", err)
		http.Error(w, "Failed to create indexing for the given address", http.StatusInternalServerError)
		return
	}

	

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`"success": "token indexing started"`))
}

var userTemplate = `
<p><a href="/logout/{{.Provider}}">logout</a></p>
<p>Name: {{.Name}} [{{.LastName}}, {{.FirstName}}]</p>
<p>Email: {{.Email}}</p>
<p>NickName: {{.NickName}}</p>
<p>Location: {{.Location}}</p>
<p>AvatarURL: {{.AvatarURL}} <img src="{{.AvatarURL}}"></p>
<p>Description: {{.Description}}</p>
<p>UserID: {{.UserID}}</p>
<p>AccessToken: {{.AccessToken}}</p>
<p>ExpiresAt: {{.ExpiresAt}}</p>
<p>RefreshToken: {{.RefreshToken}}</p>
`
