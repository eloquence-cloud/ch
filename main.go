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

type markdownEntry struct {
	filePath string
	message  string
}

func main() {
	var messageFile string
	var message string

	flag.StringVar(&messageFile, "f", "", "Insert a message paragraph from a file into the markdown")
	flag.StringVar(&message, "m", "", "Insert a message paragraph into the markdown")

	flag.Usage = func() {
		fmt.Println("Usage: ch [OPTIONS] [FILE_PATHS]...")
		fmt.Println()
		fmt.Println("Generate markdown for the specified files, directories, and messages.")
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
	}

	flag.Parse()

	var entries []markdownEntry

	for _, arg := range flag.Args() {
		if messageFile != "" {
			content, err := readMessageFile(messageFile)
			if err != nil {
				log.Fatal(err)
			}
			entries = append(entries, markdownEntry{message: content})
			messageFile = ""
		} else if message != "" {
			entries = append(entries, markdownEntry{message: message})
			message = ""
		} else {
			entries = append(entries, markdownEntry{filePath: arg})
		}
	}

	var markdown strings.Builder

	for _, entry := range entries {
		if entry.message != "" {
			markdown.WriteString(entry.message + "\n\n")
		} else if entry.filePath != "" {
			err := processPath(entry.filePath, &markdown)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	if err := clipboard.Init(); err != nil {
		log.Printf("Failed to initialize clipboard: %v", err)
		saveMarkdownToFile(markdown.String())
	} else {
		clipboard.Write(clipboard.FmtText, []byte(markdown.String()))
		fmt.Println("Markdown copied to the clipboard.")
	}
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
		return processDirectory(path, markdown)
	}

	if !strings.HasPrefix(fileInfo.Name(), ".") {
		return generateMarkdown(path, markdown)
	}

	return nil
}

func processDirectory(dirPath string, markdown *strings.Builder) error {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		filePath := filepath.Join(dirPath, file.Name())
		if file.IsDir() {
			if err := processDirectory(filePath, markdown); err != nil {
				return err
			}
		} else if !strings.HasPrefix(file.Name(), ".") {
			if err := generateMarkdown(filePath, markdown); err != nil {
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

func saveMarkdownToFile(markdown string) {
	tempFile, err := ioutil.TempFile("", "ch_markdown_*.md")
	if err != nil {
		log.Printf("Failed to create temporary file: %v", err)
		return
	}
	defer tempFile.Close()

	if _, err := tempFile.WriteString(markdown); err != nil {
		log.Printf("Failed to write markdown to file: %v", err)
		return
	}

	fmt.Printf("Markdown saved to file: %s\n", tempFile.Name())
}
