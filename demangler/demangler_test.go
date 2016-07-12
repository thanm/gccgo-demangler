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
		"N10_main.Mango",
		"AN5_int328e",
		"AN5_int32e",
	}
	var cooked = []string{
		"int64",
		"main.Mango",
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
		"FppN5_int32pN5_int64erN4_boolIeee",
	}
	var cooked = []string{
		"func{()}",
		"func{(*int32, *int64) (bool, interface{})}",
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
		"interface{foo func{()}}",
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

func TestStruct(t *testing.T) {
	var raw = []string{
		"Se",
		"S1_bN4_bool3_bslAN4_boole3_f32N7_float323_f64N7_float643_u32N6_uint323_u64N6_uint643_i32N5_int323_i64N5_int644_pi32pN5_int324_fptrFe4_c128N10_complex1282_baAN5_uint832ee",
	}
	var cooked = []string{
		"struct{}",
		"struct{b bool, bsl []bool, f32 float32, f64 float64, u32 uint32, u64 uint64, i32 int32, i64 int64, pi32 *int32, fptr func{()}, c128 complex128, ba [32]uint8}",
	}
	for pos, r := range raw {
		c := Demangle(r)
		if c != cooked[pos] {
			save := Verbctl
			Verbctl = 2
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
			Verbctl = save
		}
	}
}
