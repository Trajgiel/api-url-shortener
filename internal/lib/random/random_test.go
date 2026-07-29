package random

import (
	"strings"
	"testing"
)

func TestNewRandomString(t *testing.T) {
	cases := []struct {
		name string
		size int
	}{
		{name: "size = 1", size: 1},
		{name: "size = 5", size: 5},
		{name: "size = 10", size: 10},
		{name: "size = 20", size: 20},
		{name: "size = 30", size: 30},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			str1 := NewRandomString(tc.size)
			str2 := NewRandomString(tc.size)

			if len(str1) != tc.size {
				t.Errorf("expected length %d, got %d", tc.size, len(str1))
			}

			if str1 == str2 {
				t.Errorf("expected different strings, got equal: %q == %q", str1, str2)
			}

			const allowedChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

			for _, c := range str1 {
				if !strings.ContainsRune(allowedChars, c) {
					t.Errorf("unexpected character %q in generated string %q", c, str1)
				}
			}
		})
	}
}

func TestNewRandomString_ZeroSize(t *testing.T) {
	str := NewRandomString(0)

	if str != "" {
		t.Errorf("expected empty string, got %q", str)
	}
}
