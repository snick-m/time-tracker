package ui

import (
	"context"
	"fmt"
	"image/color"
	"strconv"
	"time"

	"time-tracker/config"
	"time-tracker/sheets"

	"fyne.io/fyne/v2"
	hook "github.com/robotn/gohook"

	// app import removed since we're using passed app instance
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	sheetsv4 "google.golang.org/api/sheets/v4"
)

// Custom dark theme
type DarkTheme struct{}

func (t *DarkTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameBackground {
		return color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	}
	if name == theme.ColorNameForeground {
		return color.NRGBA{R: 240, G: 240, B: 240, A: 255}
	}
	if name == theme.ColorNameButton {
		return color.NRGBA{R: 60, G: 60, B: 60, A: 255}
	}
	if name == theme.ColorNameInputBackground {
		return color.NRGBA{R: 40, G: 40, B: 40, A: 255}
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (t *DarkTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *DarkTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *DarkTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

var (
	x             float32 = 0.0
	w             float32 = 200.0
	window        fyne.Window
	windowVisible bool = false
	escapeVisible bool = false

	dateEntry        *widget.Entry
	hoursEntry       *widget.Entry
	descEntry        *widget.Entry
	projectEntry     *widget.Entry
	branchEntry      *widget.Entry
	commitStartEntry *widget.Entry
	commitEndEntry   *widget.Entry
)

func resetPopup() {
	fyne.Do(func() {
		currentDate := time.Now().Format("01/02/2006")
		dateEntry.SetText(currentDate)
		dateEntry.SetValidationError(nil)
		hoursEntry.SetText("")
		hoursEntry.SetValidationError(nil)
		descEntry.SetText("")
		descEntry.SetValidationError(nil)
		projectEntry.SetText("")
		projectEntry.SetValidationError(nil)
		branchEntry.SetText("")
		commitStartEntry.SetText("")
		commitEndEntry.SetText("")

		window.Hide()
		windowVisible = false
	})
}

func BuildPopup(
	ctx context.Context,
	cfg *config.Config,
	sheetService *sheetsv4.Service,
	app fyne.App,
) {
	if window != nil {
		window.Close()
	}
	// Use existing app instance
	drv := app.Driver()
	if drv, ok := drv.(desktop.Driver); ok {
		window = drv.CreateSplashWindow()
	} else {
		window = app.NewWindow("Time Entry")
	}

	// Apply dark theme
	app.Settings().SetTheme(&DarkTheme{})

	// Current date in MM/DD/YYYY format
	currentDate := time.Now().Format("01/02/2006")

	hContainer := container.NewWithoutLayout()

	// Create form fields
	dateEntry = widget.NewEntry()
	dateEntry.SetText(currentDate)
	dateEntry.SetPlaceHolder("Date (MM/DD/YYYY)")
	dateEntry.Resize(fyne.NewSize(w, 40))
	dateEntry.Move(fyne.NewPos(x, 0))
	x += w + 10

	hoursEntry = widget.NewEntry()
	hoursEntry.SetPlaceHolder("Hours")
	hoursEntry.Resize(fyne.NewSize(w, 40))
	hoursEntry.Move(fyne.NewPos(x, 0))
	x += w + 10

	descEntry = widget.NewEntry()
	descEntry.SetPlaceHolder("Description")
	descEntry.Resize(fyne.NewSize(w, 40))
	descEntry.Move(fyne.NewPos(x, 0))
	x += w + 10

	projectEntry = widget.NewEntry()
	projectEntry.SetPlaceHolder("Repo/Project")
	projectEntry.Resize(fyne.NewSize(w, 40))
	projectEntry.Move(fyne.NewPos(x, 0))
	x += w + 10

	branchEntry = widget.NewEntry()
	branchEntry.SetPlaceHolder("Branch")
	branchEntry.Resize(fyne.NewSize(w, 40))
	branchEntry.Move(fyne.NewPos(x, 0))
	x += w + 10

	commitStartEntry = widget.NewEntry()
	commitStartEntry.SetPlaceHolder("Commit Hash (Start)")
	commitStartEntry.Resize(fyne.NewSize(w*2, 40))
	commitStartEntry.Move(fyne.NewPos(x, 0))
	x += w*2 + 10

	commitEndEntry = widget.NewEntry()
	commitEndEntry.SetPlaceHolder("Commit Hash (End)")
	commitEndEntry.Resize(fyne.NewSize(w*2, 40))
	commitEndEntry.Move(fyne.NewPos(x, 0))
	x += w*2 + 10

	// Add entries with padding
	hContainer.Add(dateEntry)
	hContainer.Add(hoursEntry)
	hContainer.Add(descEntry)
	hContainer.Add(projectEntry)
	hContainer.Add(branchEntry)
	hContainer.Add(commitStartEntry)
	hContainer.Add(commitEndEntry)

	// Validators
	dateEntry.Validator = func(text string) error {
		if text == "" {
			return fmt.Errorf("date is required")
		}
		if _, err := time.Parse("01/02/2006", text); err != nil {
			return fmt.Errorf("invalid date format, use MM/DD/YYYY")
		}
		return nil
	}
	hoursEntry.Validator = func(text string) error {
		if text == "" {
			return fmt.Errorf("hours are required")
		}
		if _, err := strconv.ParseFloat(text, 64); err != nil {
			return fmt.Errorf("invalid hours value, must be a number")
		}
		return nil
	}
	descEntry.Validator = func(text string) error {
		if text == "" {
			return fmt.Errorf("description is required")
		}
		return nil
	}
	projectEntry.Validator = func(text string) error {
		if text == "" {
			return fmt.Errorf("project/repo is required")
		}
		return nil
	}

	submitBtn := widget.NewButton("Submit", func() {
		// Validate input
		hours := hoursEntry.Text
		if _, err := strconv.ParseFloat(hours, 64); err != nil {
			dialog.ShowError(fmt.Errorf("invalid hours value"), window)
			return
		}

		if hours == "" || descEntry.Text == "" || projectEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("hours, description and project are required"), window)
			return
		}

		// Prepare row data
		row := []any{
			dateEntry.Text,
			hours,
			descEntry.Text,
			projectEntry.Text,
			branchEntry.Text,
			commitStartEntry.Text,
			commitEndEntry.Text,
		}

		// Show progress dialog
		progress := dialog.NewCustomWithoutButtons("Submitting",
			widget.NewProgressBarInfinite(), window)
		progress.Show()

		fyne.Do(func() {
			// Submit to Google Sheets
			err := sheets.AppendRow(ctx, sheetService, cfg.SpreadsheetID, cfg.SheetName, row)

			// Close progress
			progress.Hide()

			exitWin := app.NewWindow("Exit Confirmation")
			if drv, ok := app.Driver().(desktop.Driver); ok {
				exitWin = drv.CreateSplashWindow()
			}
			exitWin.Resize(fyne.NewSize(200, 150))
			exitWin.Show()
			exitWin.RequestFocus()
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("submission failed: %v", err), exitWin)
				})
			} else {
				fyne.Do(func() {
					dialog.ShowCustomWithoutButtons("Success", widget.NewLabel("Time entry added!"), exitWin)
				})
				// Clear entries
				resetPopup()
			}
			// Wait a moment before closing in a goroutine
			go func() {
				time.Sleep(1 * time.Second)
				fyne.Do(exitWin.Close)
			}()
		})
	})
	submitBtn.Move(fyne.NewPos(x, 0))
	submitBtn.Resize(fyne.NewSize(w*2/3, 40))
	x += w * 2 / 3

	hContainer.Add(submitBtn)

	paddedContainer := container.New(layout.NewCustomPaddedLayout(20, 20, 20, 20), hContainer)
	window.SetContent(paddedContainer)
	window.Resize(fyne.NewSize(x+40, 40+40)) // for padding

	hook.Register(hook.KeyDown, []string{"esc"}, func(e hook.Event) {
		if windowVisible && !escapeVisible {
			escapeVisible = true
			fyne.Do(func() {
				exitWin := app.NewWindow("Exit Confirmation")
				if drv, ok := app.Driver().(desktop.Driver); ok {
					exitWin = drv.CreateSplashWindow()
				}
				exitWin.Resize(fyne.NewSize(250, 200))
				exitWin.Show()
				exitWin.RequestFocus()
				fyne.Do(func() {
					dialog.ShowConfirm("Exit",
						"Are you sure you want to exit?",
						func(confirmed bool) {
							if confirmed {
								resetPopup()
								window.Hide()
								windowVisible = false
								fyne.Do(exitWin.Close)
							} else {
								fyne.Do(exitWin.Close)
							}
							escapeVisible = false
						},
						exitWin)
				})
			})
		}
	})
}

func ShowEntryPopup(
	ctx context.Context,
	cfg *config.Config,
	sheetService *sheetsv4.Service,
	app fyne.App,
) {
	if !escapeVisible {
		window.Show()
		window.RequestFocus()
		windowVisible = true
	}
}
