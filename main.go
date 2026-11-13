package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	name       = flag.String("name", "", "find files with names matching regex (case-senstive)")
	one        = flag.Bool("1", false, "print only the shallowest matching entry (ties broken by name)")
	numThreads = flag.Int("num-threads", runtime.NumCPU(), "number of threads to process")
	fileType   = flag.String("type", "", "filter on file type attribute")
)

func init() {
	runtime.GOMAXPROCS(*numThreads)
}

func depth(p string) int64 {
	return int64(strings.Count(p, "/"))
}

// lowerBound sets b to d if d is smaller than the current value of b
func lowerBound(b *atomic.Int64, d int64) {
	for {
		cur := b.Load()
		if d >= cur || b.CompareAndSwap(cur, d) {
			return
		}
	}
}

// shallower reports whether path a comes before path b in -1 mode
func shallower(a, b string) bool {
	if depth(a) != depth(b) {
		return depth(a) < depth(b)
	}
	return a < b
}

// worker processes directories from dtp. In -1 mode, bestDepth holds the depth
// of the shallowest match so far. The worker skips directories whose entries
// are deeper than that, so the search does not wait at each level but still
// finds every match at the shallowest depth.
func worker(dtp *StringQueue, fp chan string, wd chan bool, wg *sync.WaitGroup, bestDepth *atomic.Int64) {
	defer wg.Done()

	for {
		dir, ok := dtp.Pop()
		if !ok {
			break
		}
		if *one && depth(dir) >= bestDepth.Load() {
			dtp.Done()
			continue
		}
		contents, _ := dirContents(dir)
		for _, c := range contents {
			if ok, _ := isDir(c); ok {
				dtp.Push(c)
			}
			if ok, _ := filter(c); ok {
				if *one {
					lowerBound(bestDepth, depth(c))
				}
				fp <- c
			}
		}
		dtp.Done()
	}
	wd <- true
}

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Invalid usage. Insufficient arguments.\n")
		pname := path.Base(os.Args[0])
		fmt.Fprintf(os.Stderr, "Usage: %s [options] DIRECTORIES...\n", pname)
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *name != "" {
		re, err := regexp.Compile(*name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid name regex - %s. Error string: %v\n", *name, err)
			os.Exit(1)
		}
		nameRegex = re
	}

	dirsToProcess := NewStringQueue()
	filteredPaths := make(chan string)
	workerDone := make(chan bool, *numThreads)

	dirs := flag.Args()
	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			fmt.Fprintf(os.Stderr, "Invalid search directory - %s. Error string: %v", dir, err)
			os.Exit(1)
		}
		dirsToProcess.Push(dir)
	}

	bestDepth := &atomic.Int64{}
	bestDepth.Store(math.MaxInt64)

	wg := &sync.WaitGroup{}
	for i := 0; i < *numThreads; i++ {
		wg.Add(1)
		go worker(dirsToProcess, filteredPaths, workerDone, wg, bestDepth)
	}

	best := ""
	workerDoneCount := 0
	done := false
	for !done {
		select {
		case f := <-filteredPaths:
			if !*one {
				fmt.Fprintf(os.Stdout, "%s\n", f)
			} else if best == "" || shallower(f, best) {
				best = f
			}
		case <-workerDone:
			workerDoneCount++
			if workerDoneCount == *numThreads {
				done = true
			}
		}
	}
	if best != "" {
		fmt.Fprintf(os.Stdout, "%s\n", best)
	}

	// Stopping all threads if not already stopped
	dirsToProcess.Close()

	// Ignoring any values written to filteredPaths
	go func() {
		ok := true
		for ok {
			_, ok = <-filteredPaths
		}
	}()

	wg.Wait()

	close(filteredPaths)
	close(workerDone)
}
