package cli

import (
	"bytes"
	"strings"
	"testing"
)

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

		output := buf.String()
		if !strings.Contains(output, "Name") || !strings.Contains(output, "IP Address") || !strings.Contains(output, "Active") {
			t.Errorf("expected headers in output, got:\n%s", output)
		}
		if !strings.Contains(output, "Phone") || !strings.Contains(output, "192.168.1.10") {
			t.Errorf("expected row data in output, got:\n%s", output)
		}
		if !strings.Contains(output, "---") {
			t.Errorf("expected separator line in output, got:\n%s", output)
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

		output := buf.String()
		if strings.Contains(output, "---") {
			t.Errorf("expected no separator line without headers, got:\n%s", output)
		}
		if !strings.Contains(output, "Phone") {
			t.Errorf("expected row data in output, got:\n%s", output)
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

		output := buf.String()
		if !strings.Contains(output, "Phone") || !strings.Contains(output, "Laptop") {
			t.Errorf("expected row data in output, got:\n%s", output)
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

		output := buf.String()
		// Should still print headers
		if !strings.Contains(output, "Name") {
			t.Errorf("expected headers in output for empty slice, got:\n%s", output)
		}
	})

	t.Run("pads columns correctly", func(t *testing.T) {
		var buf bytes.Buffer
		devices := []Device{
			{Name: "A", IP: "1.1.1.1", Active: true},
			{Name: "LongName", IP: "192.168.100.200", Active: false},
		}

		err := PrintStructTable(&buf, devices, []string{"N", "IP", "Active"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) < 4 {
			t.Fatalf("expected at least 4 lines (header, separator, 2 rows), got %d", len(lines))
		}

		// All data lines should have consistent column separators
		for _, line := range lines {
			if strings.Contains(line, "---") {
				continue
			}
			if strings.Count(line, " | ") != 2 {
				t.Errorf("expected 2 column separators in line, got: %s", line)
			}
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
