package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jmcvetta/randutil"
	"github.com/voxelbrain/goptions"
)

// ShortLength - The number of characters for a short fortune
const ShortLength int = 160

// FileInfo - Information for a Fortune file
type FileInfo struct {
	filePath      string
	fortunes      []int
	shortFortunes []int
	longFortunes  []int
	weightShort   int
	weightLong    int
	weight        int
}

func (f *FileInfo) processFortunes(fortunes []string) {
	for index, fortune := range fortunes {
		if len(fortune) >= ShortLength {
			f.longFortunes = append(f.longFortunes, index)
			f.weightLong++
		}
		if len(fortune) < ShortLength {
			f.shortFortunes = append(f.shortFortunes, index)
			f.weightLong++
		}
		f.fortunes = append(f.fortunes, index)
		f.weight++
	}
}

func (f *FileInfo) setPath(path string) {
	f.filePath = path
}

func (f FileInfo) getFortune(short bool, long bool) (string, error) {
	content, err := ioutil.ReadFile(f.filePath)
	if err != nil {
		log.Fatal(err)
	}
	text := string(content)
	fortunes := splitFile(&text)
	if short {
		index, err := randutil.IntRange(0, len(f.shortFortunes))
		return fortunes[index], err
	} else if long {
		index, err := randutil.IntRange(0, len(f.longFortunes))
		return fortunes[index], err
	} else {
		index, err := randutil.IntRange(0, len(f.fortunes))
		return fortunes[index], err
	}
}

func (f FileInfo) getWeight(short bool, long bool) int {
	if short {
		return f.weightShort
	} else if long {
		return f.weightLong
	} else {
		return f.weight
	}
}

func main() {
	options := struct {
		Folder string        `goptions:"-f, --Folder, description='Fortunes folder'"`
		Short  bool          `goptions:"-s, --short, description='Short fortunes only'"`
		Long   bool          `goptions:"-l, --long, description='Long fortunes only'"`
		Wait   time.Duration `goptions:"-w, --wait, description='Delay before displaying fortune'"`
		List   bool          `goptions:"-t, --list, description='Print out the fortune files in the folder'"`
		Help   goptions.Help `goptions:"-h, --help, description='Show this help'"`
	}{ // Default values goes here
		Wait:   0 * time.Second,
		Folder: os.Getenv("FORTUNES_FOLDER"),
	}
	goptions.ParseAndFail(&options)
	files := readAllFiles(options.Folder, &options.Short, &options.Long)
	if options.List {
		fortuneFiles := listFiles(options.Folder)
		for _, file := range fortuneFiles {
			fmt.Printf("%s\n", file)
		}
	} else {
		fortune := selectFortune(&files, &options.Short, &options.Long)
		displayFortune(&fortune, &options.Wait)
	}
}

func listFiles(folder string) []string {
	var files []string
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if isFortuneFile(path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return files
}

// Read all files in a directory
func readAllFiles(folder string, short *bool, long *bool) []FileInfo {
	var files []FileInfo
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if isFortuneFile(path) {
			fortunes := readFile(path, short, long)
			var f FileInfo
			f.setPath(path)
			f.processFortunes(fortunes)
			files = append(files, f)
		}
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
			Weight: file.getWeight(*short, *long),
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
	var filteredText []string
	if *short {
		for _, item := range splitText {
			if len(item) <= ShortLength {
				filteredText = append(filteredText, item)
			}
		}
		return filteredText
	}
	if *long {
		for _, item := range splitText {
			if len(item) > ShortLength {
				filteredText = append(filteredText, item)
			}
		}
		return filteredText
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

func isFortuneFile(filename string) bool {
	var isFortune bool = true
	/* list of "illegal" suffixes" */
	var suffixList = []string{
		"dat", "pos", "c", "h", "p", "i", "f",
		"pas", "ftn", "ins.c", "ins,pas",
		"ins.ftn", "sml"}
	for _, suf := range suffixList {
		if strings.HasSuffix(filename, suf) {
			isFortune = false
			break
		}
	}
	return isFortune
}
