package cli

import (
	"bytes"
	"strings"
	"testing"
)

// splitLines splits output into lines, removing only the trailing newline
// (preserving trailing spaces on each line).
func splitLines(s string) []string {
	s = strings.TrimSuffix(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func TestPrintStructTable(t *testing.T) {
	type Device struct {
		Name   string
		IP     string
		Active bool
	}

	t.Run("prints table with headers", func(t *testing.T) {
		var buf bytes.Buffer
		devices := []Device{
			{Name: "Phone", IP: "192.168.1.10", Active: true},
			{Name: "Laptop", IP: "192.168.1.20", Active: false},
		}

		err := PrintStructTable(&buf, devices, []string{"Name", "IP Address", "Active"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		lines := splitLines(buf.String())

		// Verify row count: header + separator + 2 data rows
		if len(lines) != 4 {
			t.Fatalf("expected 4 lines (header, separator, 2 rows), got %d:\n%s", len(lines), buf.String())
		}

		// Verify complete header row with proper spacing
		// "IP Address" (10 chars) is longest in column 2, so "192.168.1.10" (12 chars) determines width
		expectedHeader := "Name   | IP Address   | Active"
		if lines[0] != expectedHeader {
			t.Errorf("expected header row:\n%q\ngot:\n%q", expectedHeader, lines[0])
		}

		// Verify separator matches total width
		expectedSeparator := strings.Repeat("-", len(expectedHeader))
		if lines[1] != expectedSeparator {
			t.Errorf("expected separator:\n%q\ngot:\n%q", expectedSeparator, lines[1])
		}

		// Verify data rows (last column is also padded to "Active" width of 6)
		expectedRow1 := "Phone  | 192.168.1.10 | true  "
		expectedRow2 := "Laptop | 192.168.1.20 | false "
		if lines[2] != expectedRow1 {
			t.Errorf("expected row 1:\n%q\ngot:\n%q", expectedRow1, lines[2])
		}
		if lines[3] != expectedRow2 {
			t.Errorf("expected row 2:\n%q\ngot:\n%q", expectedRow2, lines[3])
		}
	})

	t.Run("prints table without headers", func(t *testing.T) {
		var buf bytes.Buffer
		devices := []Device{
			{Name: "Phone", IP: "192.168.1.10", Active: true},
		}

		err := PrintStructTable(&buf, devices, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		lines := splitLines(buf.String())

		// Verify row count: only 1 data row, no header or separator
		if len(lines) != 1 {
			t.Fatalf("expected 1 line (data row only), got %d:\n%s", len(lines), buf.String())
		}

		// Verify no separator exists
		if strings.Contains(buf.String(), "---") {
			t.Errorf("expected no separator line without headers, got:\n%s", buf.String())
		}

		// Verify exact row content
		expectedRow := "Phone | 192.168.1.10 | true"
		if lines[0] != expectedRow {
			t.Errorf("expected row:\n%q\ngot:\n%q", expectedRow, lines[0])
		}
	})

	t.Run("handles pointer slice", func(t *testing.T) {
		var buf bytes.Buffer
		devices := []*Device{
			{Name: "Phone", IP: "192.168.1.10", Active: true},
			nil,
			{Name: "Laptop", IP: "192.168.1.20", Active: false},
		}

		err := PrintStructTable(&buf, devices, []string{"Name", "IP", "Active"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		lines := splitLines(buf.String())

		// Verify row count: header + separator + 3 data rows (including nil)
		if len(lines) != 5 {
			t.Fatalf("expected 5 lines (header, separator, 3 rows), got %d:\n%s", len(lines), buf.String())
		}

		// Verify header row
		expectedHeader := "Name   | IP           | Active"
		if lines[0] != expectedHeader {
			t.Errorf("expected header row:\n%q\ngot:\n%q", expectedHeader, lines[0])
		}

		// Verify separator
		expectedSeparator := strings.Repeat("-", len(expectedHeader))
		if lines[1] != expectedSeparator {
			t.Errorf("expected separator:\n%q\ngot:\n%q", expectedSeparator, lines[1])
		}

		// Verify data rows (nil row should have empty values, last column padded)
		expectedRow1 := "Phone  | 192.168.1.10 | true  "
		expectedRow2 := "       |              |       " // nil pointer row
		expectedRow3 := "Laptop | 192.168.1.20 | false "
		if lines[2] != expectedRow1 {
			t.Errorf("expected row 1:\n%q\ngot:\n%q", expectedRow1, lines[2])
		}
		if lines[3] != expectedRow2 {
			t.Errorf("expected row 2 (nil):\n%q\ngot:\n%q", expectedRow2, lines[3])
		}
		if lines[4] != expectedRow3 {
			t.Errorf("expected row 3:\n%q\ngot:\n%q", expectedRow3, lines[4])
		}
	})

	t.Run("returns error for non-slice", func(t *testing.T) {
		var buf bytes.Buffer
		err := PrintStructTable(&buf, "not a slice", nil)
		if err == nil {
			t.Fatal("expected error for non-slice input")
		}
		if !strings.Contains(err.Error(), "must be a slice") {
			t.Errorf("expected 'must be a slice' error, got: %v", err)
		}
	})

	t.Run("returns error for non-struct elements", func(t *testing.T) {
		var buf bytes.Buffer
		err := PrintStructTable(&buf, []string{"a", "b"}, nil)
		if err == nil {
			t.Fatal("expected error for non-struct elements")
		}
		if !strings.Contains(err.Error(), "must be structs") {
			t.Errorf("expected 'must be structs' error, got: %v", err)
		}
	})

	t.Run("returns error for header length mismatch", func(t *testing.T) {
		var buf bytes.Buffer
		devices := []Device{{Name: "Phone", IP: "192.168.1.10", Active: true}}

		err := PrintStructTable(&buf, devices, []string{"Name", "IP"}) // missing one header
		if err == nil {
			t.Fatal("expected error for header length mismatch")
		}
		if !strings.Contains(err.Error(), "headers length") {
			t.Errorf("expected 'headers length' error, got: %v", err)
		}
	})

	t.Run("returns error for struct with no exported fields", func(t *testing.T) {
		type unexported struct {
			name string
			ip   string
		}
		var buf bytes.Buffer
		items := []unexported{{name: "test", ip: "1.2.3.4"}}

		err := PrintStructTable(&buf, items, nil)
		if err == nil {
			t.Fatal("expected error for no exported fields")
		}
		if !strings.Contains(err.Error(), "no exported fields") {
			t.Errorf("expected 'no exported fields' error, got: %v", err)
		}
	})

	t.Run("handles empty slice", func(t *testing.T) {
		var buf bytes.Buffer
		devices := []Device{}

		err := PrintStructTable(&buf, devices, []string{"Name", "IP", "Active"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		lines := splitLines(buf.String())

		// Verify row count: header + separator only
		if len(lines) != 2 {
			t.Fatalf("expected 2 lines (header, separator), got %d:\n%s", len(lines), buf.String())
		}

		// Verify header row (widths based on header text since no data)
		expectedHeader := "Name | IP | Active"
		if lines[0] != expectedHeader {
			t.Errorf("expected header row:\n%q\ngot:\n%q", expectedHeader, lines[0])
		}

		// Verify separator
		expectedSeparator := strings.Repeat("-", len(expectedHeader))
		if lines[1] != expectedSeparator {
			t.Errorf("expected separator:\n%q\ngot:\n%q", expectedSeparator, lines[1])
		}
	})

	t.Run("pads columns correctly for long values", func(t *testing.T) {
		var buf bytes.Buffer
		devices := []Device{
			{Name: "A", IP: "1.1.1.1", Active: true},
			{Name: "LongName", IP: "192.168.100.200", Active: false},
		}

		err := PrintStructTable(&buf, devices, []string{"N", "IP", "Active"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		lines := splitLines(buf.String())

		// Verify row count
		if len(lines) != 4 {
			t.Fatalf("expected 4 lines (header, separator, 2 rows), got %d:\n%s", len(lines), buf.String())
		}

		// Column widths: "LongName" (8) > "N" (1), "192.168.100.200" (15) > "IP" (2), "Active" (6) > "false" (5)
		expectedHeader := "N        | IP              | Active"
		if lines[0] != expectedHeader {
			t.Errorf("expected header row:\n%q\ngot:\n%q", expectedHeader, lines[0])
		}

		expectedSeparator := strings.Repeat("-", len(expectedHeader))
		if lines[1] != expectedSeparator {
			t.Errorf("expected separator:\n%q\ngot:\n%q", expectedSeparator, lines[1])
		}

		// Data rows (last column also padded to "Active" width of 6)
		expectedRow1 := "A        | 1.1.1.1         | true  "
		expectedRow2 := "LongName | 192.168.100.200 | false "
		if lines[2] != expectedRow1 {
			t.Errorf("expected row 1:\n%q\ngot:\n%q", expectedRow1, lines[2])
		}
		if lines[3] != expectedRow2 {
			t.Errorf("expected row 2:\n%q\ngot:\n%q", expectedRow2, lines[3])
		}
	})
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		input    string
		width    int
		expected string
	}{
		{"abc", 5, "abc  "},
		{"abc", 3, "abc"},
		{"abc", 2, "abc"},
		{"", 3, "   "},
	}

	for _, tt := range tests {
		result := padRight(tt.input, tt.width)
		if result != tt.expected {
			t.Errorf("padRight(%q, %d) = %q, want %q", tt.input, tt.width, result, tt.expected)
		}
	}
}
