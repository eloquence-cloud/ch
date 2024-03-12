package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.design/x/clipboard"
)

var (
	message     string
	messageFile string
)

func init() {
	flag.StringVar(&message, "m", "", "Insert a message paragraph into the markdown")
	flag.StringVar(&message, "message", "", "Insert a message paragraph into the markdown")
	flag.StringVar(&messageFile, "f", "", "Insert a message paragraph from a file into the markdown")
	flag.StringVar(&messageFile, "message-file", "", "Insert a message paragraph from a file into the markdown")
}

func main() {
	flag.Usage = func() {
		fmt.Println("Usage: ch [OPTIONS] [FILE_PATHS]...")
		fmt.Println()
		fmt.Println("Generate markdown for the specified files and directories.")
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	var markdown strings.Builder

	if message != "" {
		markdown.WriteString(message + "\n\n")
	} else if messageFile != "" {
		content, err := readMessageFile(messageFile)
		if err != nil {
			log.Fatal(err)
		}
		if content != "" {
			markdown.WriteString(content + "\n\n")
		}
	}

	for _, path := range flag.Args() {
		err := processPath(path, &markdown)
		if err != nil {
			log.Fatal(err)
		}
	}

	err := clipboard.Init()
	if err != nil {
		log.Fatal(err)
	}

	clipboard.Write(clipboard.FmtText, []byte(markdown.String()))
	fmt.Println("Markdown copied to the clipboard.")
}

func readMessageFile(path string) (string, error) {
	// Check if the provided path exists
	if _, err := os.Stat(path); err == nil {
		return readFile(path)
	}

	// Check if the provided path with .ch extension exists
	pathWithExt := path + ".ch"
	if _, err := os.Stat(pathWithExt); err == nil {
		return readFile(pathWithExt)
	}

	// Check if the provided name exists in ~/.ch directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	pathInHome := filepath.Join(homeDir, ".ch", filepath.Base(path)+".ch")
	if _, err := os.Stat(pathInHome); err == nil {
		return readFile(pathInHome)
	}

	return "", fmt.Errorf("message file not found: %s", path)
}

func readFile(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func processPath(path string, markdown *strings.Builder) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path not found: %s", path)
	}

	if fileInfo.IsDir() {
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return err
		}
		for _, file := range files {
			if !file.IsDir() && !strings.HasPrefix(file.Name(), ".") {
				err := generateMarkdown(filepath.Join(path, file.Name()), markdown)
				if err != nil {
					return err
				}
			}
		}
	} else {
		if !strings.HasPrefix(fileInfo.Name(), ".") {
			err := generateMarkdown(path, markdown)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func generateMarkdown(filePath string, markdown *strings.Builder) error {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	markdown.WriteString(fmt.Sprintf("\n`%s`\n\n", filePath))
	markdown.WriteString("```\n")
	markdown.WriteString(string(content))
	markdown.WriteString("\n```\n\n")

	return nil
}
