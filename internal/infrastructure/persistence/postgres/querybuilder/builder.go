package querybuilder

import (
	"fmt"
	"strings"
)

// Builder helps construct dynamic SQL WHERE clauses with parameterized arguments.
// It automatically tracks argument positions for PostgreSQL-style $1, $2, etc.
type Builder struct {
	conditions []string
	args       []interface{}
	argPos     int
}

// New creates a new QueryBuilder starting at argument position 1.
func New() *Builder {
	return &Builder{
		conditions: make([]string, 0),
		args:       make([]interface{}, 0),
		argPos:     1,
	}
}

// AddCondition adds a condition with a single argument.
// The format should use %d as placeholder for the argument position.
// Example: qb.AddCondition("company_id = $%d", companyID)
func (b *Builder) AddCondition(format string, arg interface{}) *Builder {
	condition := fmt.Sprintf(format, b.argPos)
	b.conditions = append(b.conditions, condition)
	b.args = append(b.args, arg)
	b.argPos++
	return b
}

// AddConditionMultiArg adds a condition with multiple uses of the same argument.
// Useful for ILIKE conditions: qb.AddConditionMultiArg("(name ILIKE $%d OR code ILIKE $%d)", 2, "%search%")
func (b *Builder) AddConditionMultiArg(format string, argCount int, arg interface{}) *Builder {
	positions := make([]interface{}, argCount)
	for i := 0; i < argCount; i++ {
		positions[i] = b.argPos
	}
	condition := fmt.Sprintf(format, positions...)
	b.conditions = append(b.conditions, condition)
	b.args = append(b.args, arg)
	b.argPos++
	return b
}

// AddConditionIf adds a condition only if the condition is true.
// Useful for optional filters.
func (b *Builder) AddConditionIf(cond bool, format string, arg interface{}) *Builder {
	if cond {
		return b.AddCondition(format, arg)
	}
	return b
}

// AddSearch adds a search condition across multiple columns.
// Example: qb.AddSearch("value", "name", "code")
func (b *Builder) AddSearch(search string, columns ...string) *Builder {
	if search == "" || len(columns) == 0 {
		return b
	}

	var parts []string
	for _, col := range columns {
		parts = append(parts, fmt.Sprintf("%s ILIKE $%d", col, b.argPos))
	}
	condition := "(" + strings.Join(parts, " OR ") + ")"
	b.conditions = append(b.conditions, condition)
	b.args = append(b.args, "%"+search+"%")
	b.argPos++
	return b
}

// WhereClause returns the WHERE clause (without "WHERE" keyword).
// Returns empty string if no conditions.
func (b *Builder) WhereClause() string {
	if len(b.conditions) == 0 {
		return ""
	}
	return strings.Join(b.conditions, " AND ")
}

// Args returns all accumulated arguments.
func (b *Builder) Args() []interface{} {
	return b.args
}

// NextPos returns the next argument position.
// Useful for adding LIMIT and OFFSET.
func (b *Builder) NextPos() int {
	return b.argPos
}

// AddLimitOffset adds limit and offset arguments and returns their positions.
func (b *Builder) AddLimitOffset(limit, offset int) (limitPos, offsetPos int) {
	limitPos = b.argPos
	b.args = append(b.args, limit)
	b.argPos++

	offsetPos = b.argPos
	b.args = append(b.args, offset)
	b.argPos++

	return limitPos, offsetPos
}
