//
// Reads gccgo AST dumps and performs symbol demangling.
//

package demangler

import (
	"errors"
	"fmt"
	"regexp"
)

var Verbctl int = 0

func verb(vlevel int, s string, a ...interface{}) {
	if Verbctl >= vlevel {
		fmt.Printf(s, a...)
		fmt.Printf("\n")
	}
}

// Regular expression for an embedded length
var emlenre *regexp.Regexp = regexp.MustCompile(`([0-9]+)[^0-9].*`)

// Get embedded length, return value and number of bytes it occupied
func getemlen(id []byte) (length int, nbytes int, err error) {
	// length
	lensl := emlenre.FindSubmatch(id)
	verb(2, "emlenre.FindSubmatch(%s) returns %v", string(id), lensl)

	if len(lensl) != 2 {
		return 0, 0, errors.New("embedded length match failed")
	}
	verb(2, "emlenre digstring is %s", string(lensl[1]))
	nmatched, serr := fmt.Sscanf(string(lensl[1]), "%d", &length)
	if serr != nil {
		return 0, 0, serr
	}
	if nmatched != 1 {
		return 0, 0, errors.New("embedded length scanf failed")
	}
	nbytes = len(lensl[1])
	verb(2, "getemlen %s returns l=%d nc=%d", string(id), length, nbytes)
	return
}

// A => array (A element [dd]e)
func dem_array(id []byte) (res []byte, consumed int, err error) {
	// element type

	elemt, elemcon, eerr := dem(id[1:])
	if eerr != nil {
		return []byte{}, 0, eerr
	}

	// slice?
	if id[elemcon+1] == 'e' {
		// success
		return []byte(fmt.Sprintf("[]%s", string(elemt))), elemcon + 2, nil
	}

	// length
	arlen, lchars, lerr := getemlen(id[1+elemcon:])
	if lerr != nil {
		return []byte{}, 0, lerr
	}

	// trailing "e"
	if id[elemcon+lchars+1] != 'e' {
		return []byte{}, 0, errors.New("trailing 'e' for array missing")
	}

	// success
	res = []byte(fmt.Sprintf("[%d]%s", arlen, string(elemt)))
	consumed = elemcon + lchars + 2
	err = nil
	return
}

// read <length> _ <name>
func dem_name(id []byte) (res []byte, consumed int, err error) {
	verb(2, "dem_name(%s)", string(id))

	// length
	length, lchars, lerr := getemlen(id)
	if lerr != nil {
		return []byte{}, 0, lerr
	}

	// underscore
	if id[lchars] != '_' {
		return []byte{}, 0, errors.New("named type missing underscore")
	}

	// success
	return id[lchars+1 : lchars+length+1], lchars + length + 1, nil
}

// I => interface (I (method-name method-type) e)
func dem_interface(id []byte) (res []byte, consumed int, err error) {
	idx := 1
	methodnames := make([][]byte, 0, 16)
	methodtypes := make([][]byte, 0, 16)

	for id[idx] != 'e' {

		// method name
		mname, mncons, mnerr := dem_name(id[idx:])
		if mnerr != nil || mncons == 0 {
			return []byte{}, 0, mnerr
		}
		methodnames = append(methodnames, mname)
		idx += mncons

		// method type
		mtype, mtcons, mterr := dem(id[idx:])
		if mterr != nil || mtcons == 0 {
			return []byte{}, 0, mterr
		}
		methodtypes = append(methodtypes, mtype)
		idx += mtcons
	}

	res = make([]byte, 0, idx)
	res = append(res, []byte("interface{")...)
	for i, mn := range methodnames {
		if i != 0 {
			res = append(res, []byte(", ")...)
		}
		res = append(res, mn...)
		res = append(res, []byte(" ")...)
		res = append(res, methodtypes[i]...)
	}
	res = append(res, []byte("}")...)
	return res, idx + 1, nil
}

// S => struct (S (field-name field-type [T dd_ tag]) [x] e)
func dem_struct(id []byte) (res []byte, consumed int, err error) {
	idx := 1
	fieldnames := make([][]byte, 0, 16)
	fieldtypes := make([][]byte, 0, 16)
	fieldtags := make([][]byte, 0, 16)

	if len(id) < 2 {
		return []byte{}, 0, errors.New("dem_struct premature EOS")
	}

	for id[idx] != 'e' {

		if idx+1 < len(id) && string(id[idx:idx+2]) == "xe" {
			verb(2, " incomparable struct")
			idx += 1
			break
		}

		// field name
		fname, fncons, fnerr := dem_name(id[idx:])
		if fnerr != nil || fncons == 0 {
			return []byte{}, 0, fnerr
		}
		fieldnames = append(fieldnames, fname)
		idx += fncons

		if idx-1 > len(id) {
			return []byte{}, 0, errors.New("dem_struct premature EOS")
		}

		// field type
		ftype, ftcons, fterr := dem(id[idx:])
		if fterr != nil || ftcons == 0 {
			return []byte{}, 0, fterr
		}
		fieldtypes = append(fieldtypes, ftype)
		idx += ftcons

		if idx-1 > len(id) {
			return []byte{}, 0, errors.New("dem_struct premature EOS")
		}

		verb(2, " field '%s' type '%s'", fname, ftype)

		if id[idx] == 'T' {
			idx += 1
			if idx-1 > len(id) {
				return []byte{}, 0, errors.New("dem_struct premature EOS")
			}
			ftag, ftgcons, ftgerr := dem_name(id[idx:])
			if ftgerr != nil || ftgcons == 0 {
				return []byte{}, 0, ftgerr
			}
			fieldtags = append(fieldtags, ftag)
			idx += fncons
		}
	}

	res = make([]byte, 0, idx)
	res = append(res, []byte("struct{")...)
	for i, mn := range fieldnames {
		if i != 0 {
			res = append(res, []byte(", ")...)
		}
		res = append(res, mn...)
		res = append(res, []byte(" ")...)
		res = append(res, fieldtypes[i]...)
	}
	res = append(res, []byte("}")...)
	return res, idx + 1, nil
}

// F => function (F [m receiver] [p params e] [r results e] e)
func dem_function(id []byte) (res []byte, consumed int, err error) {
	idx := 1

	verb(1, "examining function %s", string(id))

	var receiverType []byte
	for id[idx] == 'm' {
		verb(1, "starting receiver type")

		// receiver
		idx += 1
		rtype, rtcons, rterr := dem(id[idx:])
		if rterr != nil || rtcons == 0 {
			verb(1, "receiver type error %v", rterr)
			return []byte{}, 0, rterr
		}
		receiverType = rtype
		idx += rtcons
	}

	var paramTypes [][]byte
	varargs := ""
	if id[idx] == 'p' {
		verb(1, "starting params")

		// parameters
		idx += 1
		for id[idx] != 'e' {
			ptype, ptcons, pterr := dem(id[idx:])
			if pterr != nil || ptcons == 0 {
				return []byte{}, 0, pterr
			}
			paramTypes = append(paramTypes, ptype)
			idx += ptcons

			verb(1, "ptype %s", string(ptype))

			for id[idx] == 'V' {
				idx += 1
				varargs = "..."
			}
		}
		verb(1, "finished params")
		idx += 1
	}

	var resultTypes [][]byte
	if id[idx] == 'r' {
		verb(1, "starting returns")

		// results
		idx += 1
		for id[idx] != 'e' {
			rtype, rtcons, rterr := dem(id[idx:])
			if rterr != nil || rtcons == 0 {
				return []byte{}, 0, rterr
			}
			resultTypes = append(resultTypes, rtype)
			idx += rtcons
		}
		verb(1, "finished returns")
		idx += 1
	}

	if id[idx] != 'e' {
		return []byte{}, 0, errors.New("func type missing terminator")
	}

	res = make([]byte, 0, idx)
	res = append(res, []byte("func{")...)
	if len(receiverType) > 0 {
		rtclause := []byte(fmt.Sprintf("R(%s) ", string(receiverType)))
		res = append(res, rtclause...)
	}
	res = append(res, []byte("(")...)
	for i, pt := range paramTypes {
		if i != 0 {
			res = append(res, []byte(", ")...)
		}
		res = append(res, pt...)
	}
	if len(varargs) > 1 {
		res = append(res, []byte(varargs)...)
	}
	res = append(res, []byte(")")...)
	if len(resultTypes) > 0 {
		res = append(res, []byte(" ")...)
		if len(resultTypes) > 1 {
			res = append(res, []byte("(")...)
		}
		for i, rt := range resultTypes {
			if i != 0 {
				res = append(res, []byte(", ")...)
			}
			res = append(res, rt...)
		}
		if len(resultTypes) > 1 {
			res = append(res, []byte(")")...)
		}
	}
	res = append(res, []byte("}")...)

	return res, idx + 1, nil
}

// M => map (M keytype __ valtype)
func dem_map(id []byte) (res []byte, consumed int, err error) {
	idx := 1

	if idx-1 > len(id) {
		return []byte{}, 0, errors.New("dem_map premature EOS")
	}

	// key type
	kt, kcon, kerr := dem(id[idx:])
	if kerr != nil {
		return []byte{}, 0, kerr
	}
	idx += kcon

	if idx-4 > len(id) {
		return []byte{}, 0, errors.New("dem_map premature EOS")
	}
	if id[idx] != '_' || id[idx+1] != '_' {
		return []byte{}, 0, errors.New("dem_map missing __")
	}
	idx += 2

	// value type
	vt, vcon, verr := dem(id[idx:])
	if verr != nil {
		return []byte{}, 0, verr
	}
	idx += vcon

	return []byte(fmt.Sprintf("map[%s]%s", string(kt), string(vt))), idx, nil
}

// C => channel (C element [s][r]e)
func dem_chan(id []byte) (res []byte, consumed int, err error) {
	idx := 1

	// element type
	et, econ, eerr := dem(id[idx:])
	if eerr != nil {
		verb(1, "chan rule failed")
		return []byte{}, 0, eerr
	}
	idx += econ
	cansend := ""
	canrecv := ""
	if id[idx] == 's' {
		cansend = "<-"
		idx += 1
	}
	if id[idx] == 'r' {
		canrecv = "<-"
		idx += 1
	}
	if id[idx] != 'e' {
		return []byte{}, 0, errors.New("chan type missing end")
	}
	idx += 1
	if cansend == "" && canrecv == "" {
		cansend = "?"
		canrecv = "?"
	} else if cansend != "" && canrecv != "" {
		cansend = ""
		canrecv = ""
	}
	return []byte(fmt.Sprintf("%schan%s{%s}", canrecv, cansend, string(et))), idx, nil
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
	verb(1, "=-=-=-=-=\ndem(%s)", string(id))

	if len(id) == 0 {
		return []byte{}, 0, errors.New("premature EOS")
	}
	switch id[0] {
	case 'E', 'z', 'v', 'b', 'n':
		return []byte(singletons[id[0]]), 1, nil
	case 'A':
		// A => array (A element [dd]e)
		return dem_array(id)
	case 'S':
		// S => struct (S (field-name field-type [T dd_ tag]) e)
		return dem_struct(id)
	case 'N':
		// N => named type (N dd_ name)
		dres, dcons, derr := dem_name(id[1:])
		if derr != nil {
			verb(2, "name rule failed")
			return []byte{}, 0, derr
		}
		return dres, dcons + 1, nil
	case 'p':
		// p => pointer (p points-to)
		pt, pcon, perr := dem(id[1:])
		if perr != nil {
			verb(1, "ptr rule failed")
			return []byte{}, 0, perr
		}
		return []byte(fmt.Sprintf("*%s", string(pt))), pcon + 1, nil
	case 'C':
		// C => channel (C element [s][r]e)
		return dem_chan(id)
	case 'M':
		// M => map (M keytype __ valtype)
		return dem_map(id)
	case 'I':
		// I => interface (I (method-name method-type) e)
		return dem_interface(id)
	case 'F':
		// F => function (F [m receiver] [p params e] [r results e] e)
		return dem_function(id)
	default:
		msg := fmt.Sprintf("unmatched char %s", string(id[0]))
		return []byte{}, 0, errors.New(msg)
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
var idsre *regexp.Regexp = regexp.MustCompile(`[\pL_\.\$][\pL\pN_\.\$]*`)

func DemangleLine(line string) string {
	verb(1, "== DemangleLine(%s)", line)
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
		verb(1, "DemangleLine: dem(%s)", string(identifier))
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
