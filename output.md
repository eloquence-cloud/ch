Have I made these changes correctly?



`main.go`

```
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.design/x/clipboard"
	"golang.org/x/term"
)

type markdownEntry struct {
	filePath string
	message  string
}

func main() {
	// print usage if we got no args or just "--help"
	if len(os.Args) == 1 || (len(os.Args) == 2 && os.Args[1] == "--help") {
		printUsage()
		return
	}
	entries, err := processArgs(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	markdown := generateMarkdown(entries)

	if !isTerminal(os.Stdout) {
		fmt.Println(markdown)
	} else {
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
}

func isTerminal(file *os.File) bool {
	return term.IsTerminal(int(file.Fd()))
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

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-f" {
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing message file path")
			}
			messageFile := args[i+1]
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
			message := args[i+1]
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

```


`main_test.go`

```
package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestProcessArgs(t *testing.T) {
	message := "Hello, world!"

	// Create temporary files
	tempDir := os.TempDir()
	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")
	err := ioutil.WriteFile(file1, []byte("File 1 content"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file1)
	err = ioutil.WriteFile(file2, []byte("File 2 content"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file2)

	// Test with file paths
	entries, err := processArgs([]string{file1, file2})
	if err != nil {
		t.Fatalf("processArgs failed: %v", err)
	}
	if len(entries) != 2 || entries[0].filePath != file1 || entries[1].filePath != file2 {
		t.Errorf("processArgs returned unexpected entries: %v", entries)
	}

	// Test with message file
	messageFile := filepath.Join(tempDir, "message.ch")
	err = ioutil.WriteFile(messageFile, []byte("Message from file"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(messageFile)

	entries, err = processArgs([]string{"-f", messageFile})
	if err != nil {
		t.Fatalf("processArgs failed: %v", err)
	}
	if len(entries) != 1 || entries[0].message != "Message from file" {
		t.Errorf("processArgs returned unexpected entries: %v", entries)
	}

	// Test with message
	entries, err = processArgs([]string{"-m", message})
	if err != nil {
		t.Fatalf("processArgs failed: %v", err)
	}
	if len(entries) != 1 || entries[0].message != message {
		t.Errorf("processArgs returned unexpected entries: %v", entries)
	}
}

func TestReadMessageFile(t *testing.T) {
	// Create a temporary message file
	tempDir := os.TempDir()
	messageFile := filepath.Join(tempDir, "message.ch")
	err := ioutil.WriteFile(messageFile, []byte("Test message"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(messageFile)

	// Test reading the message file
	content, err := readMessageFile(messageFile)
	if err != nil {
		t.Fatal(err)
	}
	if content != "Test message" {
		t.Errorf("Expected message: %s, got: %s", "Test message", content)
	}
}

func TestGenerateMarkdown(t *testing.T) {
	// Create temporary files
	tempDir := os.TempDir()
	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")
	err := ioutil.WriteFile(file1, []byte("File 1 content"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file1)
	err = ioutil.WriteFile(file2, []byte("File 2 content"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file2)

	entries := []markdownEntry{
		{message: "Test message 1"},
		{filePath: file1},
		{message: "Test message 2"},
		{filePath: file2},
	}

	markdown := generateMarkdown(entries)

	expected := "Test message 1\n\n" +
		"\n`" + file1 + "`\n\n```\nFile 1 content\n```\n\n" +
		"Test message 2\n\n" +
		"\n`" + file2 + "`\n\n```\nFile 2 content\n```\n\n"

	if markdown != expected {
		t.Errorf("Generated markdown does not match expected.\nExpected:\n%s\nGot:\n%s", expected, markdown)
	}
}

func TestProcessPath(t *testing.T) {
	// Create a temporary file
	tempDir := os.TempDir()
	file := filepath.Join(tempDir, "file.txt")
	err := ioutil.WriteFile(file, []byte("File content"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file)

	var markdown strings.Builder
	err = processPath(file, &markdown)
	if err != nil {
		t.Fatalf("processPath failed: %v", err)
	}

	expected := "\n`" + file + "`\n\n```\nFile content\n```\n\n"
	if markdown.String() != expected {
		t.Errorf("processPath output does not match expected.\nExpected:\n%s\nGot:\n%s", expected, markdown.String())
	}
}

func TestProcessDirectory(t *testing.T) {
	// Create a temporary directory with files and subdirectories
	tempDir := os.TempDir()
	testDir := filepath.Join(tempDir, "test")
	err := os.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testDir)

	err = ioutil.WriteFile(filepath.Join(testDir, "file1.txt"), []byte("File 1 content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	subDir := filepath.Join(testDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(subDir, "file2.txt"), []byte("File 2 content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var markdown strings.Builder
	err = processDirectory(testDir, &markdown)
	if err != nil {
		t.Fatal(err)
	}

	expected := "\n`" + filepath.Join(testDir, "file1.txt") + "`\n\n```\nFile 1 content\n```\n\n" +
		"\n`" + filepath.Join(subDir, "file2.txt") + "`\n\n```\nFile 2 content\n```\n\n"
	if markdown.String() != expected {
		t.Errorf("processDirectory output does not match expected.\nExpected:\n%s\nGot:\n%s", expected, markdown.String())
	}
}

func TestAppendFileToMarkdown(t *testing.T) {
	// Create a temporary file
	tempDir := os.TempDir()
	file := filepath.Join(tempDir, "file.txt")
	err := ioutil.WriteFile(file, []byte("File content"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file)

	var markdown strings.Builder
	err = appendFileToMarkdown(file, &markdown)
	if err != nil {
		t.Fatalf("appendFileToMarkdown failed: %v", err)
	}

	expected := "\n`" + file + "`\n\n```\nFile content\n```\n\n"
	if markdown.String() != expected {
		t.Errorf("appendFileToMarkdown output does not match expected.\nExpected:\n%s\nGot:\n%s", expected, markdown.String())
	}
}

func TestSaveMarkdownToFile(t *testing.T) {
	markdown := "# Test Markdown\n\nThis is a test."

	// Create a temporary directory for saving the markdown file
	tempDir := os.TempDir()
	err := saveMarkdownToFile(markdown, tempDir)
	if err != nil {
		t.Fatalf("saveMarkdownToFile failed: %v", err)
	}

	// Check if the file was created and contains the expected content
	files, err := filepath.Glob(filepath.Join(tempDir, "ch_markdown_*.md"))
	if err != nil {
		t.Fatalf("Failed to find saved markdown file: %v", err)
	}
	if len(files) == 0 {
		t.Errorf("No saved markdown file found")
		return
	}

	content, err := ioutil.ReadFile(files[0])
	if err != nil {
		t.Fatalf("Failed to read saved markdown file: %v", err)
	}
	if string(content) != markdown {
		t.Errorf("Saved markdown does not match expected.\nExpected:\n%s\nGot:\n%s", markdown, string(content))
	}

	// Clean up the temporary directory
	os.RemoveAll(tempDir)
}

func TestProcessArgsWithInterspersedMessages(t *testing.T) {
	// Create temporary files
	tempDir := os.TempDir()
	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")
	err := ioutil.WriteFile(file1, []byte("File 1 content"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file1)
	err = ioutil.WriteFile(file2, []byte("File 2 content"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file2)

	// Create temporary message files
	messageFile1 := filepath.Join(tempDir, "message1.ch")
	messageFile2 := filepath.Join(tempDir, "message2.ch")
	err = ioutil.WriteFile(messageFile1, []byte("Message 1"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(messageFile1)
	err = ioutil.WriteFile(messageFile2, []byte("Message 2"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(messageFile2)

	// Test with interspersed messages
	args := []string{"-f", messageFile1, file1, "-m", "Inline message", file2, "-f", messageFile2}
	entries, err := processArgs(args)
	if err != nil {
		t.Fatalf("processArgs failed: %v", err)
	}

	expected := []markdownEntry{
		{message: "Message 1"},
		{filePath: file1},
		{message: "Inline message"},
		{filePath: file2},
		{message: "Message 2"},
	}

	if !reflect.DeepEqual(entries, expected) {
		t.Errorf("processArgs returned unexpected entries.\nExpected: %v\nGot: %v", expected, entries)
	}
}

```


