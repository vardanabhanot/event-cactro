package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"eventbooking/config"
	"eventbooking/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	mw "eventbooking/middleware"
)

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` // "organizer" or "customer"
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register handles POST /api/auth/register
func Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Role = strings.TrimSpace(req.Role)

	if req.Name == "" || req.Email == "" || req.Password == "" {
		jsonError(w, "name, email, and password are required", http.StatusBadRequest)
		return
	}
	if req.Role != "organizer" && req.Role != "customer" {
		jsonError(w, "role must be 'organizer' or 'customer'", http.StatusBadRequest)
		return
	}

	// Check for duplicate email
	existing, err := models.FindUserByEmail(req.Email)
	if err != nil {
		jsonError(w, "database error", http.StatusInternalServerError)
		return
	}
	if existing != nil {
		jsonError(w, "email already registered", http.StatusConflict)
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		jsonError(w, "failed to hash password", http.StatusInternalServerError)
		return
	}

	user, err := models.CreateUser(req.Name, req.Email, string(hashed), req.Role)
	if err != nil {
		jsonError(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	jsonOK(w, user, http.StatusCreated)
}

// Login handles POST /api/auth/login
func Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	user, err := models.FindUserByEmail(req.Email)
	if err != nil {
		jsonError(w, "database error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		jsonError(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		jsonError(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	claims := &mw.Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(config.App.JWTSecret))
	if err != nil {
		jsonError(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]interface{}{
		"token": signed,
		"user":  user,
	}, http.StatusOK)
}
