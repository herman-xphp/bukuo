package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type JournalStatus string

const (
	JournalStatusDraft    JournalStatus = "DRAFT"
	JournalStatusPosted   JournalStatus = "POSTED"
	JournalStatusReversed JournalStatus = "REVERSED"
)

// JounalEntry - Header jurnal
type JounalEntry struct {
	ID          uuid.UUID     `json:"id"`
	CompanyID   uuid.UUID     `json:"company_id"`
	EntryNumber string        `json:"entry_number"` // JE-2026-0001
	EntryDate   time.Time     `json:"entry_date"`
	PeriodID    uuid.UUID     `json:"period_id"`
	Description string        `json:"description"`
	Status      JournalStatus `json:"status"`
	CreatedBy   uuid.UUID     `json:"created_by"`
	CreatedAt   time.Time     `json:"created_at"`
	Lines       []JournalLine `json:"lines"`
}

// JournalLine - Baris jurnal (debit/credit)
type JournalLine struct {
	ID           uuid.UUID       `json:"id"`
	JounalID     uuid.UUID       `json:"journal_id"`
	LineNumber   int             `json:"line_number"`
	AccountID    uuid.UUID       `json:"account_id"`
	Description  string          `json:"description"`
	DebitAmount  decimal.Decimal `json:"debit_amount"`
	CreditAmount decimal.Decimal `json:"credit_amount"`
}

// TotalDebit - Jumlah semua debit
func (j *JounalEntry) TotalDebit() decimal.Decimal {
	total := decimal.Zero
	for _, line := range j.Lines {
		total = total.Add(line.DebitAmount)
	}
	return total
}

// TotalCredit - Jumlah semua credit
func (j *JounalEntry) TotalCredit() decimal.Decimal {
	total := decimal.Zero
	for _, line := range j.Lines {
		total = total.Add(line.CreditAmount)
	}
	return total
}

// IsBalanced - WAJIB: Debit = Credit
func (j *JounalEntry) IsBalanced() bool {
	return j.TotalDebit().Equal(j.TotalCredit())
}
