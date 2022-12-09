# RiskScan.go

My first go program, made to learn the language.

## To run the program
`go run .\riskScan.go --dir <directory to scan>  --out <output_file.json>`

## Last minute change
- I removed the use of Walk as I feel it was too constraining in the end. By using actual recursion, I have a point to add multithreading if needed.

## What is missing
- To make it perfectly safe and production-ready, this project desperately needs some unit tests.
- Some optimisations can probably be made as some of the implementations are pretty naive.
- Error-handling is also incomplete.
- I read about Go routines but didn't implement them yet. It would make the recursion more efficient.
- The project could benefit  from some cleanup and organisation by splitting in multiple files.