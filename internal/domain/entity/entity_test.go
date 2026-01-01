package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestNewAccount(t *testing.T) {
	companyID := uuid.New()
	acc := NewAccount(companyID, "1-1100", "Kas", AccountTypeAsset)

	if acc.ID == uuid.Nil {
		t.Error("Expected ID to be generated")
	}
	if acc.CompanyID != companyID {
		t.Error("CompanyID mismatch")
	}
	if acc.Code != "1-1100" {
		t.Errorf("Expected code 1-1100, got %s", acc.Code)
	}
	if acc.Type != AccountTypeAsset {
		t.Error("Expected ASSET type")
	}
	if !acc.IsActive {
		t.Error("Expected account to be active by default")
	}
	if !acc.IsPostable {
		t.Error("Expected account to be postable by default")
	}
}

func TestAccount_NormalBalance(t *testing.T) {
	tests := []struct {
		accountType AccountType
		expected    string
	}{
		{AccountTypeAsset, "DEBIT"},
		{AccountTypeExpense, "DEBIT"},
		{AccountTypeLiability, "CREDIT"},
		{AccountTypeEquity, "CREDIT"},
		{AccountTypeRevenue, "CREDIT"},
	}

	for _, tt := range tests {
		acc := &Account{Type: tt.accountType}
		result := acc.NormalBalance()
		if result != tt.expected {
			t.Errorf("For %s expected %s, got %s", tt.accountType, tt.expected, result)
		}
	}
}

func TestAccount_CanPost(t *testing.T) {
	// Postable account - should succeed
	acc := &Account{IsPostable: true}
	if err := acc.CanPost(); err != nil {
		t.Errorf("Expected no error for postable account, got %v", err)
	}

	// Not postable - should fail
	acc = &Account{IsPostable: false}
	if err := acc.CanPost(); err == nil {
		t.Error("Expected error for non-postable account")
	}
}

func TestNewJournalEntry(t *testing.T) {
	companyID := uuid.New()
	periodID := uuid.New()
	userID := uuid.New()
	date := time.Now()

	journal := NewJournalEntry(companyID, periodID, userID, date, "Test entry")

	if journal.ID == uuid.Nil {
		t.Error("Expected ID to be generated")
	}
	if journal.Status != JournalStatusDraft {
		t.Error("Expected DRAFT status")
	}
	if journal.Description != "Test entry" {
		t.Error("Description mismatch")
	}
}

func TestJournalEntry_AddLine(t *testing.T) {
	journal := &JournalEntry{}

	// Valid debit line
	err := journal.AddLine(uuid.New(), "Test", decimal.NewFromInt(1000), decimal.Zero)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(journal.Lines) != 1 {
		t.Error("Expected 1 line")
	}
	if journal.Lines[0].LineNumber != 1 {
		t.Error("Expected line number 1")
	}

	// Valid credit line
	err = journal.AddLine(uuid.New(), "Test2", decimal.Zero, decimal.NewFromInt(1000))
	if err != nil {
		t.Errorf("Unexpected error for credit line: %v", err)
	}
	if len(journal.Lines) != 2 {
		t.Error("Expected 2 lines")
	}
}

func TestJournalEntry_IsBalanced(t *testing.T) {
	journal := &JournalEntry{
		Lines: []JournalLine{
			{DebitAmount: decimal.NewFromInt(1000), CreditAmount: decimal.Zero},
			{DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromInt(1000)},
		},
	}

	if !journal.IsBalanced() {
		t.Error("Expected journal to be balanced")
	}

	// Unbalanced
	journal.Lines[1].CreditAmount = decimal.NewFromInt(500)
	if journal.IsBalanced() {
		t.Error("Expected journal to be unbalanced")
	}
}

func TestJournalEntry_Validate(t *testing.T) {
	journal := &JournalEntry{
		Lines: []JournalLine{
			{DebitAmount: decimal.NewFromInt(1000), CreditAmount: decimal.Zero},
			{DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromInt(1000)},
		},
	}

	if err := journal.Validate(); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Less than 2 lines
	journal.Lines = journal.Lines[:1]
	if err := journal.Validate(); err == nil {
		t.Error("Expected error for less than 2 lines")
	}
}

func TestJournalEntry_Post(t *testing.T) {
	// Create a valid journal with 2 balanced lines
	journal := &JournalEntry{
		Status: JournalStatusDraft,
		Lines: []JournalLine{
			{DebitAmount: decimal.NewFromInt(1000), CreditAmount: decimal.Zero},
			{DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromInt(1000)},
		},
	}
	userID := uuid.New()

	err := journal.Post(userID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if journal.Status != JournalStatusPosted {
		t.Error("Expected POSTED status")
	}
	if journal.PostedBy == nil || *journal.PostedBy != userID {
		t.Error("PostedBy mismatch")
	}
	if journal.PostedAt == nil {
		t.Error("PostedAt should be set")
	}

	// Can't post already posted
	err = journal.Post(userID)
	if err == nil {
		t.Error("Expected error for already posted journal")
	}
}

func TestJournalEntry_CanReverse(t *testing.T) {
	// Posted journal can be reversed
	journal := &JournalEntry{Status: JournalStatusPosted}
	if !journal.CanReverse() {
		t.Error("Posted journal should be reversible")
	}

	// Draft cannot
	journal.Status = JournalStatusDraft
	if journal.CanReverse() {
		t.Error("Draft journal should not be reversible")
	}

	// Already reversed cannot
	journal.Status = JournalStatusReversed
	if journal.CanReverse() {
		t.Error("Reversed journal should not be reversible")
	}
}
