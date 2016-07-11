
package demangler

import (
	"testing"
)

func TestBasic(t *testing.T) {
	var raw = []string{
		"",
		"99",
		"E",
		"Ex",
		"z",
		"v",
		"b",
		"n",
	}
	var cooked = []string{
		"",
		"99",
		"error",
		"Ex",
		"string",
		"void",
		"boolean",
		"nil",
	}
	for pos, r := range raw {
		c := Demangle(r)
		if c != cooked[pos] {
			t.Errorf("raw=%s decoded='%s' wanted '%s'",
				r, c, cooked[pos])
		}
	}
}

func TestArray(t *testing.T) {
	var raw = []string{
		"N5_int64",
		"AN5_int328e",
		"AN5_int32e",
	}
	var cooked = []string{
		"int64",
		"[8]int32",
		"[]int32",
	}
	for pos, r := range raw {
		c := Demangle(r)
		if c != cooked[pos] {
			d, consumed, err := dem([]byte(r))
			if err != nil {
				t.Errorf("raw=%s decoded='%s' wanted '%s' err=%v",
					r, c, cooked[pos], err)
			} else if len(c) != consumed {
				t.Errorf("raw=%s decoded='%s' wanted '%s' consumed=%d len=%d",
					r, c, cooked[pos], consumed, len(c))
			} else {
				t.Errorf("raw=%s decoded='%s' wanted '%s' no error?",
					r, string(d), cooked[pos])
			}
		}
	}
}

func TestPointer(t *testing.T) {
	var raw = []string{
		"pN5_int64",
		"pIe",
	}
	var cooked = []string{
		"*int64",
		"*interface{}",
	}
	for pos, r := range raw {
		c := Demangle(r)
		if c != cooked[pos] {
			d, consumed, err := dem([]byte(r))
			if err != nil {
				t.Errorf("raw=%s decoded='%s' wanted '%s' err=%v",
					r, c, cooked[pos], err)
			} else if len(c) != consumed {
				t.Errorf("raw=%s decoded='%s' wanted '%s' consumed=%d len=%d",
					r, c, cooked[pos], consumed, len(c))
			} else {
				t.Errorf("raw=%s decoded='%s' wanted '%s' no error?",
					r, string(d), cooked[pos])
			}
		}
	}
}

func TestFunction(t *testing.T) {
	var raw = []string{
		"Fe",
		"FmpN10_main.MangopN3_interN3_intee",
	}
	var cooked = []string{
		"func{()}",
		"splot",
	}
	for pos, r := range raw {
		c := Demangle(r)
		if c != cooked[pos] {
			d, consumed, err := dem([]byte(r))
			if err != nil {
				t.Errorf("raw=%s decoded='%s' wanted '%s' err=%v",
					r, c, cooked[pos], err)
			} else if len(c) != consumed {
				t.Errorf("raw=%s decoded='%s' wanted '%s' consumed=%d len=%d",
					r, c, cooked[pos], consumed, len(c))
			} else {
				t.Errorf("raw=%s decoded='%s' wanted '%s' no error?",
					r, string(d), cooked[pos])
			}
		}
	}
}

func TestInterface(t *testing.T) {
	var raw = []string{
		"Ie",
		"I3_fooFee",
	}
	var cooked = []string{
		"interface{}",
		"interface{foo func()}",
	}
	for pos, r := range raw {
		c := Demangle(r)
		if c != cooked[pos] {
			d, consumed, err := dem([]byte(r))
			if err != nil {
				t.Errorf("raw=%s decoded='%s' wanted '%s' err=%v",
					r, c, cooked[pos], err)
			} else if len(c) != consumed {
				t.Errorf("raw=%s decoded='%s' wanted '%s' consumed=%d len=%d",
					r, c, cooked[pos], consumed, len(c))
			} else {
				t.Errorf("raw=%s decoded='%s' wanted '%s' no error?",
					r, string(d), cooked[pos])
			}
		}
	}
}
