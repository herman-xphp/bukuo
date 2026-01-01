package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Errors
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotActive      = errors.New("user is not active")
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
	ID           uuid.UUID  `json:"id"`
	CompanyID    uuid.UUID  `json:"company_id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"` // Never expose in JSON
	Name         string     `json:"name"`
	Role         UserRole   `json:"role"`
	IsActive     bool       `json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// NewUser creates a new user with hashed password
func NewUser(companyID uuid.UUID, email, password, name string, role UserRole) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:           uuid.New(),
		CompanyID:    companyID,
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		Role:         role,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

// CheckPassword verifies the password
func (u *User) CheckPassword(password string) error {
	if !u.IsActive {
		return ErrUserNotActive
	}
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if err != nil {
		return ErrInvalidCredentials
	}
	return nil
}

// UpdateLastLogin updates the last login timestamp
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
}

// CanManageUsers checks if user can manage other users
func (u *User) CanManageUsers() bool {
	return u.Role == UserRoleOwner || u.Role == UserRoleAdmin
}

// CanPostJournals checks if user can post journals
func (u *User) CanPostJournals() bool {
	return u.Role == UserRoleOwner || u.Role == UserRoleAdmin || u.Role == UserRoleAccountant
}
