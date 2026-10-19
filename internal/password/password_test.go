package password

import (
	"strings"
	"testing"
)

func TestEncodeKnownValue(t *testing.T) {
	if got, want := Encode("Deneme123*"), "38YWTB0QKYOBU08YNFEI1"; got != want {
		t.Fatalf("Encode = %s, beklenen %s", got, want)
	}
}

func TestEncodeLength(t *testing.T) {
	for _, tc := range []struct {
		plain string
		size  int
	}{
		{"", 0},
		{"a", 7},
		{"abcd", 7},
		{"abcde", 14},
		{"abcdefghijkl", 21},
	} {
		got := Encode(tc.plain)
		if len(got) != tc.size {
			t.Errorf("%q: uzunluk %d, beklenen %d", tc.plain, len(got), tc.size)
		}
		if strings.Trim(got, alphabet) != "" {
			t.Errorf("%q: alfabe dışı karakter: %s", tc.plain, got)
		}
	}
}

func TestBitOrderIsPermutation(t *testing.T) {
	var seen [32]bool
	for _, b := range bitOrder {
		if seen[b] {
			t.Fatalf("bit %d iki kez kullanılmış", b)
		}
		seen[b] = true
	}
}
