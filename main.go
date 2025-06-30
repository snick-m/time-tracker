package main

import (
	"context"
	_ "embed"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time-tracker/config"
	"time-tracker/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/getlantern/systray"
	sheetsv4 "google.golang.org/api/sheets/v4"
)

//go:embed assets/icon.ico
var iconData []byte

var (
	cfg          *config.Config
	sheetService *sheetsv4.Service
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Google Sheets service
	if service, err := config.GetSheetService(ctx); err == nil {
		sheetService = service
		log.Println("Google Sheets service initialized")
	} else {
		log.Printf("Failed to create sheet service: %v", err)
	}

	// Create Fyne app
	uiApp := app.New()

	// Start system tray in a separate goroutine
	go startSystemTray(ctx, cancel, uiApp)

	// Start hotkey listener in a separate goroutine
	if cfg.Hotkey != "" {
		go ui.RegisterHotkey(ctx, cfg, sheetService, uiApp)
	} else {
		log.Println("No hotkey configured")
	}

	// Wait for exit signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	uiApp.Run()

	<-sigCh
	log.Println("Shutting down...")
}

func startSystemTray(ctx context.Context, cancel context.CancelFunc, uiApp fyne.App) {
	systray.Run(
		func() {
			// Set icon
			systray.SetIcon(iconData)
			systray.SetTitle("Time Tracker")
			systray.SetTooltip("Google Sheets Time Tracker")

			// Setup menu
			mAdd := systray.AddMenuItem("Add Time Entry", "Add new time entry")
			mConfig := systray.AddMenuItem("Configure", "Change settings")
			mQuit := systray.AddMenuItem("Exit", "Quit application")

			// If sheet service failed to initialize, disable add entry
			if sheetService == nil {
				mAdd.Disable()
			}

			fyne.Do(func() {
				ui.BuildPopup(ctx, cfg, sheetService, uiApp)
			})

			for {
				select {
				case <-mAdd.ClickedCh:
					if sheetService != nil {
						fyne.Do(func() {
							ui.ShowEntryPopup(ctx, cfg, sheetService, uiApp)
						})
					} else {
						log.Println("Sheet service not available")
					}
				case <-mConfig.ClickedCh:
					fyne.Do(configureApplication)
				case <-mQuit.ClickedCh:
					systray.Quit()
					cancel()
					return
				}
			}
		},
		func() {
			log.Println("Cleaning up system tray")
		},
	)
	os.Exit(0) // Ensure we exit cleanly when systray is closed
}

func configureApplication() {
	// Create app for configuration
	a := app.New()
	w := a.NewWindow("Configuration")

	// Dark theme
	a.Settings().SetTheme(&ui.DarkTheme{})

	spreadsheetIDEntry := widget.NewEntry()
	spreadsheetIDEntry.SetText(cfg.SpreadsheetID)

	sheetNameEntry := widget.NewEntry()
	sheetNameEntry.SetText(cfg.SheetName)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Google Sheet ID", Widget: spreadsheetIDEntry},
			{Text: "Sheet Name", Widget: sheetNameEntry},
		},
		OnSubmit: func() {
			// Update config
			cfg.SpreadsheetID = spreadsheetIDEntry.Text
			cfg.SheetName = sheetNameEntry.Text
			if err := config.SaveConfig(cfg); err != nil {
				dialog.ShowError(err, w)
			} else {
				dialog.ShowInformation("Success", "Configuration updated!", w)
				w.Close()
			}
		},
	}

	w.SetContent(form)
	w.Resize(fyne.NewSize(400, 100))
	w.ShowAndRun()
}
