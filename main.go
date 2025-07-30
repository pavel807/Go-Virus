package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func buildBundle(app *tview.Application, logView *tview.TextView, mainAppPath, embeddedAppPath string) {
	logView.Clear()

	// Helper function to print to log and redraw
	logAndDraw := func(message string) {
		fmt.Fprintln(logView, message)
		app.Draw()
	}

	if mainAppPath == "" || embeddedAppPath == "" {
		logAndDraw("[red]Ошибка: Пожалуйста, укажите пути к обоим приложениям.")
		return
	}

	logAndDraw("Начало сборки...")
	logAndDraw(fmt.Sprintf("Главное приложение: %s", mainAppPath))
	logAndDraw(fmt.Sprintf("Встраиваемое приложение: %s", embeddedAppPath))
	logAndDraw("--------------------")

	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "app-bundler-")
	if err != nil {
		logAndDraw(fmt.Sprintf("Ошибка при создании временной директории: %s", err))
		return
	}
	defer os.RemoveAll(tempDir)
	logAndDraw(fmt.Sprintf("Временная директория создана: %s", tempDir))

	// Copy files to the temporary directory
	mainAppDest := filepath.Join(tempDir, "main_app")
	embeddedAppDest := filepath.Join(tempDir, "embedded_app")

	if err := copyFile(mainAppPath, mainAppDest); err != nil {
		logAndDraw(fmt.Sprintf("Ошибка при копировании главного приложения: %s", err))
		return
	}
	logAndDraw("Главное приложение скопировано.")

	if err := copyFile(embeddedAppPath, embeddedAppDest); err != nil {
		logAndDraw(fmt.Sprintf("Ошибка при копировании встраиваемого приложения: %s", err))
		return
	}
	logAndDraw("Встраиваемое приложение скопировано.")

	// Generate go.mod for the bundle
	goModPath := filepath.Join(tempDir, "go.mod")
	goModContent := "module app-bundle\n\ngo 1.16\n"
	if err := ioutil.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		logAndDraw(fmt.Sprintf("Ошибка при создании go.mod: %s", err))
		return
	}
	logAndDraw("go.mod создан.")

	// Generate main.go for the bundle
	mainGoPath := filepath.Join(tempDir, "main.go")
	mainGoContent := `
package main

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

//go:embed main_app
var mainApp []byte

//go:embed embedded_app
var embeddedApp []byte

func main() {
	fmt.Println("Запуск главного приложения...")
	runApp(mainApp, "main_app_temp")
	fmt.Println("Запуск встраиваемого приложения...")
	runApp(embeddedApp, "embedded_app_temp")
	fmt.Println("Все готово.")
}

func runApp(appData []byte, fileName string) {
	tmpfile, err := ioutil.TempFile("", fileName)
	if err != nil {
		fmt.Println("Ошибка при создании временного файла:", err)
		return
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(appData); err != nil {
		fmt.Println("Ошибка при записи во временный файл:", err)
		return
	}
	tmpfile.Close()

	if err := os.Chmod(tmpfile.Name(), 0755); err != nil {
		fmt.Println("Ошибка при изменении прав доступа к файлу:", err)
		return
	}

	cmd := exec.Command(tmpfile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("Ошибка при запуске приложения:", err)
	}
}
`
	if err := ioutil.WriteFile(mainGoPath, []byte(mainGoContent), 0644); err != nil {
		logAndDraw(fmt.Sprintf("Ошибка при создании main.go: %s", err))
		return
	}
	logAndDraw("main.go создан.")

	// Build the bundle
	logAndDraw("Запуск сборки бандла...")
	cmd := exec.Command("go", "build", "-o", "bundle")
	cmd.Dir = tempDir
	cmd.Stdout = logView
	cmd.Stderr = logView
	if err := cmd.Run(); err != nil {
		logAndDraw(fmt.Sprintf("[red]Ошибка при сборке бандла: %s", err))
		return
	}
	logAndDraw("Сборка бандла завершена.")

	// Copy the bundle to the current directory
	bundlePath := filepath.Join(tempDir, "bundle")
	destPath := "bundle"
	if err := copyFile(bundlePath, destPath); err != nil {
		logAndDraw(fmt.Sprintf("Ошибка при копировании бандла: %s", err))
		return
	}
	logAndDraw(fmt.Sprintf("Бандл скопирован в: %s", destPath))
	logAndDraw("--------------------")
	logAndDraw("[green]Сборка успешно завершена!")
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = out.ReadFrom(in)
	if err != nil {
		return err
	}
	return out.Close()
}

func main() {
	app := tview.NewApplication()

	// Log view
	logView := tview.NewTextView().
		SetScrollable(true).
		SetDynamicColors(true)
	logView.SetBorder(true).SetTitle("Логи")

	// Form for selecting files and building
	form := tview.NewForm().
		AddInputField("Главное приложение:", "", 40, nil, nil).
		AddInputField("Встраиваемое приложение:", "", 40, nil, nil)
	form.SetBorder(true).SetTitle("Выбор приложений")

	// Build button
	buildButton := tview.NewButton("--- BUILD ---").SetSelectedFunc(func() {
		mainApp := form.GetFormItem(0).(*tview.InputField).GetText()
		embeddedApp := form.GetFormItem(1).(*tview.InputField).GetText()
		go buildBundle(app, logView, mainApp, embeddedApp)
	})
	buildButton.SetBackgroundColor(tcell.ColorDarkGreen)


	// Layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(buildButton, 3, 1, false).
		AddItem(logView, 0, 2, false)

	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}
