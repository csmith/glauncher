package calc

import (
	"math"
	"testing"
)

func TestEvaluate(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"2 + 3", 5},
		{"10 - 4", 6},
		{"3 * 7", 21},
		{"20 / 4", 5},
		{"10 % 3", 1},
		{"2 ^ 10", 1024},
		{"2 + 3 * 4", 14},
		{"(2 + 3) * 4", 20},
		{"-5 + 3", -2},
		{"2 ^ 3 ^ 2", 512},
		{"10 / 0", 0},
		{"3.14 * 2", 6.28},
		{"1 + 2 * (3 + 4) ^ 2", 99},
		{"  42  ", 42},
		{"0.5 * 0.5", 0.25},
		{"100 % 7", 2},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if tt.input == "10 / 0" {
				_, err := evaluate(tt.input)
				if err == nil {
					t.Error("expected error for division by zero")
				}
				return
			}
			got, err := evaluate(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if math.Abs(got-tt.expected) > 1e-9 {
				t.Errorf("evaluate(%q) = %g, want %g", tt.input, got, tt.expected)
			}
		})
	}
}

func TestEvaluateErrors(t *testing.T) {
	tests := []struct {
		input string
	}{
		{""},
		{"abc"},
		{"2 + "},
		{"(2 + 3"},
		{"2 + 3)"},
		{"10 % 0"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := evaluate(tt.input)
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestLooksLikeExpression(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"2 + 3", true},
		{"hello", false},
		{"42", false},
		{"3 * 7", true},
		{"(1 + 2)", true},
		{"10 % 3", true},
		{"2 ^ 8", true},
		{"", false},
		{"firefox", false},
		{"3.14 / 2", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := looksLikeExpression(tt.input)
			if got != tt.want {
				t.Errorf("looksLikeExpression(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatResult(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{42, "42"},
		{3.14, "3.14"},
		{0.25, "0.25"},
		{-7, "-7"},
		{1024, "1024"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := formatResult(tt.input)
			if got != tt.expected {
				t.Errorf("formatResult(%g) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
