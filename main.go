package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

// TODO(fabio): Use path lib to concat dir path to file name so it doesn't matter if  ends with / or not.
// TODO(fabio): Recursively enter directories. If Roam Research title has slash, it exports as dir with file.

const TITLE_PREFIX = "#+TITLE:"
const MAX_BULLET_NESTING = 20 // The max number of indents we will properly convert
var ErrIsDir = errors.New("is directory")

func main() {
	ROAM_DIR_PTR := flag.String("d", "", "Directory containing your Roam Research export")
	flag.Parse()
	ROAM_DIR := *ROAM_DIR_PTR
	if ROAM_DIR == "" {
		log.Fatalf("Please make sure you use the -d flag to specify the directory containing your Roam Research exported files")
	}

	files, err := ioutil.ReadDir(ROAM_DIR)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		fmt.Println(f.Name())
		if strings.HasSuffix(f.Name(), ".db") {
			continue // Ignore it
		}
		path := fmt.Sprintf("%s%s", ROAM_DIR, f.Name())
		record, err := NewRecord(path)
		if err != nil {
			if err == ErrIsDir {
				continue
			}
			log.Fatal(err)
		}
		if !record.HasTitle() {
			title := fmt.Sprintf("%s %s\n", TITLE_PREFIX, f.Name()[:len(f.Name())-3])
			record.Prepend(title)
			if err != nil {
				log.Fatal(err)
			}
		}
		record.FixBidirectionalLinks()
		names := record.FindAllBidirectionalLinkNames()
		record.FormatDates(names)
		record.FixTaskKeywords()
		record.FormatBullets()
		record.RemoveRoamStyling()
		record.UnderscoreLinkedFileNames(names)
		record.ConvertHashTagsToBidirectionalLinks()
		record.Save()

		// Rename file to underscore version and change suffix to .org
		oldPath := path
		newPath := strings.ReplaceAll(strings.Replace(path, ".md", ".org", 1), " ", "_")
		err = os.Rename(oldPath, newPath)
		if err != nil {
			log.Fatal(err)
		}
	}
}

type Record struct {
	Filename string
	Contents string
}

func NewRecord(filename string) (*Record, error) {
	if stat, err := os.Stat(filename); err == nil && stat.IsDir() {
		return nil, ErrIsDir
	}
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Record{
		Filename: filename,
		Contents: string(content),
	}, nil
}

func (r *Record) ConvertHashTagsToBidirectionalLinks() {
	// Replace hash tags with links
	regex := `\s#([^[+[][^\s]+)`
	replacement := ` [[file:$1.org][$1]]`
	re := regexp.MustCompile(regex)
	r.Contents = re.ReplaceAllString(r.Contents, replacement)

	// Remove hash tag in front of links
	regex = `#(\[\[file:)`
	replacement = `$1`
	re = regexp.MustCompile(regex)
	r.Contents = re.ReplaceAllString(r.Contents, replacement)
}

func (r *Record) FormatDates(names []string) {
	for _, n := range names {
		// Remove th, rd, st, nd,
		possibleDateStr := removeDaySuffixes(n)
		t, err := time.Parse("January 2 2006", possibleDateStr)
		if err != nil {
			continue // Skip any non-dates
		}
		regex := fmt.Sprintf(`\[\[file:%s.org\]\[%s\]\]`, n, n)
		replacement := t.Format("<2006-01-02 Mon>")
		var re = regexp.MustCompile(regex)
		r.Contents = re.ReplaceAllString(r.Contents, replacement)
	}
}

func removeDaySuffixes(possibleDateStr string) string {
	firstRegex := `st,`
	replacement := ``
	re := regexp.MustCompile(firstRegex)
	modifiedDateStr := re.ReplaceAllString(possibleDateStr, replacement)

	secondRegex := `nd,`
	re = regexp.MustCompile(secondRegex)
	modifiedDateStr = re.ReplaceAllString(modifiedDateStr, replacement)

	thirdRegex := `rd,`
	re = regexp.MustCompile(thirdRegex)
	modifiedDateStr = re.ReplaceAllString(modifiedDateStr, replacement)

	fourthRegex := `th,`
	re = regexp.MustCompile(fourthRegex)
	modifiedDateStr = re.ReplaceAllString(modifiedDateStr, replacement)

	return modifiedDateStr
}

func (r *Record) FixTaskKeywords() {
	// Fix DONE tasks
	regex := `{{\[\[file:DONE.org\]\[DONE\]\]}}`
	replacement := `DONE`
	re := regexp.MustCompile(regex)
	r.Contents = re.ReplaceAllString(r.Contents, replacement)

	// Fix TODO tasks
	regex = `{{\[\[file:TODO.org\]\[TODO\]\]}}`
	replacement = `TODO`
	re = regexp.MustCompile(regex)
	r.Contents = re.ReplaceAllString(r.Contents, replacement)

	// Move scheduled timestamps to next line with SCHEDULED: keyword
	regex = `(.*)TODO(.*)<(.*)>(.*)`
	replacement = `$1 TODO$2 $4
$1 SCHEDULED: <$3>`
	re = regexp.MustCompile(regex)
	r.Contents = re.ReplaceAllString(r.Contents, replacement)

	// Get rid of leading `-` in front of SCHEDULED keyword
	regex = `(-  )SCHEDULED:`
	replacement = `   SCHEDULED:`
	re = regexp.MustCompile(regex)
	r.Contents = re.ReplaceAllString(r.Contents, replacement)
}

func (r *Record) FormatBullets() {
	for i := MAX_BULLET_NESTING; i >= 0; i-- {
		regex := fmt.Sprintf("%*s", i*4, "- ")
		replacement := bullets(i)
		var re = regexp.MustCompile(regex)
		r.Contents = re.ReplaceAllString(r.Contents, replacement)
	}
	regex := `  \*\*`
	replacement := `**`
	var re = regexp.MustCompile(regex)
	r.Contents = re.ReplaceAllString(r.Contents, replacement)
}

func bullets(num int) string {
	bullets := ""
	for i := 0; i <= num; i++ {
		bullets = bullets + "*"
	}
	return bullets + " "
}

func (r *Record) RemoveRoamStyling() {
	regex := `### `
	replacement := ``
	var re = regexp.MustCompile(regex)
	r.Contents = re.ReplaceAllString(r.Contents, replacement)

	regex = `## `
	replacement = ``
	re = regexp.MustCompile(regex)
	r.Contents = re.ReplaceAllString(r.Contents, replacement)
}

func (r *Record) FixBidirectionalLinks() {
	regex := `\[\[([^file:][a-zA-Z0-9 !&,-.()?':]*)\]\]`
	replacement := `[[file:$1.org][$1]]`
	var re = regexp.MustCompile(regex)
	r.Contents = re.ReplaceAllString(r.Contents, replacement)
}

func (r *Record) FindAllBidirectionalLinkNames() []string {
	regex := `\[\[file:([^\]]*)`
	var re = regexp.MustCompile(regex)
	matches := re.FindAllStringSubmatch(r.Contents, -1)
	names := []string{}
	for _, m := range matches {
		names = append(names, m[1][:len(m[1])-4])
	}
	return names
}

func (r *Record) UnderscoreLinkedFileNames(names []string) {
	for _, n := range names {
		underscoredName := strings.ReplaceAll(n, " ", "_")
		regex := fmt.Sprintf(`\[\[file:%s.org\]\[%s\]\]`, n, n)
		replacement := fmt.Sprintf(`[[file:%s.org][%s]]`, underscoredName, n)
		var re = regexp.MustCompile(regex)
		r.Contents = re.ReplaceAllString(r.Contents, replacement)
	}
}

func (r *Record) Save() {
	err := ioutil.WriteFile(r.Filename, []byte(r.Contents), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func (r *Record) HasTitle() bool {
	return strings.Contains(r.Contents, TITLE_PREFIX)
}

func (r *Record) Prepend(content string) {
	r.Contents = content + r.Contents
}