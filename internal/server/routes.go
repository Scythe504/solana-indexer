package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/markbates/goth/gothic"
	"github.com/scythe504/solana-indexer/internal/database"
	"github.com/scythe504/solana-indexer/internal/utils"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := mux.NewRouter()

	// Apply CORS middleware
	r.Use(s.corsMiddleware)

	r.HandleFunc("/", s.HelloWorldHandler)

	r.HandleFunc("/health", s.healthHandler)

	r.HandleFunc("/auth/callback/{provider}", s.getAuthHandler)

	r.HandleFunc("/auth/{provider}", s.beginAuthHandler)

	r.HandleFunc("/logout/{provider}", s.logoutHandler)

	r.HandleFunc("/webhook/{str}", s.createWebhook)

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
		http.Redirect(w, r, "http://localhost:3000", http.StatusFound)
		return
	}
	now := time.Now()

	account := database.Account{
		ID:                 utils.GenerateUUID(),
		UserID:             dbUser.ID, // Use the database user ID here
		ProviderType:       "oidc",
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

	http.Redirect(w, r, "http://localhost:3000", http.StatusFound)
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
func (s *Server) createWebhook(w http.ResponseWriter, r *http.Request) {
    str := mux.Vars(r)["orca-webhook"]
    r = r.WithContext(context.WithValue(context.Background(), "str", str))

    var (
        heliusApiKey        = os.Getenv("HELIUS_API_KEY")
        heliusApiUrl        = os.Getenv("HELIUS_API_URL")
        heliusWebhookSecret = os.Getenv("HELIUS_WEBHOOK_SECRET")
        publicUrl           = os.Getenv("PUBLIC_URL")
    )

    // Check if required environment variables are set
    if heliusApiKey == "" || heliusApiUrl == "" || publicUrl == "" {
        log.Println("Missing required environment variables for Helius webhook")
        http.Error(w, "Server configuration error", http.StatusInternalServerError)
        return
    }

    body := map[string]interface{}{
        "webhookURL": fmt.Sprintf("%s/webhook/%s", publicUrl, str),
        "webhookType": "enhanced",
        "transactionTypes": []string{
            "NFT_SALE",
            "TRANSFER",
        },
        "accountAddresses": []string{"orcaEKTdK7LKz57vaAYr9QeNsVEPfiu6QeMU1kektZE"},
        "txnStatus": "success",
    }

    // Only add auth header if secret is available
    if heliusWebhookSecret != "" {
        body["authHeader"] = fmt.Sprintf("Bearer %s", heliusWebhookSecret)
    }

    jsonBody, err := json.Marshal(body)
    if err != nil {
        log.Printf("Error marshaling webhook request body: %v", err)
        http.Error(w, "Failed to create webhook request", http.StatusInternalServerError)
        return
    }

    // Create the request with proper URL
    url := fmt.Sprintf("%s/webhooks?api-key=%s", heliusApiUrl, heliusApiKey)
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
    if err != nil {
        log.Printf("Error creating request: %v", err)
        http.Error(w, "Failed to create webhook request", http.StatusInternalServerError)
        return
    }

    // Set content type header
    req.Header.Set("Content-Type", "application/json")

    // Send the request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Printf("Error sending webhook creation request: %v", err)
        http.Error(w, "Failed to connect to Helius API", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    // Read response body
    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Printf("Error reading response body: %v", err)
        http.Error(w, "Failed to read API response", http.StatusInternalServerError)
        return
    }

    // Check response status
    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        log.Printf("Helius API error: %s - %s", resp.Status, string(respBody))
        http.Error(w, fmt.Sprintf("Helius API error: %s", string(respBody)), http.StatusBadRequest)
        return
    }

    // Return the successful response to the client
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(respBody)
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
