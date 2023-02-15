// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/dalzilio/hue/pkg/formula"
	flag "github.com/spf13/pflag"
)

// builddate stores the compilation date for the executable, in %Y%m%d format.
// Set in the Makefile with the result of:
//
//	date -u +%d/%m/%Y
var builddate string = "2020/01/01"

// gitversion stores the current git version. Set in the makefile with the result
// of:
//
//	git describe --tags --dirty --always
var gitversion string = "v0"

func main() {
	var flaghelp = flag.BoolP("help", "h", false, "print this message")
	var flagversion = flag.Bool("version", false, "print version number and generation date of hsimplify")

	flag.CommandLine.SortFlags = false

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "hsimplify version %s -- %s -- LAAS/CNRS\n", gitversion, builddate)
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nfiles:\n")
		fmt.Fprintf(os.Stderr, "   outfile: output is always on stdout\n")
		fmt.Fprintf(os.Stderr, "   errorfile: error are always printed on stderr\n")
	}

	flag.Parse()
	N := len(flag.Args())

	if *flaghelp {
		flag.Usage()
		os.Exit(0)
	}

	if *flagversion {
		fmt.Printf("hsimplify version %s -- %s -- LAAS/CNRS\n", gitversion, builddate)
	}

	switch {
	case N != 1:
		fmt.Println("should have exactly one XML formula file")
		flag.Usage()
		os.Exit(1)
	}

	// we capture panics
	defer func() {
		if r := recover(); r != nil {
			log.Fatal("Error in generation: cannot compute")
			os.Exit(1)
		}
	}()

	formulafile := flag.Arg(0)

	if filepath.Ext(formulafile) != ".xml" {
		log.Fatal("Wrong file extension!")
		os.Exit(1)
		return
	}

	xmlFile, err := os.Open(formulafile)
	if err != nil {
		log.Fatal("Error opening file:", err)
		os.Exit(1)
		return
	}
	defer xmlFile.Close()
	decoder := formula.NewDecoder(xmlFile)
	queries, err := decoder.Build()
	if err != nil {
		log.Fatal("Error decoding Formula file:", err)
		os.Exit(1)
		return
	}
	fmt.Fprintf(os.Stdout, "# %d formulas\n", len(queries))
	fmt.Fprint(os.Stdout, formula.PrintQueries(queries))
}
