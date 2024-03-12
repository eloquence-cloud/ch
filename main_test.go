package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestProcessPath(t *testing.T) {
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
	err = ioutil.WriteFile(filepath.Join(testDir, "file2.txt"), []byte("File 2 content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	subDir := filepath.Join(testDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(subDir, "file3.txt"), []byte("File 3 content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var markdown strings.Builder
	err = processPath(testDir, &markdown)
	if err != nil {
		t.Fatal(err)
	}

	expected := "\n`" + filepath.Join(testDir, "file1.txt") + "`\n\n```\nFile 1 content\n```\n\n" +
		"\n`" + filepath.Join(testDir, "file2.txt") + "`\n\n```\nFile 2 content\n```\n\n" +
		"\n`" + filepath.Join(subDir, "file3.txt") + "`\n\n```\nFile 3 content\n```\n\n"
	if markdown.String() != expected {
		t.Errorf("Expected markdown: %s, got: %s", expected, markdown.String())
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
		t.Errorf("Expected markdown: %s, got: %s", expected, markdown.String())
	}
}
