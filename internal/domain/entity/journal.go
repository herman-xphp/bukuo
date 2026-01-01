package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Errors
var (
	ErrJournalNotBalanced = errors.New("debit tidak sama dengan credit")
	ErrMinimumLines       = errors.New("minimal 2 baris jurnal")
	ErrAlreadyPosted      = errors.New("jurnal sudah diposting")
	ErrNotDraft           = errors.New("jurnal bukan draft")
	ErrDualAmount         = errors.New("baris tidak boleh memiliki debit dan credit sekaligus")
	ErrNotPendingApproval = errors.New("jurnal tidak menunggu persetujuan")
	ErrCannotReject       = errors.New("hanya jurnal pending approval yang bisa ditolak")
)

// JournalStatus represents the status of a journal entry
type JournalStatus string

const (
	JournalStatusDraft           JournalStatus = "DRAFT"
	JournalStatusPendingApproval JournalStatus = "PENDING_APPROVAL"
	JournalStatusApproved        JournalStatus = "APPROVED"
	JournalStatusRejected        JournalStatus = "REJECTED"
	JournalStatusPosted          JournalStatus = "POSTED"
	JournalStatusReversed        JournalStatus = "REVERSED"
)

// JournalEntry represents a journal entry header
type JournalEntry struct {
	ID          uuid.UUID     `json:"id"`
	CompanyID   uuid.UUID     `json:"company_id"`
	PeriodID    uuid.UUID     `json:"period_id"`
	EntryNumber string        `json:"entry_number"`
	EntryDate   time.Time     `json:"entry_date"`
	Description string        `json:"description"`
	Status      JournalStatus `json:"status"`
	SourceType  string        `json:"source_type"`
	SourceID    *uuid.UUID    `json:"source_id"`
	CreatedBy   uuid.UUID     `json:"created_by"`
	CreatedAt   time.Time     `json:"created_at"`
	PostedAt    *time.Time    `json:"posted_at"`
	PostedBy    *uuid.UUID    `json:"posted_by"`
	// Approval workflow fields
	ApprovedAt   *time.Time    `json:"approved_at"`
	ApprovedBy   *uuid.UUID    `json:"approved_by"`
	RejectedAt   *time.Time    `json:"rejected_at"`
	RejectedBy   *uuid.UUID    `json:"rejected_by"`
	RejectReason string        `json:"reject_reason,omitempty"`
	Lines        []JournalLine `json:"lines"`
}

// JournalLine represents a journal entry line item
type JournalLine struct {
	ID           uuid.UUID       `json:"id"`
	JournalID    uuid.UUID       `json:"journal_id"`
	LineNumber   int             `json:"line_number"`
	AccountID    uuid.UUID       `json:"account_id"`
	Description  string          `json:"description"`
	DebitAmount  decimal.Decimal `json:"debit_amount"`
	CreditAmount decimal.Decimal `json:"credit_amount"`
}

// NewJournalEntry creates a new journal entry
func NewJournalEntry(companyID, periodID, createdBy uuid.UUID, date time.Time, description string) *JournalEntry {
	return &JournalEntry{
		ID:          uuid.New(),
		CompanyID:   companyID,
		PeriodID:    periodID,
		EntryDate:   date,
		Description: description,
		Status:      JournalStatusDraft,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		Lines:       make([]JournalLine, 0),
	}
}

// AddLine adds a line to the journal entry
func (j *JournalEntry) AddLine(accountID uuid.UUID, description string, debit, credit decimal.Decimal) error {
	if debit.IsPositive() && credit.IsPositive() {
		return ErrDualAmount
	}
	j.Lines = append(j.Lines, JournalLine{
		ID:           uuid.New(),
		JournalID:    j.ID,
		LineNumber:   len(j.Lines) + 1,
		AccountID:    accountID,
		Description:  description,
		DebitAmount:  debit,
		CreditAmount: credit,
	})
	return nil
}

// TotalDebit returns the total debit amount
func (j *JournalEntry) TotalDebit() decimal.Decimal {
	total := decimal.Zero
	for _, line := range j.Lines {
		total = total.Add(line.DebitAmount)
	}
	return total
}

// TotalCredit returns the total credit amount
func (j *JournalEntry) TotalCredit() decimal.Decimal {
	total := decimal.Zero
	for _, line := range j.Lines {
		total = total.Add(line.CreditAmount)
	}
	return total
}

// IsBalanced checks if debit equals credit
func (j *JournalEntry) IsBalanced() bool {
	return j.TotalDebit().Equal(j.TotalCredit())
}

// Validate validates the journal entry
func (j *JournalEntry) Validate() error {
	if len(j.Lines) < 2 {
		return ErrMinimumLines
	}
	if !j.IsBalanced() {
		return ErrJournalNotBalanced
	}
	return nil
}

// Post marks the journal as posted
func (j *JournalEntry) Post(userID uuid.UUID) error {
	// Allow posting from both DRAFT and APPROVED status
	if j.Status != JournalStatusDraft && j.Status != JournalStatusApproved {
		return ErrAlreadyPosted
	}
	if err := j.Validate(); err != nil {
		return err
	}
	now := time.Now()
	j.Status = JournalStatusPosted
	j.PostedAt = &now
	j.PostedBy = &userID
	return nil
}

// CanReverse checks if the journal can be reversed
func (j *JournalEntry) CanReverse() bool {
	return j.Status == JournalStatusPosted
}

// SubmitForApproval submits journal for approval
func (j *JournalEntry) SubmitForApproval() error {
	if j.Status != JournalStatusDraft {
		return ErrNotDraft
	}
	if err := j.Validate(); err != nil {
		return err
	}
	j.Status = JournalStatusPendingApproval
	return nil
}

// Approve approves the journal
func (j *JournalEntry) Approve(userID uuid.UUID) error {
	if j.Status != JournalStatusPendingApproval {
		return ErrNotPendingApproval
	}
	now := time.Now()
	j.Status = JournalStatusApproved
	j.ApprovedAt = &now
	j.ApprovedBy = &userID
	return nil
}

// Reject rejects the journal with reason
func (j *JournalEntry) Reject(userID uuid.UUID, reason string) error {
	if j.Status != JournalStatusPendingApproval {
		return ErrCannotReject
	}
	now := time.Now()
	j.Status = JournalStatusRejected
	j.RejectedAt = &now
	j.RejectedBy = &userID
	j.RejectReason = reason
	return nil
}

// NeedsApproval checks if journal exceeds threshold
func (j *JournalEntry) NeedsApproval(threshold decimal.Decimal) bool {
	return j.TotalDebit().GreaterThan(threshold)
}
