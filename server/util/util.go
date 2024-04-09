package util

import (
	"fmt"
	"os"
	"strings"
)

func FindNodesFromFile(filename string, myAddress string) (nodes []string) {
	nodes = []string{}

	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
	}

	// src: https://golangdocs.com/golang-byte-array-to-string
	contents := string(fileBytes[:])

	servers := strings.Split(contents, "\n")
	for _, s := range servers {
		if s != myAddress {
			nodes = append(nodes, s)
		}
	}

	return
}

func AppendToFile(filename, text string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(text + "\n")
	if err != nil {
		return err
	}
	return nil
}
