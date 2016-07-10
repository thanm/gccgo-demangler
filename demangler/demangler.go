//
// Reads gccgo AST dumps and performs symbol demangling.
//

package demangler

import (
	"errors"
	"fmt"
	"regexp"
)

// Regular expression for an embedded length
var emlenre *regexp.Regexp = regexp.MustCompile(`([0-9])+[^0-9].*`)

// Get embedded length, return value and num chars it took up
func getemlen(id []byte) (length int, nchars int, err error) {
	// length
	lensl := emlenre.FindSubmatch(id)
	if len(lensl) != 2 {
		return 0, 0, errors.New("embedded length match failed")
	}
	nmatched, serr := fmt.Sscanf(string(lensl[1]), "%d", &length)
	if serr != nil {
		return 0, 0, serr
	}
	if nmatched != 1 {
		return 0, 0, errors.New("embedded length scanf failed")
	}
	nchars = len(lensl[1])
	fmt.Printf("getemlen %s returns l=%d nc=%d\n", string(id), length, nchars)
	return
}

// A => array (A element [dd]e)
func dem_array(id []byte) (res []byte, consumed int, err error) {
	// element type
	elemt, elemcon, eerr := dem(id[1:])
	if eerr != nil {
		return []byte{}, 0, eerr
	}
	fmt.Printf("dem_array: elemcon=%d\n", elemcon)

	// length
	arlen, lchars, lerr := getemlen(id[1+elemcon:])
	if lerr != nil {
		return []byte{}, 0, lerr
	}

	// trailing "e"
	if id[elemcon + lchars + 1] != 'e' {
		return []byte{}, 0, errors.New("trailing 'e' for array missing")
	}

	// success
	res = []byte(fmt.Sprintf("[%d]%s", arlen, string(elemt)))
	consumed = elemcon + lchars + 2
	err = nil
	return
}

// N => named type (N dd_ name)
func dem_named(id []byte) (res []byte, consumed int, err error) {
	// length
	length, lchars, lerr := getemlen(id[1:])
	if lerr != nil {
		return []byte{}, 0, lerr
	}

	// underscore
	if id[lchars+1] != '_' {
		return []byte{}, 0, errors.New("named type missing underscore")
	}

	// success
	return id[lchars+2:lchars+length+2], lchars+length+2, nil
}


// A => array (A element [dd]e)
// b => boolean
// C => channel (C element [s][r]e)
// c => complex (c [a]bits e)
// E => error
// f => float (f [a]bits e)
// F => function (F [m receiver] [p params e] [r results e] e)
// i => integer (i [a][u]bits e)
// I => interface (I (method-name method-type) e)
// M => map (M keytype __ valtype)
// n => nil
// N => named type (N dd_ name)
// p => pointer (p points-to)
// S => struct (S (field-name field-type [T dd_ tag]) e)
// v => void
// V => varargs [varargs-type]
// z => string

var singletons = map[byte]string{
	'E': "error",
	'z': "string",
	'v': "void",
	'b': "boolean",
	'n': "nil",
}

func dem(id []byte) (res []byte, consumed int, err error) {
	if len(id) == 0 {
		return []byte{}, 0, errors.New("premature EOS")
	}
	switch id[0] {
	case 'E', 'z', 'v', 'b', 'n':
		return []byte(singletons[id[0]]), 1, nil
	case 'A':
		// A => array (A element [dd]e)
		return dem_array(id)
	case 'N':
		// N => named type (N dd_ name)
		return dem_named(id)
	default:
		return []byte{}, 0, errors.New("unmatched")
	}
	return []byte{}, 0, errors.New("what happened?")
}

func Demangle(token string) string {
	btoken := []byte(token)
	dtoken, consumed, err := dem(btoken)
	if err != nil {
		dtoken = btoken
	}
	if len(token) != consumed {
		dtoken = btoken
	}

	return string(dtoken)
}

// Regular expression for a go identifier
var idsre *regexp.Regexp = regexp.MustCompile(`\pL[\pL\pN]*`)

func DemangleLine(line string) string {
	bytes := []byte(line)
	m := idsre.FindAllSubmatchIndex(bytes, -1)
	if len(m) == 0 {
		return line
	}
	res := []byte{}
	sslot := 0
	for _, s := range idsre.FindAllSubmatchIndex(bytes, -1) {
		res = append(res, bytes[sslot:s[0]]...)
		identifier := bytes[s[0]:s[1]]
		dem, consumed, err := dem(identifier)
		if err != nil || len(identifier) != consumed {
			dem = identifier
		}
		res = append(res, dem...)
		sslot = s[1]
	}
	res = append(res, bytes[sslot:len(bytes)]...)
	return string(res)
}
