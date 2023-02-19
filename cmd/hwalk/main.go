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
	"time"

	"github.com/dalzilio/hue/pkg/formula"
	"github.com/dalzilio/hue/pkg/hlnet"
	"github.com/dalzilio/hue/pkg/pnml"
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

type qflags struct {
	showqueries   *bool
	testfsimplify *bool
	hidetrivial   *bool
	hideundef     *bool
	showwitness   *bool
}

func main() {

	// Current known limitations on the models in the MCC
	//
	// * cannot fire transition "Start" on model DatabaseWithMutex; extra
	//   variable in the output arcs
	// * cannot check fireability on model PhilosophersDyn; more variables on
	//   the transition condition than on the input arc + some arcs have an <All>
	//   expression in their pattern. Same with VehicularWifi-COL-none ;
	//   UtilityControlRoom ; and PhilosophersDyn (already forbidden)

	var flaghelp = flag.BoolP("help", "h", false, "print this message")
	var flagstat = flag.BoolP("stat", "s", false, "print statistics information")
	var flagversion = flag.Bool("version", false, "print version number and generation date then quit")

	var netfile = flag.StringP("net", "n", "", "path to PNML file (see option -d if absent)")
	var propfile = flag.String("xml", "", "path to XML file with the Reachability formulas")
	var dirfile = flag.StringP("directory", "d", "", "path to folder containing model.pnml and XML formulas (DEFAULT)")

	var flagreach = flag.BoolP("reachability", "r", false, "check ReachabilityCardinality.xml file")
	var flagfire = flag.BoolP("fireability", "f", false, "check ReachabilityFireability.xml file")

	// to support calls like "--select-queries 12,14"
	var selectQueries []int
	flag.IntSliceVar(&selectQueries, "select-queries", []int{}, "comma separated list of queries id")

	rflags := qflags{}
	rflags.hidetrivial = flag.Bool("hide-trivial", false, "hide results for trivial queries")
	rflags.hideundef = flag.Bool("hide-undef", false, "hide queries with UNDEF result")
	rflags.showwitness = flag.Bool("show-witness", false, "print witness for enabled transitions")
	rflags.showqueries = flag.Bool("show-queries", false, "print queries on standard output")
	rflags.testfsimplify = flag.Bool("test-simplify", false, "print warning if formulas before and after simplification give different results")

	flag.CommandLine.SortFlags = false

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "hwalk version %s -- %s -- LAAS/CNRS\n", gitversion, builddate)
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
		fmt.Printf("hue version %s -- %s -- LAAS/CNRS\n", gitversion, builddate)
		os.Exit(1)
	case *flaghelp:
		flag.Usage()
		os.Exit(0)
	case *netfile == "" && *propfile != "":
		fmt.Println("cannot use option --xml without --net")
		flag.Usage()
		os.Exit(0)
	case *propfile != "" && !*flagreach && !*flagfire:
		fmt.Println("cannot use option --xml without -r or -f")
		flag.Usage()
		os.Exit(0)
	case *netfile != "" && *dirfile != "":
		fmt.Println("cannot use options --net and --directory together")
		flag.Usage()
		os.Exit(0)
	case *flagfire && *flagreach:
		fmt.Println("cannot use options -f and -r together")
		flag.Usage()
		os.Exit(0)
	case (*flagfire || *flagreach) && *propfile == "" && *dirfile == "" && N != 1:
		fmt.Println("cannot use -f and -r without a property file")
		flag.Usage()
		os.Exit(0)
	case N != 0 && (*netfile != "" && *dirfile != ""):
		fmt.Println("bad command line, extra parameter with option --net")
		flag.Usage()
		os.Exit(1)
	case N > 1:
		fmt.Println("bad command line, too many files")
		flag.Usage()
		os.Exit(1)
	case N == 1 && *netfile == "" && *dirfile == "":
		*dirfile = flag.Arg(0)
	}

	sort.Ints(selectQueries)

	// // we capture panics
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		log.Fatal("error in generation: cannot compute")
	// 		os.Exit(1)
	// 	}
	// }()

	var xmlFile *os.File
	var err error

	start := time.Now()

	if *dirfile != "" {
		// we check the XML directory exists
		xmlDirInfo, err := os.Stat(*dirfile)
		if err != nil {
			log.Println("Error with directory:", err)
			os.Exit(1)
			return
		}
		if !xmlDirInfo.IsDir() {
			log.Printf("Path %s is not a directory\n", *dirfile)
			os.Exit(1)
			return
		}
		// we try to open the model.xml file from it
		xmlFile, err = os.Open(filepath.Join(*dirfile, "model.pnml"))
		if err != nil {
			log.Println("Error opening file:", err)
			os.Exit(1)
			return
		}
		defer xmlFile.Close()
	}

	if *netfile != "" {
		xmlFile, err = os.Open(*netfile)
		if err != nil {
			log.Println("Error opening file:", err)
			os.Exit(1)
			return
		}
		defer xmlFile.Close()
	}

	decoder := pnml.NewDecoder(xmlFile)

	var p = new(pnml.Net)
	err = decoder.Build(p)
	if err != nil {
		log.Println("Error decoding PNML file:", err)
		os.Exit(1)
		return
	}

	hl, err := hlnet.Build(p)
	if err != nil {
		log.Println("Error decoding PNML file:", err)
		os.Exit(1)
		return
	}

	// we need fireability information with options -f and --show-witness
	s := hlnet.NewStepper(hl, *flagfire || *rflags.showwitness)

	if *flagstat {
		elapsed := time.Since(start)
		fmt.Fprintf(os.Stdout, "# net %s, %d place(s), %d transition(s), %.3fs\n", hl.Name, len(hl.Places), len(hl.Trans), elapsed.Seconds())
		fmt.Fprintf(os.Stdout, "%s\n", hl)
		fmt.Fprintf(os.Stdout, "%s\n", s)
	}

	if *rflags.showwitness {
		s.PrintWitnesses()
	}

	if *flagfire || *flagreach {
		switch {
		case *propfile != "":
			xmlFile, err = os.Open(*propfile)
		case *dirfile != "":
			if *flagreach {
				xmlFile, err = os.Open(filepath.Join(*dirfile, "ReachabilityCardinality.xml"))
			}
			if *flagfire {
				xmlFile, err = os.Open(filepath.Join(*dirfile, "ReachabilityFireability.xml"))
			}
		default:
			fmt.Println("bad command line, missing XML properties")
			flag.Usage()
			os.Exit(1)
		}

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

		if len(selectQueries) != 0 {
			// we mark the queries that need to be skip
			for k := range queries {
				queries[k].Skip = true
			}
			for _, k := range selectQueries {
				queries[k].Skip = false
			}
		}

		checkquery(s, queries, rflags)
	}
}

func checkquery(s *hlnet.Stepper, queries []formula.Query, flags qflags) {
	for _, q := range queries {
		if q.Skip {
			continue
		}
		v := s.EvaluateQueries(q)
		if *flags.testfsimplify {
			if !s.EvaluateAndTestSimplify(q) {
				fmt.Println("----------------------------------")
				fmt.Printf("SIMPLIFY ERROR in formula %s\n", q.ID)
				fmt.Fprintf(os.Stdout, "ORIGINAL: %s\n", q.Original.String())
				fmt.Fprintf(os.Stdout, "SIMPLIFY: %s\n", q.Formula.String())
				fmt.Println("----------------------------------")
			}
		}
		if !*flags.hideundef || (*flags.hideundef && (v != formula.UNDEF)) {
			if !*flags.hidetrivial || (*flags.hidetrivial && !q.IsTrivial()) {
				if *flags.showqueries {
					fmt.Fprint(os.Stdout, q.String())
					fmt.Fprintf(os.Stdout, "VERDICT\t%s\n", v)
					fmt.Println("----------------------------------")
				} else {
					fmt.Fprintf(os.Stdout, "FORMULA %s %s\n", q.ID, v)
				}
			}
		}
	}
}
