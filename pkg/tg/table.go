package tg

import (
	"fmt"
	"strings"
)

const (
	gutter = 3
)

type Table struct {
	columns      []string
	columnWidths []int
	rows         [][]string
	HideHeading  bool
}

// NewTable creates a new table object
func NewTable(columns ...string) *Table {
	columnWidths := make([]int, len(columns))
	for idx, col := range columns {
		columnWidths[idx] = len(col)
	}

	return &Table{
		columns:      columns,
		columnWidths: columnWidths,
	}
}

// Append adds a row of values to the table
func (t *Table) Append(values ...string) {
	if len(values) != len(t.columns) {
		panic("term.Append: too many values to append")
	}

	t.rows = append(t.rows, values)
	for idx, v := range values {
		if len(v) > t.columnWidths[idx] {
			t.columnWidths[idx] = len(v)
		}
	}
}

// Render prints the table to the stdout
func (t *Table) Render() {
	values := make([]string, len(t.columns))
	for idx, col := range t.columns {
		values[idx] = fmt.Sprintf("%s%s", strings.ToUpper(col), strings.Repeat(" ", t.columnWidths[idx]-len(col)))
	}
	if !t.HideHeading {
		Println(Bold, Underline, White, strings.Join(values, strings.Repeat(" ", gutter)), Reset)
	}
	for _, row := range t.rows {
		for idx := range t.columns {
			col := row[idx]
			values[idx] = fmt.Sprintf("%s%s", col, strings.Repeat(" ", t.columnWidths[idx]-len(col)))
		}
		Println(White, strings.Join(values, strings.Repeat(" ", gutter)), Reset)
	}
}
