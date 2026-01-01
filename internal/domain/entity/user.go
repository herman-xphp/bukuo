package entity

import (
	"errors"
	"regexp"
	"time"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Errors
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotActive      = errors.New("user is not active")
	ErrWeakPassword       = errors.New("password must be at least 8 characters with uppercase, lowercase, and number")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrAccountLocked      = errors.New("account is locked due to too many failed attempts")
)

// UserRole represents the role of a user
type UserRole string

const (
	UserRoleOwner      UserRole = "OWNER"
	UserRoleAdmin      UserRole = "ADMIN"
	UserRoleAccountant UserRole = "ACCOUNTANT"
	UserRoleViewer     UserRole = "VIEWER"
)

// User represents a user in the system
type User struct {
	ID             uuid.UUID  `json:"id"`
	CompanyID      uuid.UUID  `json:"company_id"`
	Email          string     `json:"email"`
	PasswordHash   string     `json:"-"` // Never expose in JSON
	Name           string     `json:"name"`
	Role           UserRole   `json:"role"`
	IsActive       bool       `json:"is_active"`
	FailedAttempts int        `json:"failed_attempts"` // For brute force protection
	LockedUntil    *time.Time `json:"locked_until"`    // Account lockout
	LastLoginAt    *time.Time `json:"last_login_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// ValidatePassword checks password strength
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrWeakPassword
	}

	var hasUpper, hasLower, hasNumber bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsNumber(c):
			hasNumber = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber {
		return ErrWeakPassword
	}

	return nil
}

// ValidateEmail checks email format
func ValidateEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

// NewUser creates a new user with hashed password
func NewUser(companyID uuid.UUID, email, password, name string, role UserRole) (*User, error) {
	// Validate email
	if err := ValidateEmail(email); err != nil {
		return nil, err
	}

	// Validate password strength
	if err := ValidatePassword(password); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:             uuid.New(),
		CompanyID:      companyID,
		Email:          email,
		PasswordHash:   string(hash),
		Name:           name,
		Role:           role,
		IsActive:       true,
		FailedAttempts: 0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil
}

// IsLocked checks if account is currently locked
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.LockedUntil)
}

// CheckPassword verifies the password
func (u *User) CheckPassword(password string) error {
	if !u.IsActive {
		return ErrUserNotActive
	}

	if u.IsLocked() {
		return ErrAccountLocked
	}

	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if err != nil {
		return ErrInvalidCredentials
	}
	return nil
}

// RecordFailedAttempt increments failed attempts and locks if exceeded
func (u *User) RecordFailedAttempt(maxAttempts int, lockoutDuration time.Duration) {
	u.FailedAttempts++
	if u.FailedAttempts >= maxAttempts {
		lockUntil := time.Now().Add(lockoutDuration)
		u.LockedUntil = &lockUntil
	}
	u.UpdatedAt = time.Now()
}

// ResetFailedAttempts resets the failed attempt counter
func (u *User) ResetFailedAttempts() {
	u.FailedAttempts = 0
	u.LockedUntil = nil
	u.UpdatedAt = time.Now()
}

// UpdateLastLogin updates the last login timestamp
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.ResetFailedAttempts()
}

// CanManageUsers checks if user can manage other users
func (u *User) CanManageUsers() bool {
	return u.Role == UserRoleOwner || u.Role == UserRoleAdmin
}

// CanPostJournals checks if user can post journals
func (u *User) CanPostJournals() bool {
	return u.Role == UserRoleOwner || u.Role == UserRoleAdmin || u.Role == UserRoleAccountant
}
