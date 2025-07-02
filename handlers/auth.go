package handlers

import (
	"forum/database"
	"forum/models"
	"log"
	"net/http"
	"time"

	"regexp"

	uuid "github.com/gofrs/uuid/v5"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func verifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// validatePasswordRequirements checks if a password meets all security requirements
func validatePasswordRequirements(password string) (bool, string) {
	if len(password) < 8 {
		return false, "Password must be at least 8 characters long"
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		case char >= 33 && char <= 47 || char >= 58 && char <= 64 || char >= 91 && char <= 96 || char >= 123 && char <= 126:
			hasSpecial = true
		}
	}

	if !hasUpper {
		return false, "Password must contain at least one uppercase letter"
	}
	if !hasLower {
		return false, "Password must contain at least one lowercase letter"
	}
	if !hasNumber {
		return false, "Password must contain at least one number"
	}
	if !hasSpecial {
		return false, "Password must contain at least one special character"
	}

	return true, ""
}

// creates a new UUID v4 string to be used as a session token.
func generateSessionToken() (string, error) {
	u, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// createUserSession creates a new session for a user and sets the session cookie
func createUserSession(w http.ResponseWriter, userID int) error {
	// Generate a new session token (UUID)
	sessionToken, err := generateSessionToken()
	if err != nil {
		return err
	}

	// Delete any existing sessions for this user to avoid multiple sessions
	_, err = database.DB.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	if err != nil {
		return err
	}

	expiration := time.Now().Add(24 * time.Hour)
	_, err = database.DB.Exec("INSERT INTO sessions (user_id, session_token, expires_at) VALUES (?, ?, ?)",
		userID, sessionToken, expiration)
	if err != nil {
		return err
	}

	// Set session cookie in user's browser
	session := &http.Cookie{
		Name:     "session",
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true, // prevents JS access for security
		Expires:  expiration,
		SameSite: http.SameSiteStrictMode,
		// Secure: true, // Uncomment for HTTPS in production
	}
	http.SetCookie(w, session)

	return nil
}

// isValidEmail checks if the email format is valid
func isValidEmail(email string) bool {
	// Basic email regex pattern
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			RenderErrorPage(w, r, http.StatusBadRequest, "Error parsing form")
			return
		}

		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		// Prepare data for template
		data := models.PageData{
			IsLoggedIn: false,
			FormData: map[string]string{
				"username": username,
				"email":    email,
			},
		}

		// Validation checks
		errors := make(map[string]string)

		if password != confirmPassword {
			errors["password"] = "Passwords do not match"
		}

		// Email validation
		if !isValidEmail(email) {
			errors["email"] = "Please enter a valid email address"
		}

		// Password requirements validation
		valid, passwordMessage := validatePasswordRequirements(password)
		if !valid {
			errors["password"] = passwordMessage
		}

		// Check if email already exists in DB
		var count int
		err = database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
		if err != nil {
			RenderErrorPage(w, r, http.StatusInternalServerError, "Database error")
			return
		}
		if count > 0 {
			errors["email"] = "Email already registered"
		}

		// Check if username already exists in DB
		err = database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
		if err != nil {
			RenderErrorPage(w, r, http.StatusInternalServerError, "Database error")
			return
		}
		if count > 0 {
			errors["username"] = "Username already taken"
		}

		// If there are validation errors, render the form again with errors
		if len(errors) > 0 {
			data.Errors = errors
			RenderTemplate(w, "register", data)
			return
		}

		hashedPassword, err := hashPassword(password)
		if err != nil {
			RenderErrorPage(w, r, http.StatusInternalServerError, "Error securing password")
			return
		}

		// Insert new user into DB
		result, err := database.DB.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
			username, email, hashedPassword)
		if err != nil {
			RenderErrorPage(w, r, http.StatusInternalServerError, "Error creating user")
			return
		}

		// Get the newly created user's ID
		userID, err := result.LastInsertId()
		if err != nil {
			RenderErrorPage(w, r, http.StatusInternalServerError, "Error getting user ID")
			return
		}

		err = createUserSession(w, int(userID))
		if err != nil {
			RenderErrorPage(w, r, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Redirect logged-in user to /forum
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
		return
	}

	data := models.PageData{
		IsLoggedIn: false,
	}
	RenderTemplate(w, "register", data)
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Printf("Error parsing form: %v", err)
			RenderErrorPage(w, r, http.StatusBadRequest, "Error parsing form")
			return
		}

		emailOrUsername := r.FormValue("email")
		password := r.FormValue("password")

		log.Printf("Attempting login for: %s", emailOrUsername)

		// Prepare data for template
		data := models.PageData{
			IsLoggedIn: false,
			FormData: map[string]string{
				"email": emailOrUsername,
			},
		}

		// Validation checks
		errors := make(map[string]string)

		// Lookup user by email or username
		var user models.User
		err = database.DB.QueryRow("SELECT id, username, email, password FROM users WHERE email = ? OR username = ?", emailOrUsername, emailOrUsername).
			Scan(&user.ID, &user.Username, &user.Email, &user.Password)
		if err != nil {
			log.Printf("Database error or user not found: %v", err)
			errors["login"] = "Invalid email/username or password"
		} else if !verifyPassword(user.Password, password) {
			log.Printf("Password mismatch for user: %s", emailOrUsername)
			errors["login"] = "Invalid email/username or password"
		}

		// If there are validation errors, render the form again with errors
		if len(errors) > 0 {
			data.Errors = errors
			RenderTemplate(w, "login", data)
			return
		}

		log.Printf("Successful login for user: %s", emailOrUsername)

		err = createUserSession(w, user.ID)
		if err != nil {
			log.Printf("Error creating session: %v", err)
			RenderErrorPage(w, r, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Redirect logged-in user to /forum
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
		return
	}

	data := models.PageData{
		IsLoggedIn: false,
	}
	RenderTemplate(w, "login", data)
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		// Delete session record
		_, err := database.DB.Exec("DELETE FROM sessions WHERE session_token = ?", cookie.Value)
		if err != nil {
			log.Printf("Error deleting session from database: %v", err)
		}
	}

	// Expire the session cookie immediately to log out on client side
	session := &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(-1 * time.Hour),
		// Secure: true, // Uncomment for HTTPS in production
	}
	http.SetCookie(w, session)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
