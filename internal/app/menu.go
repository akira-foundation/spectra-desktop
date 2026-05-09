package app

import (
	"runtime"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) BuildApplicationMenu() *menu.Menu {
	appMenu := menu.NewMenu()

	if runtime.GOOS == "darwin" {
		appMenu.Append(menu.AppMenu())
	}

	fileMenu := appMenu.AddSubmenu("File")
	fileMenu.AddText("New project", keys.CmdOrCtrl("n"), a.emitMenu("menu:new-project"))
	fileMenu.AddSeparator()
	fileMenu.AddText("Sync endpoints", keys.CmdOrCtrl("r"), a.emitMenu("menu:sync"))
	fileMenu.AddText("Export OpenAPI", keys.CmdOrCtrl("e"), a.emitMenu("menu:export-openapi"))

	if runtime.GOOS == "darwin" {
		appMenu.Append(menu.EditMenu())
	}

	viewMenu := appMenu.AddSubmenu("View")
	viewMenu.AddText("Dashboard", keys.CmdOrCtrl("1"), a.emitMenuPage("dashboard"))
	viewMenu.AddText("API Inspector", keys.CmdOrCtrl("2"), a.emitMenuPage("inspector"))
	viewMenu.AddText("Collections", keys.CmdOrCtrl("3"), a.emitMenuPage("collections"))
	viewMenu.AddText("Scratch", keys.CmdOrCtrl("4"), a.emitMenuPage("scratch"))
	viewMenu.AddText("Accounts", keys.CmdOrCtrl("5"), a.emitMenuPage("accounts"))
	viewMenu.AddText("Mock server", keys.CmdOrCtrl("6"), a.emitMenuPage("mock"))
	viewMenu.AddText("Snapshots", keys.CmdOrCtrl("7"), a.emitMenuPage("changelog"))
	viewMenu.AddSeparator()
	viewMenu.AddText("Settings", keys.CmdOrCtrl(","), a.emitMenuPage("settings"))
	viewMenu.AddSeparator()
	viewMenu.AddText("Toggle compact toolbar", keys.Combo("t", keys.CmdOrCtrlKey, keys.ShiftKey), a.emitMenu("menu:toggle-compact-toolbar"))

	projectMenu := appMenu.AddSubmenu("Project")
	projectMenu.AddText("Open command palette", keys.CmdOrCtrl("k"), a.emitMenu("menu:command-palette"))
	projectMenu.AddText("Open Authentication drawer", keys.Combo("a", keys.CmdOrCtrlKey, keys.ShiftKey), a.emitMenu("menu:auth-drawer"))
	projectMenu.AddSeparator()
	projectMenu.AddText("Start mock server", nil, a.emitMenu("menu:mock-start"))
	projectMenu.AddText("Stop mock server", nil, a.emitMenu("menu:mock-stop"))

	if runtime.GOOS == "darwin" {
		appMenu.Append(menu.WindowMenu())
	}

	return appMenu
}

func (a *App) emitMenu(event string) func(*menu.CallbackData) {
	return func(_ *menu.CallbackData) {
		if a.ctx == nil {
			return
		}
		wruntime.EventsEmit(a.ctx, event)
	}
}

func (a *App) emitMenuPage(page string) func(*menu.CallbackData) {
	return func(_ *menu.CallbackData) {
		if a.ctx == nil {
			return
		}
		wruntime.EventsEmit(a.ctx, "menu:navigate", page)
	}
}
