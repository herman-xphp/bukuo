package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewAccountingPeriod(t *testing.T) {
	companyID := uuid.New()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	period := NewAccountingPeriod(companyID, "January 2026", start, end)

	if period.ID == uuid.Nil {
		t.Error("Expected ID to be generated")
	}
	if period.Status != PeriodStatusOpen {
		t.Error("Expected OPEN status")
	}
	if period.Name != "January 2026" {
		t.Error("Name mismatch")
	}
}

func TestAccountingPeriod_CanPost(t *testing.T) {
	// Open period
	period := &AccountingPeriod{Status: PeriodStatusOpen}
	if err := period.CanPost(); err != nil {
		t.Errorf("Open period should allow posting: %v", err)
	}

	// Closed period
	period.Status = PeriodStatusClosed
	if err := period.CanPost(); err == nil {
		t.Error("Closed period should not allow posting")
	}

	// Locked period
	period.Status = PeriodStatusLocked
	if err := period.CanPost(); err == nil {
		t.Error("Locked period should not allow posting")
	}
}

func TestAccountingPeriod_Close(t *testing.T) {
	period := &AccountingPeriod{Status: PeriodStatusOpen}
	userID := uuid.New()

	err := period.Close(userID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if period.Status != PeriodStatusClosed {
		t.Error("Expected CLOSED status")
	}
	if period.ClosedBy == nil || *period.ClosedBy != userID {
		t.Error("ClosedBy mismatch")
	}

	// Can't close already closed
	err = period.Close(userID)
	if err == nil {
		t.Error("Expected error for already closed period")
	}
}

func TestNewUser(t *testing.T) {
	companyID := uuid.New()

	user, err := NewUser(companyID, "test@example.com", "password123", "Test User", UserRoleOwner)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if user.ID == uuid.Nil {
		t.Error("Expected ID to be generated")
	}
	if user.Email != "test@example.com" {
		t.Error("Email mismatch")
	}
	if user.PasswordHash == "" {
		t.Error("Password should be hashed")
	}
	if user.PasswordHash == "password123" {
		t.Error("Password should not be stored in plain text")
	}
	if !user.IsActive {
		t.Error("Expected user to be active by default")
	}
}

func TestUser_CheckPassword(t *testing.T) {
	user, _ := NewUser(uuid.New(), "test@example.com", "password123", "Test", UserRoleOwner)

	// Correct password
	if err := user.CheckPassword("password123"); err != nil {
		t.Errorf("Password should match: %v", err)
	}

	// Wrong password
	if err := user.CheckPassword("wrongpassword"); err == nil {
		t.Error("Wrong password should fail")
	}

	// Inactive user
	user.IsActive = false
	if err := user.CheckPassword("password123"); err == nil {
		t.Error("Inactive user should fail")
	}
}

func TestUser_Permissions(t *testing.T) {
	// Owner can manage users
	owner := &User{Role: UserRoleOwner}
	if !owner.CanManageUsers() {
		t.Error("Owner should be able to manage users")
	}
	if !owner.CanPostJournals() {
		t.Error("Owner should be able to post journals")
	}

	// Admin can manage users
	admin := &User{Role: UserRoleAdmin}
	if !admin.CanManageUsers() {
		t.Error("Admin should be able to manage users")
	}

	// Accountant cannot manage users but can post
	accountant := &User{Role: UserRoleAccountant}
	if accountant.CanManageUsers() {
		t.Error("Accountant should not be able to manage users")
	}
	if !accountant.CanPostJournals() {
		t.Error("Accountant should be able to post journals")
	}

	// Viewer cannot do either
	viewer := &User{Role: UserRoleViewer}
	if viewer.CanManageUsers() {
		t.Error("Viewer should not be able to manage users")
	}
	if viewer.CanPostJournals() {
		t.Error("Viewer should not be able to post journals")
	}
}

func TestNewCompany(t *testing.T) {
	company := NewCompany("PT Test", "123456789")

	if company.ID == uuid.Nil {
		t.Error("Expected ID to be generated")
	}
	if company.Name != "PT Test" {
		t.Error("Name mismatch")
	}
	if company.TaxID != "123456789" {
		t.Error("TaxID mismatch")
	}
}
