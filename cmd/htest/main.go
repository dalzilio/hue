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

	_ "embed"
)

//go:embed total.txt
var oracleFile string

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: htest file.txt")
		os.Exit(1)
		return
	}

	resultfile, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal("Error opening file:", err)
		os.Exit(1)
		return
	}

	results := make(map[string]bool)
	keys := []string{}
	resultScanner := bufio.NewScanner(resultfile)
	for resultScanner.Scan() {
		line := strings.Split(resultScanner.Text(), " ")

		if len(line) != 3 || line[0] != "FORMULA" {
			log.Fatal("parsing total.txt found " + ZipString(line, "", "", " ") + "\n")
			os.Exit(1)
			return
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

	fmt.Printf("%d distinct results\n", len(results))

	oracle := make(map[string]bool)
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
		panic("parsing total.txt found " + ZipString(line, "", "", " ") + "\n")
	}

	fmt.Printf("%d results in oracle\n", len(oracle))

	countnew := 0
	counterrors := 0
	sort.Strings(keys)

	for _, q := range keys {
		res := results[q]
		answ, ok := oracle[q]
		if !ok {
			countnew++
			continue
		}
		if answ != res {
			counterrors++
			if answ {
				fmt.Println("FORMULA " + q + " TRUE")
			}
			fmt.Println("FORMULA " + q + " FALSE")
		}
	}

	fmt.Printf("%d new results\n", countnew)
	fmt.Printf("%d differences\n", counterrors)

}

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
