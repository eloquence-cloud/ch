package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestProcessArgs(t *testing.T) {
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
	messageFile := filepath.Join(tempDir, "message.txt")
	err = ioutil.WriteFile(messageFile, []byte("Message from file"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(messageFile)

	entries, err = processArgs([]string{"@" + messageFile})
	if err != nil {
		t.Fatalf("processArgs failed: %v", err)
	}
	if len(entries) != 1 || entries[0].message != "Message from file" {
		t.Errorf("processArgs returned unexpected entries: %v", entries)
	}

	// Test with inline message
	entries, err = processArgs([]string{"@", "Inline message"})
	if err != nil {
		t.Fatalf("processArgs failed: %v", err)
	}
	if len(entries) != 1 || entries[0].message != "Inline message" {
		t.Errorf("processArgs returned unexpected entries: %v", entries)
	}
}

// ... (other test functions remain unchanged)

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
	messageFile1 := filepath.Join(tempDir, "message1.txt")
	messageFile2 := filepath.Join(tempDir, "message2.txt")
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
	args := []string{"@" + messageFile1, file1, "@", "Inline message", file2, "@" + messageFile2}
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
