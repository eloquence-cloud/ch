package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"golang.design/x/clipboard"
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

	// Test with say subcommand.
	entries, err := processSubcommands([]string{"say", "Hello world!"})
	if err != nil {
		t.Fatalf("processSubcommands failed: %v", err)
	}
	expected := []markdownEntry{{message: "Hello world!"}}
	if !reflect.DeepEqual(entries, expected) {
		t.Errorf("processSubcommands returned unexpected entries.\nExpected: %v\n  Actual: %v", expected, entries)
	}
	// Should get same result if the args are multiple words.
	entries, err = processSubcommands([]string{"say", "Hello", "world!"})
	if err != nil {
		t.Fatalf("processSubcommands failed: %v", err)
	}
	expected = []markdownEntry{{message: "Hello world!"}}
	if !reflect.DeepEqual(entries, expected) {
		t.Errorf("processSubcommands returned unexpected entries.\nExpected: %v\n  Actual: %v", expected, entries)
	}

	// Test with attach subcommand.
	entries, err = processSubcommands([]string{"attach", file1, tempDir + ",", "attach", file2})
	if err != nil {
		t.Fatalf("processSubcommands failed: %v", err)
	}
	expected = []markdownEntry{{filePath: file1}, {filePath: tempDir}, {filePath: file2}}
	if !reflect.DeepEqual(entries, expected) {
		t.Errorf("processSubcommands returned unexpected entries.\nExpected: %v\n  Actual: %v", expected, entries)
	}

	// Test with insert subcommand
	entries, err = processSubcommands([]string{"insert", file1, file2})
	if err != nil {
		t.Fatalf("processSubcommands failed: %v", err)
	}
	expected = []markdownEntry{{message: "File 1 content"}, {message: "File 2 content"}}
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
		"say", "Message 1", ",", "attach", file1 + ",", "insert", file2, ",", "exec", "echo", "Exec", "output,", "say", "Message 2",
	})
	if err != nil {
		t.Fatalf("processSubcommands failed: %v", err)
	}
	expected = []markdownEntry{
		{message: "Message 1"},
		{filePath: file1},
		{message: "File 2 content"},
		{output: "Exec output\n"},
		{message: "Message 2"},
	}
	if !reflect.DeepEqual(entries, expected) {
		t.Errorf("processSubcommands returned unexpected entries.\nExpected: %v\n  Actual: %v", expected, entries)
	}
}

func TestAttachSub(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")
	dir1 := filepath.Join(tempDir, "dir1")

	if err := os.WriteFile(file1, []byte("File 1 content"), 0644); err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("File 2 content"), 0644); err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	if err := os.Mkdir(dir1, 0755); err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}

	tests := []struct {
		name    string
		args    []string
		want    []markdownEntry
		wantErr bool
	}{
		{
			name:    "Single file",
			args:    []string{file1},
			want:    []markdownEntry{{filePath: file1}},
			wantErr: false,
		},
		{
			name:    "Multiple files",
			args:    []string{file1, file2},
			want:    []markdownEntry{{filePath: file1}, {filePath: file2}},
			wantErr: false,
		},
		{
			name:    "Directory",
			args:    []string{dir1},
			want:    []markdownEntry{{filePath: dir1}},
			wantErr: false,
		},
		{
			name:    "File and directory",
			args:    []string{file1, dir1},
			want:    []markdownEntry{{filePath: file1}, {filePath: dir1}},
			wantErr: false,
		},
		{
			name:    "Non-existent file",
			args:    []string{"nonexistent.txt"},
			want:    []markdownEntry{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := attachSub(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("attachSub() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("attachSub() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestPasteSub(t *testing.T) {
	// Set clipboard content for testing
	clipboard.Write(clipboard.FmtText, []byte("Clipboard content"))

	entries, err := pasteSub(nil)
	if err != nil {
		t.Fatalf("pasteSub failed: %v", err)
	}

	expected := []markdownEntry{{message: "Clipboard content"}}
	if !reflect.DeepEqual(entries, expected) {
		t.Errorf("pasteSub returned unexpected entries.\nExpected: %v\n  Actual: %v", expected, entries)
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
		{filePath: tempDir},
		{message: "Message 2"},
		{output: "Command output"},
	}

	expected := "Message 1\n\n" +

		"`" + tempFile + "`\n" +
		"```\n" +
		"File 1 content```\n\n" +

		// Should repeat tempFile as our rendering of tempDir, since tempFile is its sole content.
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

func TestMain(m *testing.M) {
	if err := clipboard.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize clipboard: %v\n", err)
		os.Exit(1)
	}
	exitCode := m.Run()

	// Clean up the clipboard after the tests are done
	clipboard.Write(clipboard.FmtText, nil)

	os.Exit(exitCode)
}
