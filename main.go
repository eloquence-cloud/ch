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
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.design/x/clipboard"
)

type markdownEntry struct {
	filePath string
	message  string
	output   string
}

func (e markdownEntry) String() string {
	return fmt.Sprintf("{filePath: %q, message: %q, output: %q}", e.filePath, e.message, e.output)
}

type subcommand struct {
	name string
	fn   func(args []string) ([]markdownEntry, error)
}

var subcommands = []subcommand{
	{"say", saySub},
	{"attach", attachSub},
	{"insert", insertSub},
	{"exec", execSub},
}

//////////// processing of subcommands ///////////////

func processSubcommands(args []string) ([]markdownEntry, error) {
	var entries []markdownEntry
	var accumCommand []string
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		if strings.HasSuffix(arg, ",") {
			argWithoutComma := strings.TrimSuffix(arg, ",")
			if len(argWithoutComma) > 0 {
				accumCommand = append(accumCommand, argWithoutComma)
			}
			subcommandEntries, err := executeSubcommand(accumCommand)
			if err != nil {
				return nil, err
			}
			entries = append(entries, subcommandEntries...)
			accumCommand = nil
		} else {
			accumCommand = append(accumCommand, arg)
		}
	}
	if len(accumCommand) > 0 {
		subcommandEntries, err := executeSubcommand(accumCommand)
		if err != nil {
			return nil, err
		}
		entries = append(entries, subcommandEntries...)
	}
	return entries, nil
}

func executeSubcommand(args []string) ([]markdownEntry, error) {
	if len(args) == 0 {
		return []markdownEntry{}, fmt.Errorf("no subcommand provided")
	}
	command := args[0]
	var matches []subcommand
	for _, sub := range subcommands {
		if strings.HasPrefix(sub.name, command) {
			matches = append(matches, sub)
		}
	}
	if len(matches) == 0 {
		return []markdownEntry{}, fmt.Errorf("unknown subcommand: %s", command)
	}
	if len(matches) > 1 {
		return []markdownEntry{}, fmt.Errorf("ambiguous subcommand: %s", command)
	}
	return matches[0].fn(args[1:])
}

func saySub(args []string) ([]markdownEntry, error) {
	message := strings.Join(args, " ")
	return []markdownEntry{{message: message}}, nil
}

func attachSub(args []string) ([]markdownEntry, error) {
	var entries []markdownEntry
	for _, filePath := range args {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return []markdownEntry{}, fmt.Errorf("file does not exist: %s", filePath)
		}
		entries = append(entries, markdownEntry{filePath: filePath})
	}
	return entries, nil
}

func insertSub(args []string) ([]markdownEntry, error) {
	var entries []markdownEntry
	for _, file := range args {
		content, err := readFile(file)
		if err != nil {
			return nil, err
		}
		entries = append(entries, markdownEntry{message: content})
	}
	return entries, nil
}

func execSub(args []string) ([]markdownEntry, error) {
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.Output()
	if err != nil {
		return []markdownEntry{}, fmt.Errorf("command execution failed: %v", err)
	}
	return []markdownEntry{{output: string(output)}}, nil
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
	files, err := os.ReadDir(dirPath)
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
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	markdown.WriteString(fmt.Sprintf("`%s`\n", filePath))
	markdown.WriteString("```\n")
	markdown.WriteString(string(content))
	markdown.WriteString("```\n\n")

	return nil
}

func readFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

//////////// generation of markdown ///////////////

func generateMarkdown(entries []markdownEntry) string {
	var markdown strings.Builder

	for _, entry := range entries {
		if entry.message != "" {
			markdown.WriteString(entry.message + "\n\n")
		} else if entry.filePath != "" {
			if err := processPath(entry.filePath, &markdown); err != nil {
				log.Printf("Failed to process path %s: %v", entry.filePath, err)
			}
		} else if entry.output != "" {
			markdown.WriteString(entry.output + "\n\n")
		}
	}

	return strings.TrimSpace(markdown.String()) + "\n"
}

//////////// main ///////////////

func printUsage() {
	fmt.Println("ch - A tool for constructing chat messages for easy pasting into AI chat UIs.")
	fmt.Println()
	fmt.Println("ch allows you to combine messages, file contents, and command outputs into a")
	fmt.Println("formatted markdown suitable for AI chat interactions. It provides a flexible")
	fmt.Println("and extensible syntax for creating chat messages with ease.")
	fmt.Println()
	fmt.Println("Usage: ch [flags] subcommand [, subcommand ...]")
	fmt.Println()
	fmt.Println("Flags (one of -c or -o is required):")
	fmt.Println("  -c           Copy the generated markdown to the clipboard")
	fmt.Println("  -o file      Write the output to the specified file")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  say message       Emit a message (replace @<space>)")
	fmt.Println("  attach path       Attach a file or directory of files (replace bare path)")
	fmt.Println("  insert file       Insert the contents of a file (replace @file)")
	fmt.Println("  exec command      Execute a command (pass command line to bash)")
	fmt.Println()
	fmt.Println("Comma separation rules:")
	fmt.Println("  - A comma at the end of a word ends that command and is not included in the word.")
	fmt.Println("  - A comma alone in a word ends that command and is not included as a word.")
	fmt.Println("  - A comma within a word is just part of that word.")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ch -c say \"Please review\", attach file1.go, say \"Thank you!\"")
	fmt.Println("  ch -o output.md say \"Here are the changes:\", insert changes.txt, attach src/")
	fmt.Println("  ch -c exec \"ls -l\", say \"Directory listing:\", attach .")
}

// Upcoming
// func printUsage() {
// 	fmt.Println("ch - A tool for constructing chat messages for easy pasting into AI chat UIs.")
// 	fmt.Println()
// 	fmt.Println("ch allows you to combine messages, file contents, and command outputs into a")
// 	fmt.Println("formatted markdown suitable for AI chat interactions. It provides a flexible")
// 	fmt.Println("and extensible syntax for creating chat messages with ease.")
// 	fmt.Println()
// 	fmt.Println("Usage: ch [flags] subcommand [, subcommand ...]")
// 	fmt.Println()
// 	fmt.Println("Flags (one of -c or -o is required):")
// 	fmt.Println("  -c           Copy the generated markdown to the clipboard")
// 	fmt.Println("  -o file      Write the output to the specified file")
// 	fmt.Println()
// 	fmt.Println("Subcommands:")
// 	fmt.Println("  say message       Emit a message (replace @<space>)")
// 	fmt.Println("  attach path       Attach a file or directory of files (replace bare path)")
// 	fmt.Println("  insert file       Insert the contents of a file (replace @file)")
// 	fmt.Println("  exec command      Execute a command (pass command line to bash)")
// 	fmt.Println()
// 	fmt.Println("Custom Subcommands:")
// 	fmt.Println("  To create a new subcommand, place either commandname.ch or commandname.sh")
// 	fmt.Println("  in ~/.ch/commands directory.")
// 	fmt.Println()
// 	fmt.Println("  A .ch file contains a sequence of ch commands separated by either newline or comma.")
// 	fmt.Println("  Arguments can be substituted by $1, $2, etc.")
// 	fmt.Println("  Flags can be substituted by $flagname.")
// 	fmt.Println()
// 	fmt.Println("  Comma separation rules:")
// 	fmt.Println("  - A comma at the end of a word ends that command and is not included in the word.")
// 	fmt.Println("  - A comma alone in a word ends that command and is not included as a word.")
// 	fmt.Println("  - A comma within a word is just part of that word.")
// 	fmt.Println()
// 	fmt.Println("Examples:")
// 	fmt.Println("  ch -c say \"Please review\" attach file1.go, say \"Thank you!\"")
// 	fmt.Println("  ch -o output.md say \"Here are the changes:\" insert changes.txt attach src/")
// 	fmt.Println("  ch -c exec \"ls -l\" say \"Directory listing:\" attach .")
// }

func main() {
	copyToClipboard := flag.Bool("c", false, "Copy the generated markdown to the clipboard")
	outputFile := flag.String("o", "", "Write the output to the specified file")
	helpFlag := flag.Bool("help", false, "Show usage information")
	flag.Parse()

	if *helpFlag {
		printUsage()
		return
	}

	subcommands := flag.Args()
	entries, err := processSubcommands(subcommands)
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
	} else if *outputFile != "" {
		if err := os.WriteFile(*outputFile, []byte(markdown), 0644); err != nil {
			log.Printf("Failed to write output to file: %v", err)
		} else {
			fmt.Printf("Markdown written to file: %s\n", *outputFile)
		}
	} else {
		fmt.Println(markdown)
	}
}
