package entity

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (u *User) SetPin(pin string) error {
	if len(pin) < 4 || len(pin) > 6 {
		return errors.New("pin must be 4-6 digits")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PinHash = string(hash)
	u.UpdatedAt = time.Now()
	return nil
}

// CheckPin checks if pin matches hash
func (u *User) CheckPin(pin string) error {
	if u.PinHash == "" {
		return errors.New("pin not set")
	}
	// We don't check for lock/active status here, assuming calling service does that or it's just a screen unlock
	// But logically, if account is inactive, PIN shouldn't work either.
	if !u.IsActive {
		return ErrUserNotActive
	}

	// Lock logic? Maybe separate lock for PIN attempts?
	// For now, simple check.
	return bcrypt.CompareHashAndPassword([]byte(u.PinHash), []byte(pin))
}
