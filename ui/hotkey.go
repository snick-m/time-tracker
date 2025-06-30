package ui

import (
	"context"
	"log"
	"strings"
	"time-tracker/config"

	"fyne.io/fyne/v2"
	hook "github.com/robotn/gohook"
	"google.golang.org/api/sheets/v4"
)

func RegisterHotkey(ctx context.Context, cfg *config.Config, sheetService *sheets.Service, app fyne.App) {
	hook.Register(hook.KeyDown, strings.Split(cfg.Hotkey, "+"), func(e hook.Event) {
		log.Println("Hotkey pressed, showing entry popup")
		fyne.Do(func() {
			ShowEntryPopup(ctx, cfg, sheetService, app)
		})
	})

	s := hook.Start()
	defer hook.End()

	<-hook.Process(s)
}
