//
// Reads gccgo AST dumps and performs symbol demangling.
//

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
)

var verbflag = flag.Int("v", 0, "Verbose trace output level")
var infileflag = flag.String("i", "", "Input file")
var outfileflag = flag.String("o", "", "Output file")

func verb(vlevel int, s string, a ...interface{}) {
	if *verbflag >= vlevel {
		fmt.Printf(s, a...)
		fmt.Printf("\n")
	}
}

func usage(msg string) {
	if len(msg) > 0 {
		fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	}
	fmt.Fprintf(os.Stderr, "usage: demangler [flags]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

var singletons = map[byte]string{
	'E': "error",
	'z': "string",
	'v': "void",
	'b': "boolean",
	'n': "nil",
}

func demangle(id []byte) (res []byte, err error) {
	if len(id) == 0 {
		return []byte{}, nil
	}
	res = []byte{}
	switch id[0] {
	case 'E', 'z', 'v', 'b', 'n':
		initial := []byte(singletons[id[0]])
		remain, err := demangle(id[1:])
		if err != nil {
			return id, err
		}
		return append(initial, remain...), nil

	case 'A':
		// A => array (A element [dd]e)

	default:
		return id, errors.New("unmatched")
	}
}

func process_line(line string) string {
	verb(1, "processing line :%s:", line)
	bytes := []byte(line)
	idsre := regexp.MustCompile(`\pL[\pL\pN]*`)
	m := idsre.FindAllSubmatchIndex(bytes, -1)
	if len(m) == 0 {
		return line
	}
	res := []byte{}
	sslot := 0
	for _, s := range idsre.FindAllSubmatchIndex(bytes, -1) {
		verb(2, "sslot: %d chunk: %v", sslot, s)
		res = append(res, bytes[sslot:s[0]]...)
		identifier := bytes[s[0]:s[1]]
		dem, err := demangle(identifier)
		if err != nil {
			dem = identifier
		}
		res = append(res, dem...)
		sslot = s[1]
	}
	res = append(res, bytes[sslot:len(bytes)]...)
	verb(1, "res = :%s:", string(res))
	return string(res)
}

func filter(inf *os.File, outf *os.File) error {
	// Copy in to out
	scanner := bufio.NewScanner(inf)
	for scanner.Scan() {
		fmt.Fprintf(outf, "%s\n", process_line(scanner.Text()))
	}
	return nil
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("demangler: ")
	flag.Parse()
	verb(1, "in main")
	if flag.NArg() != 0 {
		usage("unknown extra args")
	}
	var err error
	var infile *os.File = os.Stdin
	if len(*infileflag) > 0 {
		verb(1, "opening %s", *infileflag)
		infile, err = os.Open(*infileflag)
		if err != nil {
			log.Fatal("%v", err)
		}
	}
	var outfile *os.File = os.Stdout
	if len(*outfileflag) > 0 {
		verb(1, "opening %s", *outfileflag)
		outfile, err = os.OpenFile(*outfileflag, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Fatal("%v", err)
		}
	}
	err = filter(infile, outfile)
	if err != nil {
		log.Fatal("%v", err)
	}
	verb(1, "leaving main")
}
