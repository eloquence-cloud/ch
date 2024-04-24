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
	ctx, err := NewContext()
	if err != nil {
		t.Fatalf("Failed to create context: %v", err)
	}
	defer ctx.Cleanup()

	// Create temporary files within the context's temporary directory
	file1, file2 := createTempFiles(t, ctx)

	testCases := []struct {
		name     string
		args     []string
		expected []markdownEntry
	}{
		{
			name:     "Say subcommand",
			args:     []string{"say", "Hello world!"},
			expected: []markdownEntry{messageEntry{message: "Hello world!"}},
		},
		{
			name:     "Say subcommand with multiple words",
			args:     []string{"say", "Hello", "world!"},
			expected: []markdownEntry{messageEntry{message: "Hello world!"}},
		},
		{
			name: "Attach subcommand",
			args: []string{"attach", file1, ctx.TempDir + ",", "attach", file2},
			expected: []markdownEntry{
				// from explicit attach of file1
				fileEntry{storagePath: file1, originalPath: file1},
				// from attach of ctx.TempDir
				fileEntry{storagePath: file1, originalPath: file1},
				fileEntry{storagePath: file2, originalPath: file2},
				// from explicit attach of file2
				fileEntry{storagePath: file2, originalPath: file2},
			},
		},
		{
			name:     "Insert subcommand",
			args:     []string{"insert", file1, file2},
			expected: []markdownEntry{messageEntry{message: "File 1 content"}, messageEntry{message: "File 2 content"}},
		},
		{
			name:     "Exec subcommand",
			args:     []string{"exec", "echo", "Exec", "output"},
			expected: []markdownEntry{outputEntry{output: "Exec output\n"}},
		},
		{
			name: "Mixed subcommands",
			args: []string{
				"say", "Message 1", ",", "attach", file1 + ",", "insert", file2, ",", "exec", "echo", "Exec", "output,", "say", "Message 2",
			},
			expected: []markdownEntry{
				messageEntry{message: "Message 1"},
				fileEntry{storagePath: file1, originalPath: file1},
				messageEntry{message: "File 2 content"},
				outputEntry{output: "Exec output\n"},
				messageEntry{message: "Message 2"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			entries, err := processSubcommands(ctx, tc.args)
			if err != nil {
				t.Fatalf("processSubcommands failed: %v", err)
			}
			if !reflect.DeepEqual(entries, tc.expected) {
				t.Errorf("testing %v\nExpected entries: %v\n  Actual entries: %v", tc.name, tc.expected, entries)
			}
		})
	}
}

func createTempFiles(t *testing.T, ctx Context) (string, string) {
	file1 := filepath.Join(ctx.TempDir, "file1.txt")
	file2 := filepath.Join(ctx.TempDir, "file2.txt")
	err := os.WriteFile(file1, []byte("File 1 content"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(file2, []byte("File 2 content"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	return file1, file2
}

func TestAttachSub(t *testing.T) {
	ctx, err := NewContext()
	if err != nil {
		t.Fatalf("Failed to create context: %v", err)
	}
	defer ctx.Cleanup()

	// Create temporary files and directories within the context's temporary directory
	file1Path := filepath.Join(ctx.TempDir, "file1.txt")
	file2Path := filepath.Join(ctx.TempDir, "file2.txt")
	err = os.WriteFile(file1Path, []byte("File 1 content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	err = os.WriteFile(file2Path, []byte("File 2 content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	subDir := filepath.Join(ctx.TempDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	file3Path := filepath.Join(subDir, "file3.txt")
	err = os.WriteFile(file3Path, []byte("File 3 content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file3: %v", err)
	}

	testCases := []struct {
		name        string
		args        []string
		expected    []markdownEntry
		expectedErr error
	}{
		{
			name:        "Single file",
			args:        []string{file1Path},
			expected:    []markdownEntry{fileEntry{storagePath: file1Path, originalPath: file1Path}},
			expectedErr: nil,
		},
		{
			name:        "Multiple files",
			args:        []string{file1Path, file2Path},
			expected:    []markdownEntry{fileEntry{storagePath: file1Path, originalPath: file1Path}, fileEntry{storagePath: file2Path, originalPath: file2Path}},
			expectedErr: nil,
		},
		{
			name:        "Directory",
			args:        []string{ctx.TempDir},
			expected:    []markdownEntry{fileEntry{storagePath: file1Path, originalPath: file1Path}, fileEntry{storagePath: file2Path, originalPath: file2Path}, fileEntry{storagePath: file3Path, originalPath: file3Path}},
			expectedErr: nil,
		},
		{
			name:        "Non-existent file",
			args:        []string{"nonexistent.txt"},
			expected:    nil,
			expectedErr: fmt.Errorf("file does not exist: nonexistent.txt"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			entries, err := attachSub(ctx, tc.args)
			if tc.expectedErr != nil {
				if err == nil || err.Error() != tc.expectedErr.Error() {
					t.Errorf("Expected error: %v, got: %v", tc.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
			if !reflect.DeepEqual(entries, tc.expected) {
				t.Errorf("Expected entries: %v, got: %v", tc.expected, entries)
			}
		})
	}
}

func TestPasteSub(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		expected []markdownEntry
		wantErr  bool
	}{
		{
			name:     "Simple text",
			content:  "Clipboard content",
			expected: []markdownEntry{messageEntry{message: "Clipboard content"}},
			wantErr:  false,
		},
		{
			name:     "Empty clipboard",
			content:  "",
			expected: []markdownEntry{messageEntry{message: ""}},
			wantErr:  false,
		},
		{
			name:     "Multiline text",
			content:  "Line 1\nLine 2\nLine 3",
			expected: []markdownEntry{messageEntry{message: "Line 1\nLine 2\nLine 3"}},
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, err := NewContext()
			if err != nil {
				t.Fatalf("Failed to create context: %v", err)
			}
			defer ctx.Cleanup()

			if !tc.wantErr {
				clipboard.Write(clipboard.FmtText, []byte(tc.content))
			} else {
				// Simulate clipboard initialization failure
				clipboard.Write(clipboard.FmtText, nil)
			}

			entries, err := pasteSub(ctx, nil)
			if tc.wantErr {
				if err == nil {
					t.Error("Expected an error, but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("pasteSub failed: %v", err)
			}

			if !reflect.DeepEqual(entries, tc.expected) {
				t.Errorf("pasteSub returned unexpected entries.\nExpected: %v\n  Actual: %v", tc.expected, entries)
			}
		})
	}
}

func TestGenerateMarkdown(t *testing.T) {
	ctx, fileWithContentPath, emptyFilePath := setupTestFiles(t)
	defer ctx.Cleanup()

	specialCharFilePath := filepath.Join(ctx.TempDir, "file with spaces.txt")
	err := os.WriteFile(specialCharFilePath, []byte("File content\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file with special characters: %v", err)
	}

	testCases := []struct {
		name     string
		entries  []markdownEntry
		expected string
	}{
		{
			name: "Single message entry",
			entries: []markdownEntry{
				messageEntry{message: "Hello, world!"},
			},
			expected: "Hello, world!\n",
		},
		{
			name: "Single message entry with special characters",
			entries: []markdownEntry{
				messageEntry{message: "Hello, `world`!"},
			},
			expected: "Hello, `world`!\n",
		},
		{
			name: "Single file entry",
			entries: []markdownEntry{
				fileEntry{storagePath: fileWithContentPath, originalPath: fileWithContentPath},
			},
			expected: "`" + fileWithContentPath + "`\n```\nFile content\n```\n",
		},
		{
			name: "Single file entry with empty content",
			entries: []markdownEntry{
				fileEntry{storagePath: emptyFilePath, originalPath: emptyFilePath},
			},
			expected: "`" + emptyFilePath + "`\n```\n```\n",
		},
		{
			name: "Single file entry with special characters in path",
			entries: []markdownEntry{
				fileEntry{storagePath: specialCharFilePath, originalPath: specialCharFilePath},
			},
			expected: "`" + specialCharFilePath + "`\n```\nFile content\n```\n",
		},
		{
			name: "Single output entry",
			entries: []markdownEntry{
				outputEntry{output: "Command output"},
			},
			expected: "Command output\n",
		},
		{
			name: "Single output entry with empty output",
			entries: []markdownEntry{
				outputEntry{output: ""},
			},
			expected: "\n",
		},
		{
			name: "Mixed entries",
			entries: []markdownEntry{
				messageEntry{message: "Message 1"},
				fileEntry{storagePath: fileWithContentPath, originalPath: fileWithContentPath},
				outputEntry{output: "Command output"},
				messageEntry{message: "Message 2"},
			},
			expected: "Message 1\n\n`" + fileWithContentPath + "`\n```\nFile content\n```\n\nCommand output\n\nMessage 2\n",
		},
		{
			name: "Mixed entries with empty content",
			entries: []markdownEntry{
				messageEntry{message: ""},
				fileEntry{storagePath: emptyFilePath, originalPath: emptyFilePath},
				outputEntry{output: ""},
				messageEntry{message: ""},
			},
			expected: "`" + emptyFilePath + "`\n```\n```\n",
		},
		{
			name:     "Empty entries",
			entries:  []markdownEntry{},
			expected: "\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			markdown := generateMarkdown(tc.entries)
			if markdown != tc.expected {
				t.Errorf("Unexpected markdown generated for %q.\nExpected:\n%q\nActual:\n%q", tc.name, tc.expected, markdown)
			}
		})
	}
}

func setupTestFiles(t *testing.T) (Context, string, string) {
	ctx, err := NewContext()
	if err != nil {
		t.Fatalf("Failed to create context: %v", err)
	}

	fileWithContentPath := filepath.Join(ctx.TempDir, "file.txt")
	err = os.WriteFile(fileWithContentPath, []byte("File content\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file with content: %v", err)
	}

	emptyFilePath := filepath.Join(ctx.TempDir, "empty.txt")
	err = os.WriteFile(emptyFilePath, []byte{}, 0644)
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	return ctx, fileWithContentPath, emptyFilePath
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
