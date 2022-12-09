package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

/*
===============
Part one: Constants, return types, variables
===============
*/

const (
	hoursInWeek = 24 * 7
	maxRisk     = 1.00
	minRisk     = 0.00
	dirArg      = "--dir"
	outArg      = "--out"
	maxResults  = 10
)

// A file path and its associated risk
type FileResult struct {
	Path string
	Risk float64
}

// A directory, potentially containing files with risks
type DirResult struct {
	Dir     string
	Results []FileResult
}

var extensionRiskMap map[string]float64

/*
===============
Part two: Risk assessment rules
===============
*/

// Checks how much risk to apply based on the file extension
func assessExtension(path string) float64 {

	// Extract the extension
	extension := filepath.Ext(path)

	risk, ok := extensionRiskMap[extension]

	if ok {
		return risk
	}

	// Extension was not recognized: ignoring it
	return 0
}

// Makes sure the risk is within the defined bounds (0.0 - 1.0)
func checkRiskRange(risk float64) float64 {
	if risk > maxRisk {
		return maxRisk
	}
	if risk < minRisk {
		return minRisk
	}
	return risk
}

// Calculates the risk of a given file as a float between 0.0 (low risk) to 1.0 (high risk)
func assessFileRisk(path string, info fs.FileInfo) float64 {
	var risk float64 = 0.0

	// If the file size is larger than 1mb → Add 0.25
	if info.Size() > 1000000 {
		risk += 0.25
	}

	risk += assessExtension(path)

	// If the file was modified in the last week → Add 0.20
	timeLastWeek := time.Now().Add(time.Hour * -hoursInWeek)
	if info.ModTime().After(timeLastWeek) {
		risk += 0.20
	}

	return risk
}

// Rules on the folder name length of a file, only the first folder parent
func assessDirNameLength(path string) float64 {
	size := len(path)
	if size < 5 {
		return 0.25
	}
	if size > 15 {
		return -0.10
	}
	return 0.5
}

// Assess the risk of a directory
func assessDirRisk(path string, info fs.FileInfo, err error) []FileResult {

	var result []FileResult

	if nil == err {
		// If the file size is lower than 1 KB ignore it.
		if info.Size() > 1000 {
			// fmt.Printf("Assessing: %v\n", path)
			var fileResult FileResult
			fileResult.Path, _ = filepath.Abs(path)
			fullRisk := assessFileRisk(path, info) + assessDirNameLength(path)
			fileResult.Risk = checkRiskRange(fullRisk)
			result = append(result, fileResult)
		}
	}
	return result
}

/*
===============
Part three: Util functions
===============
*/

// Initializes a map with risk values for all the extensions we check. This should only be run once at the start of the application.
// TODO This could potentially be externalized to a json/csv file and read as config.
func initExtensionRiskMap() map[string]float64 {
	extensionValues := make(map[string]float64)

	// If the file has the extension [zip, tar] → Add 0.15
	extensionValues[".zip"] = 0.15
	extensionValues[".tar"] = 0.15

	// If the file is an image [png, jpeg] → Remove 0.20
	extensionValues[".png"] = -0.20
	extensionValues[".jpg"] = -0.20
	extensionValues[".jpeg"] = -0.20

	// If the extension is [csv, json] → Add 0.75
	extensionValues[".csv"] = 0.75
	extensionValues[".json"] = 0.75

	return extensionValues
}

// Reads the command line arguments.
// We need values for '--dir' and '--out'. Doesn't matter the order, ignore other args.
// There is probably a better way of doing this in a library somewhere but I don't know enough Go to know about it...
func readCommandLineArgs() map[string]string {
	args := os.Args[1:]
	size := len(args)

	var result map[string]string = make(map[string]string)

	for i := 0; i < size; i++ {
		// Get the root directory
		if dirArg == args[i] && i+1 < size {
			result[dirArg] = args[i+1]
			// Skip a step because we consumed the value already
			i++
		}

		// Get the output file name
		if outArg == args[i] && i+1 < size {
			result[outArg] = args[i+1]
			// Skip a step because we consumed the value already
			i++
		}
	}

	return result
}

// Finds the smallest risk and its index in an array of FileResult
func findSmallestRisk(results [maxResults]FileResult) (float64, int) {
	smallestRisk := maxRisk
	smallestRiskIndex := 0

	for i, r := range results {
		if r.Risk < smallestRisk {
			smallestRisk = r.Risk
			smallestRiskIndex = i
		}

	}

	return smallestRisk, smallestRiskIndex
}

// Reduce the number of results to the maximum allowed
func trimDownResults(results []FileResult) []FileResult {

	if len(results) <= maxResults {
		return results
	}

	trimmedResults := [maxResults]FileResult{}

	// TODO this is definitely not the most efficient way to do this
	for i, result := range results {
		if i < maxResults {
			trimmedResults[i] = result
		} else {
			smallestRisk, smallestRiskIndex := findSmallestRisk(trimmedResults)
			if smallestRisk < result.Risk {
				// Replace the lowest risk with then new one
				trimmedResults[smallestRiskIndex] = result
			}
		}
	}

	return trimmedResults[:]
}

// Gets the directory from an absolute path
func getDir(path string) string {
	splitString := strings.Split(path, string(os.PathSeparator))
	everythingButTheFileName := splitString[:1]
	var dir string
	strings.Join(everythingButTheFileName, string(os.PathSeparator))
	return dir
}

// Write the DirResultStructure to the output file
func writeJsonToFile(outFile *os.File, data DirResult) {
	encoder := json.NewEncoder(outFile)
	encoder.SetIndent("", "    ")
	// fmt.Printf("Object before writing: %v\n", data)
	encoder.Encode(data)
}

/*
===============
 Part four: Main
===============
*/

func main() {

	// Init
	extensionRiskMap = initExtensionRiskMap()

	// Command line arguments without the program name
	args := readCommandLineArgs()

	rootDir, dirExists := args[dirArg]
	outFileName, outExists := args[outArg]

	if !dirExists || !outExists {
		fmt.Println("Both '--dir' and '--out' need to be set. Exiting.")
		return
	}

	var finalResult DirResult

	absoluteDir, _ := filepath.Abs(rootDir)
	finalResult.Dir = absoluteDir

	var currentResults []FileResult

	// We use Walk instead of WalkDir because we are going to need the FileInfo at every step.
	filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			// We assume here that once we change dir here we have seen all the files in the previous dir
			// So we save the results for the current directory, but only the 10 most risky
			for _, r := range trimDownResults(currentResults) {
				finalResult.Results = append(finalResult.Results, r)
			}
			currentResults = nil
		} else {
			dirResult := assessDirRisk(path, info, err)
			for _, r := range dirResult {
				currentResults = append(currentResults, r)
			}
		}
		return nil
	})

	// Final save after ending the walk
	for _, r := range trimDownResults(currentResults) {
		finalResult.Results = append(finalResult.Results, r)
	}

	// TODO probably better to check if file exists or not
	outFile, fileOpenErr := os.OpenFile(outFileName, os.O_CREATE, os.ModePerm)
	defer outFile.Close()

	if nil != fileOpenErr {
		fmt.Printf("Error while opening the output file: %v\n", fileOpenErr)
		return
	}

	writeJsonToFile(outFile, finalResult)

}
