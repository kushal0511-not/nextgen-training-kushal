package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	"github.com/nextgen-training-kushal/Day-11/trie"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

type Report struct {
	TotalWords      int          `json:"total_words"`
	MisspelledWords int          `json:"misspelled_words"`
	Corrections     []Correction `json:"corrections"`
}

type Correction struct {
	Word        string   `json:"word"`
	Suggestions []string `json:"suggestions"`
}

type Job struct {
	Word string
}

type Result struct {
	Correction Correction
}

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not create CPU profile: %v\n", err)
			return
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintf(os.Stderr, "could not start CPU profile: %v\n", err)
			return
		}
		defer pprof.StopCPUProfile()
	}

	start := time.Now()
	root := trie.NewTrieNode()
	f, err := os.Open("words.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	fmt.Println("Indexing words from words.txt...")
	scanner := bufio.NewScanner(f)
	var allWords []string
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" {
			root.Insert(word)
			// Limit sample for demo to 5000 words due to EditDistance complexity
			if len(allWords) < 500000 {
				allWords = append(allWords, word)
			}
		}
	}
	totalTime := time.Since(start)
	fmt.Printf("Total indexing time: %s\n", totalTime)
	fmt.Println("Enter word for auto completion")
	var userInput string
	fmt.Scanln(&userInput)
	strs, _ := root.AutoComplete(userInput, 10)
	fmt.Println("Suggestions for", userInput, "are:", strings.Join(strs, ", "))
	in, err := os.Open("input.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer in.Close()

	var poolSize = runtime.GOMAXPROCS(0)
	jobs := make(chan Job, 100)
	results := make(chan Result, 100)
	var wg sync.WaitGroup
	fmt.Printf("Processing input with %d workers...\n", poolSize)

	// Start workers
	for i := 0; i < poolSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				bst := NewBST()
				for _, w := range allWords {
					dist := EditDistance(job.Word, w, 2)
					if dist > 2 {
						continue
					}
					_, freq := root.Search(w)
					bst.Insert(w, dist, freq)
				}
				results <- Result{
					Correction: Correction{
						Word:        job.Word,
						Suggestions: bst.GetSuggestions(),
					},
				}
			}
		}()
	}

	var report Report
	// Result collector
	var collectorWg sync.WaitGroup
	collectorWg.Add(1)
	go func() {
		defer collectorWg.Done()
		for res := range results {
			report.Corrections = append(report.Corrections, res.Correction)
		}
	}()

	// Producer
	scanner = bufio.NewScanner(in)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" {
			report.TotalWords++
			found, _ := root.Search(word)
			if !found {
				report.MisspelledWords++
				jobs <- Job{Word: word}
			}
		}
	}
	close(jobs)

	// Wait for workers to finish
	wg.Wait()
	close(results)
	// Wait for collector to finish
	collectorWg.Wait()

	fmt.Printf("Total time taken: %s\n", time.Since(start))

	out, err := os.Create("report.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer out.Close()

	err = json.NewEncoder(out).Encode(report)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Report written to report.json")

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not create memory profile: %v\n", err)
			return
		}
		defer f.Close()
		if err := pprof.WriteHeapProfile(f); err != nil {
			fmt.Fprintf(os.Stderr, "could not write memory profile: %v\n", err)
			return
		}
	}
}
