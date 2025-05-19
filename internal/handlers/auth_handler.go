package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8" // For more robust validation if needed

	"github.com/google/uuid"
	"github.com/twomotive/dropwise/internal/auth"
	"github.com/twomotive/dropwise/internal/config"
	db "github.com/twomotive/dropwise/internal/database/sqlc"
	"github.com/twomotive/dropwise/internal/server/httputils"
)

// AuthHandler handles HTTP requests for authentication.
type AuthHandler struct {
	APIConfig *config.APIConfig
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(apiCfg *config.APIConfig) *AuthHandler {
	return &AuthHandler{APIConfig: apiCfg}
}

// --- Request Structs ---

// RegisterUserRequest defines the expected request body for user registration.
type RegisterUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginUserRequest defines the expected request body for user login.
type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// --- Response Structs ---

// UserResponse defines the user information returned to the client.
// It excludes sensitive information like the password hash.
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LoginResponse defines the response body for a successful login.
type LoginResponse struct {
	Token  string    `json:"token"`
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}

// Helper to convert db.CreateUserRow to UserResponse
func toUserResponseFromCreate(dbUser db.CreateUserRow) UserResponse {
	return UserResponse{
		ID:        dbUser.ID,
		Email:     dbUser.Email,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
	}
}

// --- Handler Implementations ---

// RegisterHandler handles new user registration.
// POST /api/v1/auth/register
func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.RespondWithError(w, http.StatusMethodNotAllowed, "Only POST method is allowed")
		return
	}

	var req RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}
	defer r.Body.Close()

	// Basic Input Validation
	req.Email = strings.TrimSpace(req.Email)
	// A more robust email validation might use a regex or a specialized library,
	// but for now, checking for non-empty and presence of "@" is a basic step.
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		httputils.RespondWithError(w, http.StatusBadRequest, "Valid email is required")
		return
	}
	if utf8.RuneCountInString(req.Password) < 8 { // Example: minimum 8 characters
		httputils.RespondWithError(w, http.StatusBadRequest, "Password must be at least 8 characters long")
		return
	}

	log.Printf("Attempting to register user with email: %s", req.Email)

	// Check if user already exists
	_, err := h.APIConfig.DB.GetUserByEmail(r.Context(), req.Email)
	if err == nil {
		// User found, so email is already taken
		log.Printf("Registration failed: email %s already exists", req.Email)
		httputils.RespondWithError(w, http.StatusConflict, "Email already registered")
		return
	}
	if err != sql.ErrNoRows {
		// An actual database error occurred
		log.Printf("Error checking for existing user %s: %v", req.Email, err)
		httputils.RespondWithError(w, http.StatusInternalServerError, "Database error while checking user existence")
		return
	}
	// sql.ErrNoRows means user does not exist, which is what we want.

	// Hash the password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password for %s: %v", req.Email, err)
		httputils.RespondWithError(w, http.StatusInternalServerError, "Failed to process password")
		return
	}

	// Create user in the database
	createUserParams := db.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashedPassword,
	}
	createdUserRow, err := h.APIConfig.DB.CreateUser(r.Context(), createUserParams)
	if err != nil {
		// This could be due to a unique constraint violation if another request registered the email
		// between the GetUserByEmail check and this CreateUser call (race condition),
		// or other database errors.
		log.Printf("Error creating user %s in database: %v", req.Email, err)
		// Consider checking for pq.Error unique_violation if using lib/pq directly for more specific error.
		httputils.RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	log.Printf("Successfully registered user with email: %s, ID: %s", createdUserRow.Email, createdUserRow.ID)
	response := toUserResponseFromCreate(createdUserRow)
	httputils.RespondWithJSON(w, http.StatusCreated, response)
}

// LoginHandler handles user login.
// POST /api/v1/auth/login
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.RespondWithError(w, http.StatusMethodNotAllowed, "Only POST method is allowed")
		return
	}

	var req LoginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}
	defer r.Body.Close()

	// Basic Input Validation
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		httputils.RespondWithError(w, http.StatusBadRequest, "Email is required")
		return
	}
	if req.Password == "" {
		httputils.RespondWithError(w, http.StatusBadRequest, "Password is required")
		return
	}

	log.Printf("Attempting to login user with email: %s", req.Email)

	// Fetch user by email
	user, err := h.APIConfig.DB.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Login failed: user with email %s not found", req.Email)
			httputils.RespondWithError(w, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		log.Printf("Database error fetching user %s for login: %v", req.Email, err)
		httputils.RespondWithError(w, http.StatusInternalServerError, "Database error during login")
		return
	}

	// Verify password
	if !auth.CheckPasswordHash(req.Password, user.HashedPassword) {
		log.Printf("Login failed: invalid password for user %s", req.Email)
		httputils.RespondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Login successful, generate JWT
	log.Printf("User %s (ID: %s) credentials verified. Generating JWT.", user.Email, user.ID)

	tokenString, err := auth.GenerateJWT(user.ID, h.APIConfig.JWTSecret, h.APIConfig.JWTExpiration)
	if err != nil {
		log.Printf("Error generating JWT for user %s (ID: %s): %v", user.Email, user.ID, err)
		httputils.RespondWithError(w, http.StatusInternalServerError, "Failed to generate authentication token")
		return
	}

	log.Printf("JWT generated successfully for user %s (ID: %s)", user.Email, user.ID)
	response := LoginResponse{
		Token:  tokenString,
		UserID: user.ID,
		Email:  user.Email,
	}
	httputils.RespondWithJSON(w, http.StatusOK, response)
}
