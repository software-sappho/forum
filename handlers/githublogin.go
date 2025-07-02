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
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
)

var (
	githubClientID     string
	githubClientSecret string
)

func init() {
	githubClientID = os.Getenv("GITHUB_CLIENT_ID")
	if githubClientID == "" {
		log.Fatal("GITHUB_CLIENT_ID environment variable not set")
	}

	githubClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")
	if githubClientSecret == "" {
		log.Fatal("GITHUB_CLIENT_SECRET environment variable not set")
	}
}

func HandleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	state := uuid.Must(uuid.NewV4()).String()

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   300,
		SameSite: http.SameSiteLaxMode,
		Secure:   false, // Set to true in production with HTTPS
	})

	redirectURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=user:email&state=%s&allow_signup=true",
		githubClientID,
		url.QueryEscape("http://localhost:8080/auth/github/callback"),
		state,
	)

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func exchangeCodeForGitHubToken(code string) (struct {
	AccessToken string `json:"access_token"`
}, error) {
	data := url.Values{}
	data.Set("client_id", githubClientID)
	data.Set("client_secret", githubClientSecret)
	data.Set("code", code)

	req, err := http.NewRequest(
		"POST",
		"https://github.com/login/oauth/access_token",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return struct {
			AccessToken string `json:"access_token"`
		}{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return struct {
			AccessToken string `json:"access_token"`
		}{}, err
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return struct {
			AccessToken string `json:"access_token"`
		}{}, err
	}

	return tokenResp, nil
}

func HandleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	// Validate state parameter
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

	tokenResp, err := exchangeCodeForGitHubToken(code)
	if err != nil {
		ClearOAuthStateCookie(w)
		RenderErrorPage(w, r, http.StatusInternalServerError, "Failed to exchange code for token")
		return
	}

	githubUser, err := getGitHubUserInfo(tokenResp.AccessToken)
	if err != nil {
		ClearOAuthStateCookie(w)
		RenderErrorPage(w, r, http.StatusInternalServerError, "Failed to get user info")
		return
	}

	// Validate we have required fields before proceeding
	if githubUser.ID == 0 || githubUser.Login == "" || githubUser.Email == "" {
		ClearOAuthStateCookie(w)
		RenderErrorPage(w, r, http.StatusInternalServerError, "Incomplete user information from GitHub")
		return
	}

	user, err := findOrCreateUserFromGitHub(githubUser)
	if err != nil {
		ClearOAuthStateCookie(w)
		RenderErrorPage(w, r, http.StatusInternalServerError, "Failed to process user account")
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

func getGitHubUserInfo(accessToken string) (struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
}, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return struct {
			ID    int    `json:"id"`
			Login string `json:"login"`
			Email string `json:"email"`
		}{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("User-Agent", "forum-app")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return struct {
			ID    int    `json:"id"`
			Login string `json:"login"`
			Email string `json:"email"`
		}{}, err
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID    int    `json:"id"`
		Login string `json:"login"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return struct {
			ID    int    `json:"id"`
			Login string `json:"login"`
			Email string `json:"email"`
		}{}, err
	}

	// If email is null, try to get it from the emails endpoint
	if userInfo.Email == "" {
		email, err := getGitHubPrimaryEmail(accessToken)
		if err != nil {
			return struct {
				ID    int    `json:"id"`
				Login string `json:"login"`
				Email string `json:"email"`
			}{}, fmt.Errorf("failed to get email from GitHub: %v", err)
		}
		userInfo.Email = email
	}

	return userInfo, nil
}

func getGitHubPrimaryEmail(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("User-Agent", "forum-app")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}

	return "", fmt.Errorf("no verified primary email found")
}

func findOrCreateUserFromGitHub(githubUser struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
}) (*models.User, error) {
	var user models.User
	var nullableGithubID sql.NullInt64
	var nullableGoogleID sql.NullString

	// ✅ Lookup by GitHub ID only
	err := database.DB.QueryRow(`
        SELECT id, username, email, github_id, google_id
        FROM users
        WHERE github_id = ?`,
		githubUser.ID,
	).Scan(&user.ID, &user.Username, &user.Email, &nullableGithubID, &nullableGoogleID)

	if err == nil {
		// Found existing user with matching GitHub ID
		if nullableGithubID.Valid {
			user.GithubID = nullableGithubID.Int64
		}
		if nullableGoogleID.Valid {
			user.GoogleID = nullableGoogleID.String
		}
		return &user, nil
	} else if err != sql.ErrNoRows {
		// Unexpected DB error
		return nil, fmt.Errorf("database error finding user by GitHub ID: %v", err)
	}

	// ✅ No existing user, create new user
	res, err := database.DB.Exec(`
        INSERT INTO users (username, email, github_id)
        VALUES (?, ?, ?)`,
		githubUser.Login, githubUser.Email, githubUser.ID,
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
		Username: githubUser.Login,
		Email:    githubUser.Email,
		GithubID: int64(githubUser.ID),
	}, nil
}

func createUserSessionInDB(userID int) (string, time.Time, error) {
	sessionToken, err := generateSessionToken()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate session token: %v", err)
	}

	expiration := time.Now().Add(24 * time.Hour)
	_, err = database.DB.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to clear old sessions: %v", err)
	}

	_, err = database.DB.Exec(
		"INSERT INTO sessions (user_id, session_token, expires_at) VALUES (?, ?, ?)",
		userID, sessionToken, expiration,
	)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to create session: %v", err)
	}

	return sessionToken, expiration, nil
}

func setSessionCookie(w http.ResponseWriter, token string, expiration time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Expires:  expiration,
		SameSite: http.SameSiteLaxMode,
		Secure:   false, // Set to true in production with HTTPS
	})
}
