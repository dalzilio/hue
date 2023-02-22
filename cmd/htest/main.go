// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	flag "github.com/spf13/pflag"

	_ "embed"
)

//go:embed total.txt
var oracleFile string

func main() {
	var flaghelp = flag.BoolP("help", "h", false, "print this message")

	var flagcompare = flag.BoolP("compare", "c", false, "compare two different result files")
	var flaglist = flag.BoolP("list", "l", false, "list the new results or the differences (with --compare)")

	flag.CommandLine.SortFlags = false

	flag.Usage = func() {
		fmt.Println("htest -- LAAS/CNRS")
		fmt.Printf("Usage of %s:\n", os.Args[0])
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

	if !*flagcompare && N != 1 {
		log.Fatal("usage: htest file.txt")
		os.Exit(1)
		return
	}

	if *flagcompare && N != 2 {
		log.Fatal("usage: htest --compare A.txt B.txt")
		os.Exit(1)
		return
	}

	resultfile, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal("Error opening file:", err)
		os.Exit(1)
		return
	}
	defer resultfile.Close()

	// ----------------------------------------------------------------------

	results := make(map[string]bool)
	keys := []string{}
	resultScanner := bufio.NewScanner(resultfile)
	for resultScanner.Scan() {
		line := strings.Split(resultScanner.Text(), " ")

		if len(line) < 3 || line[0] != "FORMULA" {
			log.Panicf("parsing %s: found %s\n", flag.Arg(1), ZipString(line, "", "", " "))
		}
		keys = append(keys, line[1])
		if line[2] == "TRUE" {
			results[line[1]] = true
			continue
		}
		if line[2] == "FALSE" {
			results[line[1]] = false
			continue
		}
	}
	sort.Strings(keys)

	if *flagcompare {
		fmt.Printf("file %s (A): %d distinct results\n", resultfile.Name(), len(results))
	} else {
		fmt.Printf("file %s: %d distinct results\n", resultfile.Name(), len(results))
	}

	// ----------------------------------------------------------------------

	if *flagcompare {
		comparefile, err := os.Open(flag.Arg(1))
		if err != nil {
			log.Fatal("Error opening file:", err)
			os.Exit(1)
			return
		}
		defer comparefile.Close()

		compare(results, keys, *flaglist, comparefile)
		return
	}

	// ----------------------------------------------------------------------

	oracle := make(map[string]bool)
	newres := []string{}
	errors := []string{}

	// we test against the oracle
	oracleScanner := bufio.NewScanner(strings.NewReader(oracleFile))
	for oracleScanner.Scan() {
		line := strings.Split(oracleScanner.Text(), " ")
		if line[2] == "TRUE" {
			oracle[line[1]] = true
			continue
		}
		if line[2] == "FALSE" {
			oracle[line[1]] = false
			continue
		}
		log.Panicf("parsing oracle: found %s\n", ZipString(line, "", "", " "))
	}

	fmt.Printf("oracle: %d results\n", len(oracle))

	for _, q := range keys {
		res := results[q]
		answ, ok := oracle[q]
		if !ok {
			newres = append(newres, q)
			continue
		}
		if answ != res {
			errors = append(errors, q)
			if answ {
				fmt.Println("FORMULA " + q + " TRUE")
			}
			fmt.Println("FORMULA " + q + " FALSE")
		}
	}

	fmt.Printf("\n%d new results\n", len(newres))
	fmt.Printf("%d differences\n", len(errors))
	if *flaglist {
		for _, q := range newres {
			fmt.Println(q)
		}
	}
}

// ----------------------------------------------------------------------

func compare(results map[string]bool, keys []string, flaglist bool, comparefile *os.File) {
	// We compare files A and B

	newresA := []string{}
	newresB := []string{}
	differences := []string{}

	oracle := make(map[string]bool)
	resultScanner := bufio.NewScanner(comparefile)
	for resultScanner.Scan() {
		line := strings.Split(resultScanner.Text(), " ")

		if len(line) < 3 || line[0] != "FORMULA" {
			log.Panicf("parsing %s: found %s\n", comparefile.Name(), ZipString(line, "", "", " "))
			os.Exit(1)
			return
		}
		if line[2] == "TRUE" {
			oracle[line[1]] = true
			continue
		}
		if line[2] == "FALSE" {
			oracle[line[1]] = false
			continue
		}
	}

	fmt.Printf("file %s (B): %d distinct results\n", comparefile.Name(), len(results))

	for _, q := range keys {
		res := results[q]
		answ, ok := oracle[q]
		if !ok {
			newresA = append(newresA, q)
			continue
		}
		if answ != res {
			differences = append(differences, q)
			if answ {
				fmt.Println("FORMULA " + q + " FALSE TRUE")
			}
			fmt.Println("FORMULA " + q + " TRUE FALSE")
		}
	}

	for q := range oracle {
		_, ok := results[q]
		if !ok {
			newresB = append(newresB, q)
		}
	}
	sort.Strings(newresB)

	fmt.Printf("%d differences\n", len(differences))
	fmt.Printf("file (A) %d unique results\n", len(newresA))
	fmt.Printf("file (B) %d unique results\n", len(newresB))

	if flaglist {
		for _, q := range newresA {
			fmt.Printf("A > %s\n", q)
		}
		for _, q := range newresB {
			fmt.Printf("B < %s\n", q)
		}
	}

}

// ----------------------------------------------------------------------

func ZipString(a []string, start, end, sep string) string {
	res := start
	for k, aa := range a {
		if k != 0 {
			res += sep
		}
		res += aa
	}
	return res + end
}
