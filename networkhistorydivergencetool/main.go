package main

import (
	"archive/zip"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	shell "github.com/ipfs/go-ipfs-api"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Define the original Response struct to parse JSON
type Response struct {
	Segments []struct {
		FromHeight               string `json:"fromHeight"`
		ToHeight                 string `json:"toHeight"`
		HistorySegmentID         string `json:"historySegmentId"`
		PreviousHistorySegmentID string `json:"previousHistorySegmentId"`
		DatabaseVersion          string `json:"databaseVersion"`
		ChainID                  string `json:"chainId"`
	} `json:"segments"`
}

// Define the modified struct with int type for ToHeight and FromHeight
type Segments struct {
	Segments []Segment `json:"segments"`
}

type Segment struct {
	FromHeight               int    `json:"fromHeight"`
	ToHeight                 int    `json:"toHeight"`
	HistorySegmentID         string `json:"historySegmentId"`
	PreviousHistorySegmentID string `json:"previousHistorySegmentId"`
	DatabaseVersion          string `json:"databaseVersion"`
	ChainID                  string `json:"chainId"`
}

var ipfsShell *shell.Shell

const defaultIPFSHost = "localhost:7001"
const defaultMinHeight = 0
const defaultMaxHeight = 1_000_000_000

func main() {

	// Check if the correct number of arguments is provided
	if len(os.Args) < 3 || len(os.Args) > 6 {
		printHelp()
		os.Exit(1)
	}

	// Extract the server addresses
	truthServer := os.Args[1]
	toCompareServer := os.Args[2]

	// Set default values for ipfsHost, minHeight, and maxHeight
	ipfsHost := defaultIPFSHost
	minHeight := defaultMinHeight
	maxHeight := defaultMaxHeight

	// Parse optional parameters if provided
	if len(os.Args) > 3 {
		ipfsHost = os.Args[3]
	}
	if len(os.Args) > 4 {
		minHeight = parseInt(os.Args[4])
	}
	if len(os.Args) > 5 {
		maxHeight = parseInt(os.Args[5])
	}

	// Print the provided values
	fmt.Printf("Truth Server: %s\n", truthServer)
	fmt.Printf("To Compare Server: %s\n", toCompareServer)
	fmt.Printf("IPFS Host: %s\n", ipfsHost)
	fmt.Printf("Minimum Height: %d\n", minHeight)
	fmt.Printf("Maximum Height: %d\n", maxHeight)

	ipfsShell = shell.NewShell(ipfsHost)

	theTruthResponse := getSegments(truthServer)
	toCompareResponse := getSegments(toCompareServer)

	toComparetoHeightToSegment := map[int]Segment{}
	for _, segment := range toCompareResponse.Segments {
		toComparetoHeightToSegment[segment.ToHeight] = segment
	}

	theTruthToHeightToSegment := map[int]Segment{}
	for _, segment := range theTruthResponse.Segments {
		theTruthToHeightToSegment[segment.ToHeight] = segment
	}

	var truthSegments []Segment
	var toCompareSegments []Segment

	for toHeight, segment := range toComparetoHeightToSegment {
		if toHeight > maxHeight {
			continue
		}

		if toHeight < minHeight {
			continue
		}

		toCompareSegments = append(toCompareSegments, segment)
		if truthSegment, ok := theTruthToHeightToSegment[toHeight]; ok {
			truthSegments = append(truthSegments, truthSegment)
		} else {
			log.Panicf("no truth segment found for height %d", toHeight)
		}

	}

	sort.SliceStable(toCompareSegments, func(j, k int) bool {
		return toCompareSegments[j].ToHeight < toCompareSegments[k].ToHeight
	})

	sort.SliceStable(truthSegments, func(j, k int) bool {
		return truthSegments[j].ToHeight < truthSegments[k].ToHeight
	})

	for i := 0; i < len(truthSegments); i++ {
		truth := truthSegments[i]
		compare := toCompareSegments[i]

		if truth.HistorySegmentID != compare.HistorySegmentID {
			fmt.Printf("First Different HistorySegmentID values for FromHeight %d ToHeight %d  Truth:%s  Compare:%s:\n", truth.FromHeight, truth.ToHeight,
				truth.HistorySegmentID, compare.HistorySegmentID)

			truthSegmentDir, err := sourceHistorySegment(truth.HistorySegmentID)
			if err != nil {
				log.Panicf("failed to source truth history segment: %s", err)
			}

			compareSegmentDir, err := sourceHistorySegment(compare.HistorySegmentID)
			if err != nil {
				log.Panicf("failed to source compare history segment: %s", err)
			}

			mismatchedFiles, err := compareDirectories(truthSegmentDir, compareSegmentDir)
			if err != nil {
				if errors.Is(err, ERR_DIFF_FILE_NUM) {
					fmt.Printf("Failed to compare segments, different number of files in dir")
					break
				} else {
					log.Panicf("failed to compare directories: %s", err)
				}
			}

			if len(mismatchedFiles) == 0 {
				fmt.Printf("No mismatched files found, contents are identical but different history segment IDs\n")
			}

			for _, file := range mismatchedFiles {
				absTruthDir, err := filepath.Abs(truthSegmentDir)
				if err != nil {
					log.Panicf("failed to get abs path: %s", err)
				}

				absCompareDir, err := filepath.Abs(compareSegmentDir)
				if err != nil {
					log.Panicf("failed to get compare path: %s", err)
				}

				fmt.Printf("\nMISMATCHED DATA: %s at fromHeight %d, toHeight %d, to see differences: diff %s %s  \n", file, truth.FromHeight,
					truth.ToHeight, filepath.Join(absTruthDir, file), filepath.Join(absCompareDir, file))
			}

			break
		}
	}

}

func printHelp() {
	fmt.Println("Usage: go run main.go <server1> <server2> [ipfsHost] [minHeight] [maxHeight]")
	fmt.Println("Parameters:")
	fmt.Println("<server1>\t\tThe first server address.")
	fmt.Println("<server2>\t\tThe second server address.")
	fmt.Println("[ipfsHost]\t\t(Optional) The IPFS host. Default: " + defaultIPFSHost)
	fmt.Println("[minHeight]\t\t(Optional) The minimum height. Default: " +
		strconv.Itoa(defaultMinHeight))
	fmt.Println("[maxHeight]\t\t(Optional) The maximum height. Default: " + strconv.Itoa(defaultMaxHeight))
}

func parseInt(arg string) int {
	value := 0
	_, err := fmt.Sscanf(arg, "%d", &value)
	if err != nil {
		fmt.Printf("Failed to parse integer from %s. Using default value.\n", arg)
	}
	return value
}

var ERR_DIFF_FILE_NUM = errors.New("Different number of files")

func compareDirectories(dir1, dir2 string) ([]string, error) {
	var nonMatchingFiles []string

	// Read the file names from the first directory
	files1, err := getFileNames(dir1)
	if err != nil {
		return nil, err
	}

	// Read the file names from the second directory
	files2, err := getFileNames(dir2)
	if err != nil {
		return nil, err
	}

	// Compare the number of files
	if len(files1) != len(files2) {
		return nil, ERR_DIFF_FILE_NUM
	}

	// Compare the file names and hashes
	for i := range files1 {
		if files1[i] != files2[i] {
			nonMatchingFiles = append(nonMatchingFiles, files1[i])
			continue
		}

		// Read the contents of the file from the first directory
		filePath1 := filepath.Join(dir1, files1[i])
		content1, err := ioutil.ReadFile(filePath1)
		if err != nil {
			return nil, err
		}

		// Read the contents of the file from the second directory
		filePath2 := filepath.Join(dir2, files2[i])
		content2, err := ioutil.ReadFile(filePath2)
		if err != nil {
			return nil, err
		}

		// Calculate the MD5 hash of the contents
		hash1 := md5.Sum(content1)
		hash2 := md5.Sum(content2)

		// Compare the hashes
		if hash1 != hash2 {
			nonMatchingFiles = append(nonMatchingFiles, files1[i])
		}
	}

	return nonMatchingFiles, nil
}

func getFileNames(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get the relative file path
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		files = append(files, relPath)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func sourceHistorySegment(segmentID string) (string, error) {

	err := os.MkdirAll(fmt.Sprintf("./segments/%s", segmentID), os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("Error creating segment directory: %w\n", err)
	}

	zipFileName := fmt.Sprintf("./segments/%s.zip", segmentID)
	err = ipfsShell.Get(segmentID, zipFileName)
	if err != nil {
		return "", fmt.Errorf("Error sourcing history segment from IPFS: %v\n", err)
	}
	fmt.Printf("History segment %s sourced successfully.\n", segmentID)

	segmentDir := fmt.Sprintf("./segments/%s", segmentID)
	os.MkdirAll(segmentDir, os.ModePerm)
	UnzipSource(zipFileName, segmentDir)

	return segmentDir, nil
}

func UnzipSource(source, destination string) error {
	reader, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	destination, err = filepath.Abs(destination)
	if err != nil {
		return err
	}

	for _, f := range reader.File {
		err := unzipFile(f, destination)
		if err != nil {
			return err
		}
	}

	return nil
}

func unzipFile(f *zip.File, destination string) error {
	filePath := filepath.Join(destination, f.Name)
	if !strings.HasPrefix(filePath, filepath.Clean(destination)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", filePath)
	}

	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
			return err
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	destinationFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	zippedFile, err := f.Open()
	if err != nil {
		return err
	}
	defer zippedFile.Close()

	if _, err := io.Copy(destinationFile, zippedFile); err != nil {
		return err
	}
	return nil
}

func getSegments(server string) Segments {

	server += "/api/v2/networkhistory/segments"
	response, err := http.Get(server)
	if err != nil {
		log.Panicf("Error accessing %s: %v\n", server, err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Panicf("Error accessing %s: %s\n", server, response.Status)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Panicf("Error reading response body from %s: %v\n", server, err)
	}

	var parsedResponse Response
	err = json.Unmarshal(body, &parsedResponse)
	if err != nil {
		log.Panicf("Error parsing JSON from %s: %v\n", server, err)
	}

	modifiedResponse := convertToModifiedResponse(parsedResponse)
	return modifiedResponse
}

// Helper function to convert Response to Segments
func convertToModifiedResponse(response Response) Segments {
	modifiedResponse := Segments{}
	for _, segment := range response.Segments {
		fromHeight, _ := strconv.Atoi(segment.FromHeight)
		toHeight, _ := strconv.Atoi(segment.ToHeight)
		modifiedSegment := struct {
			FromHeight               int    `json:"fromHeight"`
			ToHeight                 int    `json:"toHeight"`
			HistorySegmentID         string `json:"historySegmentId"`
			PreviousHistorySegmentID string `json:"previousHistorySegmentId"`
			DatabaseVersion          string `json:"databaseVersion"`
			ChainID                  string `json:"chainId"`
		}{
			FromHeight:               fromHeight,
			ToHeight:                 toHeight,
			HistorySegmentID:         segment.HistorySegmentID,
			PreviousHistorySegmentID: segment.PreviousHistorySegmentID,
			DatabaseVersion:          segment.DatabaseVersion,
			ChainID:                  segment.ChainID,
		}
		modifiedResponse.Segments = append(modifiedResponse.Segments, modifiedSegment)
	}
	return modifiedResponse
}
