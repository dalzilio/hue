<!-- PROJECT LOGO -->
<br />
<p align="center">
  <a href="https://github.com/dalzilio/hue">
    <img src="docs/hue.png" alt="Logo" width="256" height="256">
  </a>

  <p align="center">
    A collection of tools for exploring High-Level Petri nets !
    <!-- <br />
    <a href="https://github.com/dalzilio/mcc#features"><strong>see what's new »</strong></a>
    <br /> -->
    <!-- <a href="https://github.com/dalzilio/mcc">View Demo</a> -->
  </p>
</p>

## About

Hue is an API and a collection of tools for checking reachability properties
directly on  High-Level Petri nets. It accepts models in
[PNML](http://www.pnml.org/) format and the same formula language used in the
[Model-Checking Contest](https://mcc.lip6.fr/) (MCC).

Example of models and formulas can be found in the `benchmarks` folder, which
contains colored models extracted from the 2022 edition of the MCC.

An oracle for the reachability competition of MCC 2023 can be found in the source repository of the `htest` command, at: [`./htest/oracle_reach2023.txt`](https://raw.githubusercontent.com/dalzilio/hue/main/cmd/htest/oracle_reach2023.txt). We also provide an oracle for the 2022 edition, [`./htest/oracle_reach2022.txt`](https://raw.githubusercontent.com/dalzilio/hue/main/cmd/htest/oracle_reach2022.txt), compiled from information provided by [Yann Thierry-Mieg](https://github.com/yanntm), see [here](https://github.com/yanntm/pnmcc-models-2022).

## Overview

*uwalk* is a prototype stepper (a simulation tool) that performs random,
parallel exploration over a high-level net without unfolding it. It can also
test for ReachabilityCardinality (option `-r`) or ReachabilityFireability
(option `-f`) properties on the fly.

*hsimplify* is a command that applies elementary simplifications over
reachability formulas and can also, in some cases, find tautologies.

*htest* is a tool that can compare verification results, for the Reachability
competition, with oracles files computed on the 2022 and 2023 edition of the
MCC.

*horacle* is an internal set of conversion utilities that were used to compile
the oracle file used by `htest` for the 2023 edition of the MCC.
