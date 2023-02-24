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
	"sync"
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

// qflags are used to pass the values of commandline flags around
type qflags struct {
	testfsimplify bool
	verbose       bool
	reach         bool
	fire          bool
	countlimit    int
}

// qlist is a concurrency-safe list of query ids that still need to be
// evaluated.
type qlist struct {
	mu              sync.Mutex
	queries         []formula.Query
	queriesactive   []bool
	nbactivequeries int // number of values set to true in queriesleft
}

type qmessage struct {
	id      int
	verdict formula.Bool
}

func main() {

	// Current known limitations on the models in the MCC
	//
	// * cannot fire transition "Start" on model DatabaseWithMutex; extra
	//   variable in the output arcs
	// * cannot check fireability on model PhilosophersDyn; more variables on
	//   the transition condition than on the input arc + some arcs have an <All>
	//   expression in their pattern. Same with VehicularWifi-COL-none
	var flaghelp = flag.BoolP("help", "h", false, "print this message")
	var flagversion = flag.Bool("version", false, "print version number and generation date then quit")

	var verbose = flag.BoolP("verbose", "v", false, "print witness for enabled transitions")
	var flagquiet = flag.BoolP("quiet", "q", true, "set output to a minimum (use -q=false to print logs)")

	var netfile = flag.StringP("net", "n", "", "path to PNML file (see option -d if absent)")
	var propfile = flag.String("xml", "", "path to XML file with the Reachability formulas")
	var dirfile = flag.StringP("directory", "d", "", "path to folder containing model.pnml and XML formulas")

	var reach = flag.BoolP("reachability", "r", false, "check ReachabilityCardinality.xml file")
	var fire = flag.BoolP("fireability", "f", false, "check ReachabilityFireability.xml file")

	// to support calls like "--select-queries 12,14"
	var selectQueries []int
	flag.IntSliceVar(&selectQueries, "select-queries", []int{}, "comma separated list of queries id")

	var countlimit = flag.IntP("limit-count", "c", 1, "limit on length of exploration path (0 means none)")
	// var flagtimelimit = flag.IntP("limit-time", "t", 0, "limit on time of exploration (0 means none)")
	var flagparallel = flag.IntP("parallel", "p", 4, "number of walkers operating in parallel")

	var testfsimplify = flag.Bool("test-simplify", false, "print warning if formulas before and after simplification give different results")
	var showqueries = flag.Bool("show-queries", false, "print queries on standard output")

	flag.CommandLine.SortFlags = false

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "uwalk version %s -- %s -- LAAS/CNRS\n", gitversion, builddate)
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nfiles:\n")
		fmt.Fprintf(os.Stderr, "   outfile: output is always on stdout\n")
		fmt.Fprintf(os.Stderr, "   errorfile: error are always printed on stderr\n")
	}

	flag.Parse()

	// ----------------------------------------------------------------------
	// Managing options

	rflags := qflags{
		testfsimplify: *testfsimplify,
		verbose:       *verbose,
		reach:         *reach,
		fire:          *fire,
		countlimit:    *countlimit,
	}

	N := len(flag.Args())

	switch {
	case *flagversion:
		fmt.Printf("uwalk version %s -- %s -- LAAS/CNRS\n", gitversion, builddate)
		os.Exit(1)
	case *flaghelp:
		flag.Usage()
		os.Exit(0)
	case *netfile == "" && *propfile != "":
		fmt.Println("cannot use option --xml without --net")
		flag.Usage()
		os.Exit(0)
	case *propfile != "" && !rflags.reach && !rflags.fire:
		fmt.Println("cannot use option --xml without -r or -f")
		flag.Usage()
		os.Exit(0)
	case *netfile != "" && *dirfile != "":
		fmt.Println("cannot use options --net and --directory together")
		flag.Usage()
		os.Exit(0)
	case rflags.fire && rflags.reach:
		fmt.Println("cannot use options -f and -r together")
		flag.Usage()
		os.Exit(0)
	case (rflags.fire || rflags.reach) && *propfile == "" && *dirfile == "" && N != 1:
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

	// we capture panics when in quiet/competition mode
	if *flagquiet {
		defer func() {
			if r := recover(); r != nil {
				log.Fatal("error in generation: cannot compute")
				os.Exit(1)
			}
		}()
	}

	// ----------------------------------------------------------------------
	// Parsing model

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

	// ----------------------------------------------------------------------
	// Building Stepper and computing initial marking. We could wait before
	// computing fireability information that we check cardinality information
	// on the initial marking, but this is not useful in practice.

	s := hlnet.NewStepper(hl)

	if rflags.verbose {
		elapsed := time.Since(start)
		fmt.Fprintf(os.Stdout, "# net %s, %d place(s), %d transition(s), %.3fs\n", hl.Name, len(hl.Places), len(hl.Trans), elapsed.Seconds())
		fmt.Fprintf(os.Stdout, "%s\n", hl)
		fmt.Println("")
	}

	// -------------------------------------------------------------------------
	// In the absence of queries: we do not parse formula and we simply execute
	// one walker.

	if !rflags.fire && !rflags.reach {
		w := hlnet.NewWorker(s)
		if rflags.verbose {
			fmt.Print(w)
		}
		count := 0
		for {
			count++
			if rflags.countlimit != 0 {
				if count >= rflags.countlimit {
					return
				}
			}
			w.FireAtRandom(rflags.verbose)
			if rflags.verbose {
				fmt.Print(w)
			}
		}
	}

	// ----------------------------------------------------------------------
	// Parsing formulas

	switch {
	case *propfile != "":
		xmlFile, err = os.Open(*propfile)
	case *dirfile != "":
		if rflags.reach {
			xmlFile, err = os.Open(filepath.Join(*dirfile, "ReachabilityCardinality.xml"))
		}
		if rflags.fire {
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

	fdecoder := formula.NewDecoder(xmlFile)

	var worklist qlist

	worklist.queries, err = fdecoder.Build()
	if err != nil {
		log.Fatal("Error decoding Formula file:", err)
		os.Exit(1)
		return
	}

	// We maintain a slice of queries that need to be evaluated. This is a slice
	// of bool such that queriesleft[i] true when queries[i] still needs to be
	// evaluated.
	worklist.queriesactive = make([]bool, len(worklist.queries))
	if len(selectQueries) == 0 {
		// This is the default. We need to include all the queries.
		for k := range worklist.queries {
			worklist.queriesactive[k] = true
		}
		worklist.nbactivequeries = len(worklist.queries)
	} else {
		for _, k := range selectQueries {
			worklist.queriesactive[k] = true
			worklist.nbactivequeries++
		}
	}

	if *showqueries {
		fmt.Println("\n----------------------------------")
		for k, b := range worklist.queriesactive {
			if b {
				fmt.Println(worklist.queries[k].String())
				fmt.Println("----------------------------------")
			}
		}
		fmt.Println("")
	}

	// We can receive messages from all the workers at the same time. We expect
	// a pair of the query index and its verdict.
	msgmain := make(chan qmessage, *flagparallel*len(worklist.queries))
	// We use msgend to signal the end of processing to the main function
	msgend := make(chan bool, 1)
	// We also use a WaitGroup to detect when all the workers have finished
	// (typically, when they have reached their count limit)
	var wg sync.WaitGroup
	wg.Add(*flagparallel)
	// We have a channel for each worker, used to receive information about
	// query verdicts
	msgworkers := make([]chan bool, *flagparallel)
	workers := make([]*hlnet.Worker, *flagparallel)
	for i := 0; i < *flagparallel; i++ {
		// We should not send more than len(queries) messages to each worker on this channel.
		msgworkers[i] = make(chan bool, len(worklist.queries))
		workers[i] = hlnet.NewWorker(s)

	}

	go querycoordinator(&worklist, rflags,
		&wg, msgend, msgmain, msgworkers)

	for i := 0; i < *flagparallel; i++ {
		// queries is only used for read access, so we can share it between
		// workers.
		go queryworker(i, workers[i], worklist.queries, &worklist, rflags,
			&wg, msgmain, msgworkers[i])
	}

	// We wait from querycoordinator to send an end event.
	<-msgend
}

// ----------------------------------------------------------------------

// querycoordinator manages the queue of query identifiers that should be
// checked. It is in charge of printing the result to the standard output. It
// receives information form workers when a verdict has been found and broadcast
// the information to all the workers.
func querycoordinator(worklist *qlist, flags qflags,
	wg *sync.WaitGroup, msgend chan<- bool, msgmain <-chan qmessage, msgworkers []chan bool) {

	go func() {
		// if all the workers have finished, we signal the main program
		wg.Wait()
		msgend <- true
	}()

	for qm := range msgmain {
		// we receive a result for query qm.id
		if worklist.queriesactive[qm.id] {
			// we update the info and print the result
			if flags.verbose {
				fmt.Println("")
			}
			worklist.mu.Lock()
			fmt.Printf("FORMULA %s %s\n", worklist.queries[qm.id].ID, qm.verdict)
			worklist.queriesactive[qm.id] = false
			worklist.nbactivequeries--
			worklist.mu.Unlock()
			if worklist.nbactivequeries == 0 {
				msgend <- true
				return
			}
			// we warn the workers that the list has been updated
			for _, c := range msgworkers {
				c <- true
			}
		}
	}
}

// queryworker manages a walker working in parallel with the others. It sends a
// message to msgmain when it has found a verdict and awaits messages on msgin
// (non blocking) to know if it needs to update its copy of queriesleft.
func queryworker(id int, w *hlnet.Worker, queries []formula.Query, worklist *qlist, flags qflags,
	wg *sync.WaitGroup, msgmain chan<- qmessage, msgin <-chan bool) {
	count := 0
	// We keep a local copy of the list of active queries
	queriesactive := make([]bool, len(queries))
	nbactivequeries := 0
	worklist.mu.Lock()
	for k, b := range worklist.queriesactive {
		if b {
			nbactivequeries++
			queriesactive[k] = true
		}
	}
	worklist.mu.Unlock()
	for {
		if nbactivequeries == 0 {
			wg.Done()
			return
		}
		if flags.countlimit != 0 {
			if count >= flags.countlimit {
				wg.Done()
				return
			}
			count++
		}

		w.FireAtRandom(flags.verbose)

		for k, b := range queriesactive {
			if !b {
				continue
			}
			var v formula.Bool

			if flags.reach {
				v = w.EvaluateCardinalityQueries(queries[k])
			} else {
				v = w.EvaluateFireabilityQueries(queries[k])
			}

			if flags.testfsimplify {
				if !w.EvaluateAndTestSimplify(queries[k]) {
					fmt.Println("----------------------------------")
					fmt.Printf("SIMPLIFY ERROR in formula %s\n", queries[k].ID)
					fmt.Fprintf(os.Stdout, "ORIGINAL: %s\n", queries[k].Original.String())
					fmt.Fprintf(os.Stdout, "SIMPLIFY: %s\n", queries[k].Formula.String())
					fmt.Println("----------------------------------")
				}
			}

			if _, ok := v.Value(); ok {
				// We have a new verdict. We send it to the coordinator.
				msgmain <- qmessage{k, v}
				// We can skip this query from now.
				queriesactive[k] = false
				nbactivequeries--
			}
		}

		// We listen to the coordinator to find if another worker has found a
		// verdict.
		select {
		case <-msgin:
			// We should check the list of active queries. BEWARE: we could know
			// the verdict of a query before the coordinator (because our
			// message is still not processed)
			worklist.mu.Lock()
			for k, b := range worklist.queriesactive {
				if !b && queriesactive[k] {
					// someone else found the result for this query
					queriesactive[k] = false
					nbactivequeries--
				}
			}
			worklist.mu.Unlock()
		default:
			// we should not block waiting
		}
	}
}
