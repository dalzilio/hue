// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/dalzilio/hue/pkg/formula"
)

func main() {
	root := os.Args[1]
	fileInfo, err := os.ReadDir(root)
	if err != nil {
		log.Fatal(err)
	}
	for _, dir := range fileInfo {
		if !dir.IsDir() {
			continue
		}
		answer("ReachabilityFireability", root, dir.Name())
		answer("ReachabilityCardinality", root, dir.Name())
	}
}

func doSimplify(q formula.Query, compet string) formula.Query {
	result := make(chan formula.Query, 1)

	if compet == "ReachabilityFireability" {
		go func() {
			q.Formula, _ = formula.BddFireabilitySimplify(q.Formula)
			result <- q
		}()
	} else {
		go func() {
			q.Formula = formula.Simplify(q.Formula)
			result <- q
		}()
	}

	select {
	case <-time.After(60 * time.Second):
		return q
	case result := <-result:
		return result
	}
}

func answer(compet string, root string, dirfile string) {
	formulafile := filepath.Join(filepath.Join(root, dirfile), compet+".xml")
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
	for k, q := range queries {
		fmt.Printf("%s-%s-%02d %d\n", dirfile, compet, k, formula.Size(q.Formula))
	}
}

func checkIfEFandSimplify(compet string, root string, dirfile string) {
	formulafile := filepath.Join(filepath.Join(root, dirfile), compet+".xml")
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
	var triv string
	for k, q := range queries {
		// added a 60s timeout, for e.g. DrinkVendingMachine-PT-10
		q = doSimplify(q, compet)

		if q.IsTrivial() {
			rez := q.Formula.(formula.BooleanConstant)
			if rez {
				triv = "CSTTRUE"
			} else {
				triv = "CSTFALSE"
			}
		} else {
			triv = "COMPLEX"
		}
		if q.IsEF {
			fmt.Printf("%s-%s-%02d EF %s\n", dirfile, compet, k, triv)
		} else {
			fmt.Printf("%s-%s-%02d AG %s\n", dirfile, compet, k, triv)
		}
	}
}
