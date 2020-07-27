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

const SHORT_LENGTH int = 160

type FileInfo struct {
	file_path string
	fortunes  int
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
	files := read_all_files(options.Folder)  // Add options as flags (short, long)
	fortune := select_fortune(&files)        // Add options as flags (short, long)
	display_fortune(&fortune, &options.Wait) // Add options as flags (wait)
}

// Read all files in a directory
func read_all_files(folder string) []FileInfo {
	var files []FileInfo
	// For file in folder
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		// read file
		if info.IsDir() {
			return nil
		}
		fortunes := read_file(path) // Add options as flags (short, long)
		files = append(files, FileInfo{path, len(fortunes)})
		return nil
	})
	if err != nil {
		panic(err)
	}
	return files
}

// Split file into fortunes
func split_file(file *string) []string {
	regex := regexp.MustCompile(`[\r\n]+%[\r\n]+`)
	return regex.Split(*file, -1) // second arg -1 means no limits for the number of substrings
}

// Print format
func display_fortune(fortune *string, wait *time.Duration) {
	time.Sleep(*wait)
	fmt.Println(*fortune)
}

// Select a random fortune
func select_fortune(files *[]FileInfo) string {
	var choices []randutil.Choice
	for _, file := range *files {
		choices = append(choices, randutil.Choice{
			Weight: file.fortunes,
			Item:   file.file_path})
	}
	result, err := randutil.WeightedChoice(choices)
	if err != nil {
		panic(err)
	} else {
		f := read_file(fmt.Sprintf("%s", result.Item))
		index, err := randutil.IntRange(0, len(f))
		if err != nil {
			panic(err)
		}
		return f[index]
	}
}

func read_file(file_path string) []string {
	// Read entire file content, giving us little control but
	// making it very simple. No need to close the file.
	content, err := ioutil.ReadFile(file_path)
	if err != nil {
		log.Fatal(err)
	}
	// Convert []byte to string and print to screen
	text := string(content)
	return split_file(&text)
}

func open_folder(folder string) []string {
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
