# App Embedder

This is a simple TUI application written in Go that allows you to run pre-configured command-line applications and view their output within the TUI.

## Description

The application presents a list of available commands on the left side of the screen and an output view on the right. When you select a command from the list, it executes and its standard output and error are streamed to the output view.

This project serves as a basic example of how to build a TUI application with `tview` and how to manage and interact with subprocesses in Go.

## How to Build

To build the application, you need to have Go installed. Then, run the following command in the project root:

```bash
go build
```

This will create an executable file named `app-embedder` (or `app-embedder.exe` on Windows).

## How to Run

After building the application, you can run it with:

```bash
./app-embedder
```

### Cross-compilation

You can build this application for different platforms. Here are some examples:

**For Windows:**
```bash
GOOS=windows GOARCH=amd64 go build -o app-embedder.exe
```

**For Linux:**
```bash
GOOS=linux GOARCH=amd64 go build -o app-embedder
```

**For macOS:**
```bash
GOOS=darwin GOARCH=amd64 go build -o app-embedder
```

## How to Use

- Use the arrow keys to navigate the list of commands.
- Press `Enter` on a command to execute it.
- The output will be displayed in the right panel.
- To quit the application, select the "Quit" item or press `q`.
