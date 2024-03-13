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
Usage: ch [@message_file | @ "inline message" | file | directory] ...

Flags:
  -c    Copy the generated markdown to the clipboard
  -help
        Show usage information

Message files:
  @message_file.txt   Include the contents of 'message_file.txt' as a message.

Inline messages:
  @ "Inline message"   Include the specified text as an inline message.

Files and directories:
  file.go              Include the contents of 'file.go' as a code block.
  directory/           Recursively include all files in 'directory/' as code blocks.

ch searches for message files in the following order:
  1. The provided path or name
  2. If the path has no extension, the full provided path or name with '.ch' added
  3. If not found, and if the provided string is just a base filename with no
     extension, then ~/.ch/<provided_name>.ch

Examples:
  ch @message.txt file1.go @ "Please review" file2.go
  ch -c @ "Here are the changes:" @changes.txt src/
```

## Examples

Include a message file, a Go file, an inline message, and another Go file:

```
ch @message.txt file1.go @ "Please review" file2.go
```

Copy the generated markdown to the clipboard with an inline message, a message file, and a directory:

```
ch -c @ "Here are the changes:" @changes.txt src/
```

## Contributing

Contributions are welcome! If you find a bug or have a feature request, please open an issue on the GitHub repository. If you'd like to contribute code, please fork the repository and submit a pull request.

## License

`ch` is released under the GNU General Public License v3.0. See [LICENSE](LICENSE) for more information.