package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/jmcvetta/randutil"
	"github.com/voxelbrain/goptions"
)

// ShortLength - The number of characters for a short fortune
const ShortLength int = 160

// FileInfo - Information for a Fortune file
type FileInfo struct {
	filePath string
	fortunes int
}

func main() {
	options := struct {
		Folder string        `goptions:"-f, --Folder, description='Fortunes folder'"`
		Short  bool          `goptions:"-s, --short, description='Short fortunes only'"`
		Long   bool          `goptions:"-l, --long, description='Long fortunes only'"`
		Wait   time.Duration `goptions:"-w, --wait, description='Delay before displaying fortune'"`
		Help   goptions.Help `goptions:"-h, --help, description='Show this help'"`
	}{ // Default values goes here
		Wait:   0 * time.Second,
		Folder: os.Getenv("FORTUNES_FOLDER"),
	}
	goptions.ParseAndFail(&options)
	files := readAllFiles(options.Folder, &options.Short, &options.Long) // Add options as flags (short, long)
	fortune := selectFortune(&files, &options.Short, &options.Long)      // Add options as flags (short, long)
	displayFortune(&fortune, &options.Wait)                              // Add options as flags (wait)
}

// Read all files in a directory
func readAllFiles(folder string, short *bool, long *bool) []FileInfo {
	var files []FileInfo
	// For file in folder
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		// read file
		if info.IsDir() {
			return nil
		}
		fortunes := readFile(path, short, long) // Add options as flags (short, long)
		files = append(files, FileInfo{path, len(fortunes)})
		return nil
	})
	if err != nil {
		panic(err)
	}
	return files
}

// Split file into fortunes
func splitFile(file *string) []string {
	regex := regexp.MustCompile(`[\r\n]+%[\r\n]+`)
	return regex.Split(*file, -1) // second arg -1 means no limits for the number of substrings
}

// Print format
func displayFortune(fortune *string, wait *time.Duration) {
	time.Sleep(*wait)
	fmt.Println(*fortune)
}

// Select a random fortune
func selectFortune(files *[]FileInfo, short *bool, long *bool) string {
	var choices []randutil.Choice
	for _, file := range *files {
		choices = append(choices, randutil.Choice{
			Weight: file.fortunes,
			Item:   file.filePath})
	}
	result, err := randutil.WeightedChoice(choices)
	if err != nil {
		panic(err)
	} else {
		f := readFile(fmt.Sprintf("%s", result.Item), short, long)
		index, err := randutil.IntRange(0, len(f))
		if err != nil {
			panic(err)
		}
		return f[index]
	}
}

func readFile(filePath string, short *bool, long *bool) []string {
	// Read entire file content, giving us little control but
	// making it very simple. No need to close the file.
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	// Convert []byte to string
	text := string(content)
	splitText := splitFile(&text)
	var returnedText []string
	if *short {
		for _, item := range splitText {
			if len(item) <= ShortLength {
				returnedText = append(returnedText, item)
			}
		}
		return returnedText
	}
	if *long {
		for _, item := range splitText {
			if len(item) > ShortLength {
				returnedText = append(returnedText, item)
			}
		}
		return returnedText
	}
	return splitText
}

func openFolder(folder string) []string {
	var files []string
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		fmt.Println(file)
	}
	return files
}
