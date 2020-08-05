---
layout: page
nav_order: 7
---

# Testing
{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

tl;dr: `make before_commit` will run most of the checks by magic.

The following tests can be run to test your code in `insights-content-service`.
Detailed information about each type of test is included in the corresponding
subsection: 

1. Unit tests: checks behaviour of all units in source code (methods, functions)

## Unit tests

Set of unit tests checks all units of source code. Additionally the code
coverage is computed and displayed. Code coverage is stored in a file
`coverage.out` and can be checked by a script named `check_coverage.sh`.

To run unit tests use the following command:

`make test`

## Coverage reports

To make a coverage report you need to start `make cover`. It will run the unit
tests, generate a coverage report and open a web browser that allows to inspect
the results of the tests and its coverage.
