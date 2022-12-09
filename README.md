# RiskScan.go

My first go program, made to learn the language.

## To run the program
`go run .\riskScan.go --dir <directory to scan>  --out <output_file.json>`

## What is missing
- To make it perfectly safe and production-ready, this project desperately needs some unit tests.
- Some optimisations can probably be made as some of the implementations are pretty naive.
- Error-handling is also incomplete.
- I read about Go routines but didn't implement them yet. It would make the Walk more efficient.