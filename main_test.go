package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestProcessSubcommands(t *testing.T) {
	// Create temporary files
	tempDir := os.TempDir()
	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")
	err := os.WriteFile(file1, []byte("File 1 content"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file1)
	err = os.WriteFile(file2, []byte("File 2 content"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file2)

	// Test with say subcommand
	entries, err := processSubcommands([]string{"say", "Hello world!"})
	if err != nil {
		t.Fatalf("processSubcommands failed: %v", err)
	}
	expected := []markdownEntry{{message: "Hello world!"}}
	if !reflect.DeepEqual(entries, expected) {
		t.Errorf("processSubcommands returned unexpected entries.\nExpected: %v\n  Actual: %v", expected, entries)
	}

	// Test with attach subcommand
	entries, err = processSubcommands([]string{"attach", file1, ",", "attach", file2})
	if err != nil {
		t.Fatalf("processSubcommands failed: %v", err)
	}
	expected = []markdownEntry{{filePath: file1}, {filePath: file2}}
	if !reflect.DeepEqual(entries, expected) {
		t.Errorf("processSubcommands returned unexpected entries.\nExpected: %v\n  Actual: %v", expected, entries)
	}

	// Test with insert subcommand
	insertFile := filepath.Join(tempDir, "insert.txt")
	err = os.WriteFile(insertFile, []byte("Insert content"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(insertFile)
	entries, err = processSubcommands([]string{"insert", insertFile})
	if err != nil {
		t.Fatalf("processSubcommands failed: %v", err)
	}
	expected = []markdownEntry{{message: "Insert content"}}
	if !reflect.DeepEqual(entries, expected) {
		t.Errorf("processSubcommands returned unexpected entries.\nExpected: %v\n  Actual: %v", expected, entries)
	}

	// Test with exec subcommand
	entries, err = processSubcommands([]string{"exec", "echo", "Exec", "output"})
	if err != nil {
		t.Fatalf("processSubcommands failed: %v", err)
	}
	expected = []markdownEntry{{output: "Exec output\n"}}
	if !reflect.DeepEqual(entries, expected) {
		t.Errorf("processSubcommands returned unexpected entries.\nExpected: %v\n  Actual: %v", expected, entries)
	}

	// Test with mixed subcommands
	entries, err = processSubcommands([]string{
		"say", "Message 1", ",", "attach", file1 + ",", "insert", insertFile, ",", "exec", "echo", "Exec", "output,", "say", "Message 2",
	})
	if err != nil {
		t.Fatalf("processSubcommands failed: %v", err)
	}
	expected = []markdownEntry{
		{message: "Message 1"},
		{filePath: file1},
		{message: "Insert content"},
		{output: "Exec output\n"},
		{message: "Message 2"},
	}
	if !reflect.DeepEqual(entries, expected) {
		t.Errorf("processSubcommands returned unexpected entries.\nExpected: %v\n  Actual: %v", expected, entries)
	}
}

func TestExecSub(t *testing.T) {
	// Test with a simple command
	entry, err := execSub([]string{"echo", "Hello,", "world!"})
	if err != nil {
		t.Fatalf("execSub failed: %v", err)
	}
	expected := markdownEntry{output: "Hello, world!\n"}
	if !reflect.DeepEqual(entry, expected) {
		t.Errorf("execSub returned unexpected entry.\nExpected: %v\n  Actual: %v", expected, entry)
	}

	// Test with a command that fails
	_, err = execSub([]string{"nonexistent-command"})
	if err == nil {
		t.Error("execSub should have returned an error for a nonexistent command")
	}
}

func TestSaySub(t *testing.T) {
	// Test with a simple message
	entry, err := saySub([]string{"Hello", "world!"})
	if err != nil {
		t.Fatalf("saySub failed: %v", err)
	}
	expected := markdownEntry{message: "Hello world!"}
	if !reflect.DeepEqual(entry, expected) {
		t.Errorf("saySub returned unexpected entry.\nExpected: %v\n  Actual: %v", expected, entry)
	}
}

func TestAttachSub(t *testing.T) {
	// Create a temporary file for testing
	tempFile := filepath.Join(os.TempDir(), "file.txt")
	err := os.WriteFile(tempFile, []byte("File content"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile)

	// Test with a valid file path
	entry, err := attachSub([]string{tempFile})
	if err != nil {
		t.Fatalf("attachSub failed: %v", err)
	}
	expected := markdownEntry{filePath: tempFile}
	if !reflect.DeepEqual(entry, expected) {
		t.Errorf("attachSub returned unexpected entry.\nExpected: %v\n  Actual: %v", expected, entry)
	}

	// Test with an invalid file path
	_, err = attachSub([]string{"nonexistent-file.txt"})
	if err == nil {
		t.Error("attachSub should have returned an error for a nonexistent file")
	}
}

func TestInsertSub(t *testing.T) {
	// Create a temporary file for testing
	tempFile := filepath.Join(os.TempDir(), "file.txt")
	err := os.WriteFile(tempFile, []byte("File content"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile)

	// Test with a valid file path
	entry, err := insertSub([]string{tempFile})
	if err != nil {
		t.Fatalf("insertSub failed: %v", err)
	}
	expected := markdownEntry{message: "File content"}
	if !reflect.DeepEqual(entry, expected) {
		t.Errorf("insertSub returned unexpected entry.\nExpected: %v\n  Actual: %v", expected, entry)
	}

	// Test with an invalid file path
	_, err = insertSub([]string{"nonexistent-file.txt"})
	if err == nil {
		t.Error("insertSub should have returned an error for a nonexistent file")
	}
}

func TestGenerateMarkdown(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a temporary file for testing
	tempFile := filepath.Join(tempDir, "file1.txt")
	err = os.WriteFile(tempFile, []byte("File 1 content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test with various entries
	entries := []markdownEntry{
		{message: "Message 1"},
		{filePath: tempFile},
		{message: "Message 2"},
		{output: "Command output"},
	}

	expected := "Message 1\n\n" +
		"`" + tempFile + "`\n" +
		"```\n" +
		"File 1 content```\n\n" +
		"Message 2\n\n" +
		"Command output\n"

	markdown := generateMarkdown(entries)
	if markdown != expected {
		t.Errorf("generateMarkdown returned unexpected markdown.\nExpected:\n%s\n  Actual:\n%s", expected, markdown)
	}
}
