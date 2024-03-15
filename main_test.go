// Copyright 2024 Dean Thompson dba Eloquence. All rights reserved.
//
// This file is part of the ch project.
//
// The ch project is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The ch project is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the ch project. If not, see <https://www.gnu.org/licenses/>.
//
// For more information, please contact Eloquence at info@eloquence.cloud.

package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestProcessArguments(t *testing.T) {
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

	// Test with file paths
	entries, err := processArguments([]string{file1, file2})
	if err != nil {
		t.Fatalf("processArguments failed: %v", err)
	}
	if len(entries) != 2 || entries[0].filePath != file1 || entries[1].filePath != file2 {
		t.Errorf("processArguments returned unexpected entries: %v", entries)
	}

	// Test with message file
	messageFile := filepath.Join(tempDir, "message.txt")
	err = os.WriteFile(messageFile, []byte("Message from file"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(messageFile)

	entries, err = processArguments([]string{"@" + messageFile})
	if err != nil {
		t.Fatalf("processArguments failed: %v", err)
	}
	if len(entries) != 1 || entries[0].message != "Message from file" {
		t.Errorf("processArguments returned unexpected entries: %v", entries)
	}

	// Test with inline message
	entries, err = processArguments([]string{"@", "Inline message"})
	if err != nil {
		t.Fatalf("processArguments failed: %v", err)
	}
	if len(entries) != 1 || entries[0].message != "Inline message" {
		t.Errorf("processArguments returned unexpected entries: %v", entries)
	}
}

func TestProcessArgumentsWithInterspersedMessages(t *testing.T) {
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

	// Create temporary message files
	messageFile1 := filepath.Join(tempDir, "message1.txt")
	messageFile2 := filepath.Join(tempDir, "message2.txt")
	err = os.WriteFile(messageFile1, []byte("Message 1"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(messageFile1)
	err = os.WriteFile(messageFile2, []byte("Message 2"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(messageFile2)

	// Test with interspersed messages
	args := []string{"@" + messageFile1, file1, "@", "Inline message", file2, "@" + messageFile2}
	entries, err := processArguments(args)
	if err != nil {
		t.Fatalf("processArguments failed: %v", err)
	}

	expected := []markdownEntry{
		{message: "Message 1"},
		{filePath: file1},
		{message: "Inline message"},
		{filePath: file2},
		{message: "Message 2"},
	}

	if !reflect.DeepEqual(entries, expected) {
		t.Errorf("processArguments returned unexpected entries.\nExpected: %v\nGot: %v", expected, entries)
	}
}
