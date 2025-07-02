package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"forum/database"
	"forum/models"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gofrs/uuid/v5"
)

var (
	googleClientID     string
	googleClientSecret string
)

func init() {
	googleClientID = os.Getenv("GOOGLE_CLIENT_ID")
	if googleClientID == "" {
		log.Fatal("GOOGLE_CLIENT_ID environment variable not set")
	}

	googleClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	if googleClientSecret == "" {
		log.Fatal("GOOGLE_CLIENT_SECRET environment variable not set")
	}
}

func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := uuid.Must(uuid.NewV4()).String()

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   300,
		SameSite: http.SameSiteLaxMode,
		Secure:   false, // Set to false for local testing without HTTPS
	})

	redirectURL := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=openid%%20email%%20profile&state=%s&prompt=select_account",
		googleClientID,
		url.QueryEscape("http://localhost:8080/auth/google/callback"),
		state,
	)

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	cookie, err := r.Cookie("oauth_state")
	if err != nil || cookie.Value != state {
		ClearOAuthStateCookie(w)
		RenderErrorPage(w, r, http.StatusInternalServerError, "Invalid OAuth state")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		ClearOAuthStateCookie(w)
		RenderErrorPage(w, r, http.StatusInternalServerError, "Code not found")
		return
	}

	tokenResp, err := exchangeCodeForToken(code)
	if err != nil {
		ClearOAuthStateCookie(w)
		RenderErrorPage(w, r, http.StatusInternalServerError, "Failed to exchange code for token")
		return
	}

	googleUser, err := getGoogleUserInfo(tokenResp.AccessToken)
	if err != nil {
		ClearOAuthStateCookie(w)
		RenderErrorPage(w, r, http.StatusInternalServerError, "Failed to get user info")
		return
	}

	user, err := findOrCreateUserFromGoogle(googleUser)
	if err != nil {
		ClearOAuthStateCookie(w)
		RenderErrorPage(w, r, http.StatusInternalServerError, "Failed to create user")
		return
	}

	sessionToken, expiration, err := createUserSessionInDB(user.ID)
	if err != nil {
		ClearOAuthStateCookie(w)
		RenderErrorPage(w, r, http.StatusInternalServerError, "Failed to create session")
		return
	}

	setSessionCookie(w, sessionToken, expiration)

	http.Redirect(w, r, "/forum", http.StatusSeeOther)

}

func exchangeCodeForToken(code string) (struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
}, error) {
	data := url.Values{}
	data.Set("client_id", googleClientID)
	data.Set("client_secret", googleClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", "http://localhost:8080/auth/google/callback")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return struct {
			AccessToken string `json:"access_token"`
			IDToken     string `json:"id_token"`
		}{}, err
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return tokenResp, err
	}

	return tokenResp, nil
}

func getGoogleUserInfo(accessToken string) (struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}, error) {
	userReq, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
		}{}, err
	}
	userReq.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	userResp, err := client.Do(userReq)
	if err != nil {
		return struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
		}{}, err
	}
	defer userResp.Body.Close()

	var googleUser struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(userResp.Body).Decode(&googleUser); err != nil {
		return googleUser, err
	}

	return googleUser, nil
}

func findOrCreateUserFromGoogle(googleUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}) (*models.User, error) {
	var user models.User
	var nullableGithubID sql.NullInt64
	var nullableGoogleID sql.NullString

	// ✅ Lookup by Google ID only
	err := database.DB.QueryRow(`
        SELECT id, username, email, google_id, github_id
        FROM users
        WHERE google_id = ?`,
		googleUser.ID,
	).Scan(&user.ID, &user.Username, &user.Email, &nullableGoogleID, &nullableGithubID)

	if err == nil {
		// Found existing user with matching Google ID
		if nullableGithubID.Valid {
			user.GithubID = nullableGithubID.Int64
		}
		if nullableGoogleID.Valid {
			user.GoogleID = nullableGoogleID.String
		}
		return &user, nil
	} else if err != sql.ErrNoRows {
		// Unexpected DB error
		return nil, fmt.Errorf("database error finding user by Google ID: %v", err)
	}

	// ✅ No existing user, create new user
	res, err := database.DB.Exec(`
        INSERT INTO users (username, email, google_id)
        VALUES (?, ?, ?)`,
		googleUser.Name, googleUser.Email, googleUser.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create new user: %v", err)
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get new user ID: %v", err)
	}

	return &models.User{
		ID:       int(lastID),
		Username: googleUser.Name,
		Email:    googleUser.Email,
		GoogleID: googleUser.ID,
	}, nil
}

// Helper function to clear oauth_state cookie
func ClearOAuthStateCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Delete cookie
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
	})
}
