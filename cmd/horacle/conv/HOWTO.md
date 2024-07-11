# How to generate oracle files and other infos

## How to generate file reachXXX

* uncompress the GlobalSummary csv file from the MCC website

* use Excel to keep only the results for RC and RF, for only one tool. Remove
  unimportant columns

* Look for lines with only `? 0` at end of line; these are models with no
  results at all. Change them to `???????????????? 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
  0`

* Change `\?` into `?` and then `\?(\d)` into `? $1`, for unknown verdicts at
  the end of a line.

* use `export LC_COLLATE="C"` and then `sort -o` the reach file to have a sorted
  result with uppercase letters always before lower case.

* __Remark:__ there are formulas for model DotAndBoxes in 2020 but no summary
  results.

## Other modifications

Some verdicts may have been corrected after the end of the contest, after manual
reviews from the tool developers. Curated answers are found in the GitHub
projects maintained by Yann Thierry-Mieg, see e.g.
<https://github.com/yanntm/pnmcc-models-2022/blob/master/install_inputs.sh> for
the 2022 data.

### Corrections for 2020, see <https://github.com/yanntm/pnmcc-models-2020/blob/master/install_inputs.sh>

* Due to ITS-Tools in 2020 believing NuPN implies one-safe, there was some
errors on the RERS17* examinations. We keep the expected results reported in the
MCC website when Tapaal answers.

* Oracles for instances of model Sudoku-COL are deemed unreliable.

### Corrections for 2021, see <https://github.com/yanntm/pnmcc-models-2021/blob/master/install_inputs.sh>

No corrections found

### Corrections for 2022, see <https://github.com/yanntm/pnmcc-models-2022/blob/master/install_inputs.sh#L52-L82>

```
sed -i -e "s/StigmergyCommit-PT-11a-ReachabilityCardinality-06 TRUE/StigmergyCommit-PT-11a-ReachabilityCardinality-06 FALSE/" oracle_reach2022.txt
sed -i -e "s/StigmergyCommit-PT-11a-ReachabilityCardinality-13 FALSE/StigmergyCommit-PT-11a-ReachabilityCardinality-13 TRUE/" oracle_reach2022.txt
sed -i -e "s/StigmergyCommit-PT-11a-ReachabilityCardinality-14 TRUE/StigmergyCommit-PT-11a-ReachabilityCardinality-14 FALSE/" oracle_reach2022.txt
sed -i -e "s/StigmergyCommit-PT-11a-ReachabilityCardinality-15 FALSE/StigmergyCommit-PT-11a-ReachabilityCardinality-15 TRUE/" oracle_reach2022.txt
```

## Files format

All data files are space separated values. We use both csv and txt for the file
extensions.

File `reachXXXX.txt` are csv with 18 values: `complete instance name`
(Model-type-instance, where type is either PT or COL) ; `expected verdicts`
(with `?` for unknown values) ; and a `difficulty level`, meaning the sum of the
confidence values of all tools that have answered the query (so one for each of
the 16 instances).

The difficulty level gives an indication on the difficulty of the query (the
less the highest).

We build the oracle file, `oracle_reachXXXX.txt`, from the `reachXXXX.txt` file
using the conv tool. The file has one line by query with its verdict and
difficulty level. The file is stored together in directory `horacle/htest`
because it is needed to build the htest tool.

The `formXXXX.csv` file has 3 columns: `complete query name`, concatenation of
the instance name and the query identifier (between 00 and 15) ; its top=level
`modality`, either AG or EF; and its `size`, in number of operators and atomic
propositions. It is generated from the `models+formulas` archive found in the
MCC website (after decompression), using the forms tool.

File `iscexXXX.csv` lists for every query if it is an invariant (INV) ; a
formula that can be decided by a counter-example (CEX) ; or if it is undecided
(UNKNOWN). The last case stem from verdicts that have no consensus.
