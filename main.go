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
	copyToClipboard := flag.Bool("c", false, "Copy the generated markdown to the clipboard")
	helpFlag := flag.Bool("help", false, "Show usage information")
	flag.Parse()

	if *helpFlag {
		printUsage()
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		return
	}

	entries, err := processArgs(args)
	if err != nil {
		log.Fatal(err)
	}

	markdown := generateMarkdown(entries)

	if *copyToClipboard {
		if err := clipboard.Init(); err != nil {
			log.Printf("Failed to initialize clipboard: %v", err)
			fmt.Println(markdown)
		} else {
			clipboard.Write(clipboard.FmtText, []byte(markdown))
			fmt.Println("Markdown copied to the clipboard.")
		}
	} else {
		fmt.Println(markdown)
	}
}

func printUsage() {
	fmt.Println("Usage: ch [@message_file | @ \"inline message\" | file | directory] ...")
	fmt.Println()
	fmt.Println("ch is a tool for generating formatted markdown suitable for AI chat messages.")
	fmt.Println("It processes a mix of message files, inline messages, files, and directories,")
	fmt.Println("and combines them into a single markdown string.")
	fmt.Println()
	fmt.Println("Messages and file contents are included in the order they appear in the arguments.")
	fmt.Println("File contents are displayed as code blocks, while messages are treated as plaintext.")
	fmt.Println()
	fmt.Println("Flags:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Message files:")
	fmt.Println("  @message_file.txt   Include the contents of 'message_file.txt' as a message.")
	fmt.Println()
	fmt.Println("Inline messages:")
	fmt.Println("  @ \"Inline message\"   Include the specified text as an inline message.")
	fmt.Println()
	fmt.Println("Files and directories:")
	fmt.Println("  file.go              Include the contents of 'file.go' as a code block.")
	fmt.Println("  directory/           Recursively include all files in 'directory/' as code blocks.")
	fmt.Println()
	fmt.Println("ch searches for message files in the following order:")
	fmt.Println("  1. The provided path or name")
	fmt.Println("  2. If the path has no extension, the full provided path or name with '.ch' added")
	fmt.Println("  3. If not found, and if the provided string is just a base filename with no")
	fmt.Println("     extension, then ~/.ch/<provided_name>.ch")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ch @message.txt file1.go @ \"Please review\" file2.go")
	fmt.Println("  ch -c @ \"Here are the changes:\" @changes.txt src/")
}

func processArgs(args []string) ([]markdownEntry, error) {
	var entries []markdownEntry

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "@") {
			if arg == "@" {
				if i+1 >= len(args) {
					return nil, fmt.Errorf("missing inline message")
				}
				message := args[i+1]
				i++
				entries = append(entries, markdownEntry{message: message})
			} else {
				messageFile := arg[1:]
				content, err := readMessageFile(messageFile)
				if err != nil {
					return nil, err
				}
				entries = append(entries, markdownEntry{message: content})
			}
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
