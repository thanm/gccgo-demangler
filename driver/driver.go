//
// Reads gccgo AST dumps and performs symbol demangling.
//

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/thanm/gccgo-demangler/demangler"
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

func filter(inf *os.File, outf *os.File) error {
	// Copy in to out
	scanner := bufio.NewScanner(inf)
	for scanner.Scan() {
		fmt.Fprintf(outf, "%s\n", demangler.DemangleLine(scanner.Text()))
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
