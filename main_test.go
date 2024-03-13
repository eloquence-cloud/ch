package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
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
