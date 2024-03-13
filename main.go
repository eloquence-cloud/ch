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

	flag.Usage = printUsage
	flag.Parse()

	entries, err := processArgs(flag.Args())
	if err != nil {
		log.Fatal(err)
	}

	markdown := generateMarkdown(entries)

	if err := clipboard.Init(); err != nil {
		log.Printf("Failed to initialize clipboard: %v", err)
		if err := saveMarkdownToFile(markdown, os.TempDir()); err != nil {
			log.Fatalf("Failed to save markdown to file: %v", err)
		}
	} else {
		clipboard.Write(clipboard.FmtText, []byte(markdown))
		fmt.Println("Markdown copied to the clipboard.")
	}
}

func printUsage() {
	fmt.Println("Usage: ch [ file | directory |  -f \"message file\" | -m \"message\" ] ...")
	fmt.Println()
	fmt.Println("Generates an AI chat message containing the specified files, directories, and messages.")
	fmt.Println("Copies the markdown to the clipboard.")
	fmt.Println()
	fmt.Println("The output markdown displays files as code blocks, optionally interspersing")
	fmt.Println("these with messages in the order they appear on the command line.")
	fmt.Println("Recursively process directories. Constructs relative paths by")
	fmt.Println("stripping out any common prefix across the entire command line.")
	fmt.Println()
	fmt.Println("Files and directories can be remote paths, like user@host:/path/to/file.")
	fmt.Println("These are retrieved via the ssh protocol.")
	fmt.Println()
	fmt.Println("The material from each message file is inserted as one or more paragraphs")
	fmt.Println("ch searches for each message file in the following order:")
	fmt.Println()
	fmt.Println("  1. the provided path or name")
	fmt.Println("  2. if the path has no extension, the full provided path or name with '.ch' added")
	fmt.Println("  3. if those aren't found, and if the provided string is just a base filename")
	fmt.Println("     with no extension, then ~/.ch/<provided name>.ch")
}

func processArgs(args []string) ([]markdownEntry, error) {
	var entries []markdownEntry
	var messageFile, message string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-f" {
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing message file path")
			}
			messageFile = args[i+1]
			i++
			content, err := readMessageFile(messageFile)
			if err != nil {
				return nil, err
			}
			entries = append(entries, markdownEntry{message: content})
		} else if arg == "-m" {
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing message")
			}
			message = args[i+1]
			i++
			entries = append(entries, markdownEntry{message: message})
		} else {
			entries = append(entries, markdownEntry{filePath: arg})
		}
	}

	return entries, nil
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

func generateMarkdown(entries []markdownEntry) string {
	var markdown strings.Builder

	for _, entry := range entries {
		if entry.message != "" {
			markdown.WriteString(entry.message + "\n\n")
		} else if entry.filePath != "" {
			if err := processPath(entry.filePath, &markdown); err != nil {
				log.Printf("Failed to process path %s: %v", entry.filePath, err)
			}
		}
	}

	return markdown.String()
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
		return appendFileToMarkdown(path, markdown)
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
			if err := appendFileToMarkdown(filePath, markdown); err != nil {
				return err
			}
		}
	}

	return nil
}

func appendFileToMarkdown(filePath string, markdown *strings.Builder) error {
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

func saveMarkdownToFile(markdown, dir string) error {
	tempFile, err := ioutil.TempFile(dir, "ch_markdown_*.md")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer tempFile.Close()

	if _, err := tempFile.WriteString(markdown); err != nil {
		return fmt.Errorf("failed to write markdown to file: %v", err)
	}

	fmt.Printf("Markdown saved to file: %s\n", tempFile.Name())
	return nil
}
