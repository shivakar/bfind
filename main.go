package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"runtime"
	"sync"
)

var (
	name       = flag.String("name", "", "find files with names matching regex (case-senstive)")
	one        = flag.Bool("1", false, "end after finding the first matching entry")
	numThreads = flag.Int("num-threads", runtime.NumCPU(), "number of threads to process")
	fileType   = flag.String("type", "", "filter on file type attribute")
)

func init() {
	runtime.GOMAXPROCS(*numThreads)
}

func worker(dtp *StringQueue, fp chan string, wd chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		dir, ok := dtp.Pop()
		if !ok {
			break
		}
		contents, _ := dirContents(dir)
		for _, c := range contents {
			if ok, _ := isDir(c); ok {
				dtp.Push(c)
			}
			if ok, _ := filter(c); ok {
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

	wg := &sync.WaitGroup{}
	for i := 0; i < *numThreads; i++ {
		wg.Add(1)
		go worker(dirsToProcess, filteredPaths, workerDone, wg)
	}

	workerDoneCount := 0
	done := false
	for !done {
		select {
		case f := <-filteredPaths:
			fmt.Fprintf(os.Stdout, "%s\n", f)
			if *one {
				done = true
			}
		case <-workerDone:
			workerDoneCount++
			if workerDoneCount == *numThreads {
				done = true
			}
		}
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
