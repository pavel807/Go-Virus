package main

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/rivo/tview"
)

// runCommand executes an external command and streams its output to the provided
// tview.TextView. It runs the command in a separate goroutine to avoid
// blocking the TUI.
func runCommand(app *tview.Application, outputView *tview.TextView, command string) {
	parts := strings.Split(command, " ")
	cmd := exec.Command(parts[0], parts[1:]...)

	// Clear the output view and show a message that the command is starting.
	outputView.Clear()
	fmt.Fprintln(outputView, "Running command:", command)
	fmt.Fprintln(outputView, "--------------------")

	// Get the command's output pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(outputView, "Error creating stdout pipe:", err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Fprintln(outputView, "Error creating stderr pipe:", err)
		return
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		fmt.Fprintln(outputView, "Error starting command:", err)
		return
	}

	// Goroutine to stream stdout and stderr to the output view
	go func() {
		// We need to use a writer that is safe for concurrent writes.
		// tview.TextView is not safe for concurrent writes, so we need to
		// queue updates to be done in the main goroutine.
		writer := tview.ANSIWriter(outputView)
		go io.Copy(writer, stdout)
		go io.Copy(writer, stderr)
	}()

	// Goroutine to wait for the command to finish
	go func() {
		err := cmd.Wait()
		app.QueueUpdateDraw(func() {
			if err != nil {
				fmt.Fprintln(outputView, "--------------------")
				fmt.Fprintln(outputView, "Command finished with error:", err)
			} else {
				fmt.Fprintln(outputView, "--------------------")
				fmt.Fprintln(outputView, "Command finished successfully.")
			}
		})
	}()
}

func main() {
	// Initialize the tview application.
	app := tview.NewApplication()

	// The text view that displays the output of the commands.
	outputView := tview.NewTextView().
		SetDynamicColors(true). // Enable color tags
		SetScrollable(true).
		SetChangedFunc(func() {
			// Redraw the application whenever the text changes.
			app.Draw()
		})
	outputView.SetBorder(true).SetTitle("Output")

	// The list of commands to run.
	list := tview.NewList().
		AddItem("ping google.com", "A simple ping command.", 'a', func() {
			runCommand(app, outputView, "ping google.com")
		}).
		AddItem("ls -l", "Lists files in the current directory.", 'b', func() {
			runCommand(app, outputView, "ls -l")
		}).
		AddItem("Quit", "Exit the application.", 'q', func() {
			app.Stop()
		})

	// The main layout, a flexbox that contains the list and the output view.
	flex := tview.NewFlex().
		// The list takes up 1/4 of the screen width.
		AddItem(list, 0, 1, true).
		// The output view takes up 3/4 of the screen width.
		AddItem(outputView, 0, 3, false)

	// Start the application.
	if err := app.SetRoot(flex, true).SetFocus(list).Run(); err != nil {
		panic(err)
	}
}
