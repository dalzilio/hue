// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

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

	var propfile = flag.String("xml", "", "path to XML file with the Reachability formulas")
	var dirfile = flag.StringP("directory", "d", "", "path to folder containing XML formulas")

	var reach = flag.BoolP("reachability", "r", false, "check ReachabilityCardinality.xml file")
	var fire = flag.BoolP("fireability", "f", false, "check ReachabilityFireability.xml file")

	var selectQueries []int

	flag.IntSliceVar(&selectQueries, "select-queries", []int{}, "comma separated list of queries id")

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

	switch {
	case *flagversion:
		fmt.Printf("usimplify version %s -- %s -- LAAS/CNRS\n", gitversion, builddate)
		os.Exit(0)
	case *flaghelp:
		flag.Usage()
		os.Exit(0)
	case !*reach && !*fire:
		fmt.Println("you should use either option -r or -f")
		flag.Usage()
		os.Exit(1)
	case *propfile != "" && *dirfile != "":
		fmt.Println("cannot use options --xml and --directory together")
		flag.Usage()
		os.Exit(1)
	case *fire && *reach:
		fmt.Println("cannot use options -f and -r together")
		flag.Usage()
		os.Exit(1)
	case N != 0:
		fmt.Println("bad command line, too many files")
		flag.Usage()
		os.Exit(1)
	}

	sort.Ints(selectQueries)

	// // we capture panics
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		log.Fatal("Error in generation: cannot compute")
	// 		os.Exit(1)
	// 	}
	// }()

	var formulafile string
	if *propfile != "" {
		formulafile = *propfile
	}
	if *dirfile != "" {
		if *reach {
			formulafile = filepath.Join(*dirfile, "ReachabilityCardinality.xml")
		}
		if *fire {
			formulafile = filepath.Join(*dirfile, "ReachabilityFireability.xml")
		}
	}

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

	selection := make([]bool, len(queries))
	if len(selectQueries) != 0 {
		for _, k := range selectQueries {
			selection[k] = true
		}
	} else {
		for i := 0; i < len(queries); i++ {
			selection[i] = true
		}
	}

	fmt.Printf("# %d formulas\n", len(queries))
	fmt.Println("=========================================")

	for k, q := range queries {
		if selection[k] {
			fmt.Println(q)
			if *fire {
				fmt.Println("-----")
				g, _ := formula.BddFireabilitySimplify(q.Original)
				fmt.Println(g.String())
			}
			fmt.Println("=========================================")
		}
	}
}
