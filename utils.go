package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
)

func filter(entry string) (bool, error) {
	match, err := filterByType(entry)
	if err != nil || !match {
		return false, err
	}
	return filterByName(*name, entry)
}

func filterByType(entry string) (bool, error) {
	if *fileType == "" {
		return true, nil
	}
	info, err := os.Stat(entry)
	if err != nil {
		return false, err
	}
	switch *fileType {
	case "d":
		return info.IsDir(), nil
	case "f":
		return info.Mode().IsRegular(), nil
	}
	fmt.Fprintf(os.Stderr, "invalid filetype: %v\n", *fileType)
	return false, nil
}

func filterByName(nameRegex, entry string) (bool, error) {
	if nameRegex == "" {
		return true, nil
	}
	return regexp.MatchString(nameRegex, entry)
}

func isDir(p string) (bool, error) {
	info, err := os.Stat(p)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func dirContents(d string) ([]string, error) {
	out := []string{}
	contents, err := ioutil.ReadDir(d)
	if err != nil {
		return out, err
	}
	for _, c := range contents {
		out = append(out, path.Join(d, c.Name()))
	}
	return out, nil
}
