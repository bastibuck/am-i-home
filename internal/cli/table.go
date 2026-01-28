package cli

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

// PrintStructTable prints a table for a slice of structs (or pointers to structs).
// Column order is determined by the struct's field declaration order (exported fields only).
// If headers is non-nil, a header row is printed followed by a separator line.
// When headers is provided, its length must match the number of exported fields.
// If headers is nil, no header row is printed.
func PrintStructTable(w io.Writer, items interface{}, headers []string) error {
	v := reflect.ValueOf(items)
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("items must be a slice")
	}

	// get struct type from slice element type
	elemType := v.Type().Elem()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}
	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("slice elements must be structs or pointers to structs")
	}

	// collect exported field indices
	var fieldIndices []int
	for i := 0; i < elemType.NumField(); i++ {
		if elemType.Field(i).PkgPath == "" { // exported
			fieldIndices = append(fieldIndices, i)
		}
	}

	if len(fieldIndices) == 0 {
		return fmt.Errorf("no exported fields found in struct")
	}

	colCount := len(fieldIndices)

	// validate headers length if provided
	if headers != nil && len(headers) != colCount {
		return fmt.Errorf("headers length (%d) must match number of exported fields (%d)", len(headers), colCount)
	}

	// initialize column widths from headers (if present)
	colWidths := make([]int, colCount)
	for i, h := range headers {
		colWidths[i] = len(h)
	}

	// gather rows and compute widths
	rows := make([][]string, 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		ev := v.Index(i)
		if ev.Kind() == reflect.Ptr {
			if ev.IsNil() {
				row := make([]string, colCount)
				rows = append(rows, row)
				continue
			}
			ev = ev.Elem()
		}

		row := make([]string, colCount)
		for j, fieldIdx := range fieldIndices {
			fv := ev.Field(fieldIdx)
			var s string
			if fv.CanInterface() {
				s = fmt.Sprint(fv.Interface())
			}
			row[j] = s
			if len(s) > colWidths[j] {
				colWidths[j] = len(s)
			}
		}
		rows = append(rows, row)
	}

	// compute total width for separator
	totalWidth := 0
	for i := 0; i < colCount; i++ {
		if i > 0 {
			totalWidth += 3 // " | "
		}
		totalWidth += colWidths[i]
	}

	// print header if provided
	if headers != nil {
		for i, h := range headers {
			if i > 0 {
				fmt.Fprint(w, " | ")
			}
			fmt.Fprint(w, padRight(h, colWidths[i]))
		}
		fmt.Fprintln(w)
		fmt.Fprintln(w, strings.Repeat("-", totalWidth))
	}

	// print rows
	for _, row := range rows {
		for i, col := range row {
			if i > 0 {
				fmt.Fprint(w, " | ")
			}
			fmt.Fprint(w, padRight(col, colWidths[i]))
		}
		fmt.Fprintln(w)
	}

	return nil
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
