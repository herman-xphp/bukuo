package validator

import (
	"errors"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/domain"
)

var (
	ErrNotBalanced  = errors.New("debit tidak sama dengan credit")
	ErrPeriodClosed = errors.New("periode sudah ditutup")
	ErrNotPostable  = errors.New("akun tidak bisa diposting")
	ErrMinLines     = errors.New("minimal 2 baris jurnal")
	ErrDualAmount   = errors.New("tidak boleh debit dan credit bersamaan")
)

type JounalValidator struct{}

func NewJounalValidator() *JounalValidator {
	return &JounalValidator{}
}

func (v *JounalValidator) Validate(
	journal *domain.JounalEntry,
	period *domain.AccountPeriod,
	accounts map[uuid.UUID]*domain.Account,
) error {
	// Rule 1: Minimal 2 baris
	if len(journal.Lines) < 2 {
		return ErrMinLines
	}

	// Rule 2: Debit = Credit (ZERO TOLERANCE!)
	if !journal.IsBalanced() {
		return ErrNotBalanced
	}

	// Rule 3: Period harus OPEN
	if !period.IsOpen() {
		return ErrPeriodClosed
	}

	// Rule 4: Validasi setiap baris
	for _, line := range journal.Lines {
		if line.DebitAmount.IsPositive() && line.CreditAmount.IsPositive() {
			return ErrDualAmount
		}

		acc, ok := accounts[line.AccountID]
		if !ok || !acc.IsPostable {
			return ErrNotPostable
		}
	}

	return nil
}
