# gccgo-demangler

Demangler for gccgo AST dumps. Walks through an input file and tries to apply demangling to anything that looks like a mangled type. Any identifier or symbol that doesn't demangle properly will just be left alone.

Example usage:

```

  Build:
  % cd $GOPATH
  % go get github.com/thanm/gccgo-demangler/gccgo-dem
  
  Run a symbol through the demangler:
  % echo I5_WriteFpAN5_uint8eerN3_intN5_erroreee | gccgo-dem
  interface{Write func{([]uint8) (int, error)}}

  Create AST dump via gccgo compile:
  % cd $GOPATH/src/github.com/thanm/gccgo-demangler/gccgo-dem
  % go build -compiler gccgo -gccgoflags -fgo-dump-ast driver.go
  % ls driver.go.dump.ast
  driver.go.dump.ast
  % gccgo-dem -i driver.go.dump.ast -o demangled.txt
  % fgrep 'struct{' demangled.txt | head -1
      tmp.39295392 (struct{res0 int, res1 error})

```
