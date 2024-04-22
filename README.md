# ch

`ch` is a command-line tool for generating formatted markdown suitable for AI chat messages. It processes a mix of message files, inline messages, files, and directories, and combines them into a single markdown string.

## Features

- Include message files, inline messages, file contents, and directory contents in the generated markdown
- Display file contents as code blocks and messages as plaintext
- Specify the order of messages and file contents in the generated markdown
- Copy the generated markdown to the clipboard with the `-c` flag
- Recursively process directories to include all files

## Installation

To install `ch`, ensure you have Go installed and run the following command:

```
go install github.com/eloquence-cloud/ch@latest
```

This will download and install the latest version of `ch` in your `$GOPATH/bin` directory.

## Usage

```
ch - A tool for constructing chat messages for easy pasting into AI chat UIs.

ch allows you to combine messages, file contents, and command outputs into a
formatted markdown suitable for AI chat interactions. It provides a flexible
and extensible syntax for creating chat messages with ease.

Usage: ch [flags] subcommand [, subcommand ...]

Flags (one of -c or -o is required):
  -c           Copy the generated markdown to the clipboard
  -o file      Write the output to the specified file

Subcommands:
  say message       Emit a message (replace @<space>)
  attach path       Attach a file or directory of files (replace bare path)
  insert file       Insert the contents of a file (replace @file)
  exec command      Execute a command (pass command line to bash)
  paste             Insert the contents of the clipboard

Comma separation rules:
  - A comma at the end of a word ends that command and is not included in the word.
  - A comma alone in a word ends that command and is not included as a word.
  - A comma within a word is just part of that word.

Examples:
  ch -c say "Please review", attach file1.go, say "Thank you!"
  ch -o output.md say "Here are the changes:", insert changes.txt, attach src/
  ch -c exec "ls -l", say "Directory listing:", attach .
```

## Contributing

Contributions are welcome! If you find a bug or have a feature request, please open an issue on the GitHub repository. If you'd like to contribute code, please fork the repository and submit a pull request.

## License

`ch` is released under the GNU General Public License v3.0. See [LICENSE](LICENSE) for more information.
