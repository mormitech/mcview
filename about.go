package main

import (
	"runtime"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// appMenu is the menu of the main window, the way each system does it. On macOS that is the application menu (About,
// Hide, Quit) plus File, Edit and Window. On Windows and Linux there are File, Edit and Help menus (About mcView is in Help). About opens a
// window of its own, so that it can have real links (the system's message boxes cannot).
func appMenu(app *App) *menu.Menu {
	m := menu.NewMenu()
	if runtime.GOOS == "darwin" {
		appMenu := m.AddSubmenu("mcView")
		appMenu.AddText("About mcView", nil, func(*menu.CallbackData) { app.showAbout() })
		appMenu.AddSeparator()
		appMenu.AddText("Hide mcView", keys.CmdOrCtrl("h"), func(*menu.CallbackData) { wruntime.Hide(app.ctx) })
		appMenu.AddSeparator()
		appMenu.AddText("Quit mcView", keys.CmdOrCtrl("q"), func(*menu.CallbackData) { wruntime.Quit(app.ctx) })
		fileMenu(app, m)
		editMenu(app, m)
		m.Append(menu.WindowMenu())
		return m
	}
	fileMenu(app, m)
	editMenu(app, m)
	help := m.AddSubmenu("Help")
	help.AddText("About mcView", nil, func(*menu.CallbackData) { app.showAbout() })
	return m
}

func fileMenu(app *App, m *menu.Menu) {
	file := m.AddSubmenu("File")
	file.AddText("New Window", keys.CmdOrCtrl("n"), func(*menu.CallbackData) { app.newWindow() })
	file.AddText("Open File…", keys.CmdOrCtrl("o"), func(*menu.CallbackData) { app.chooseImage() })
	file.AddText("Create Image", nil, func(*menu.CallbackData) { app.emit("menu-create") }) // a new 1280 × 720 image, in edit mode
	file.AddText("Save As…", keys.CmdOrCtrl("s"), func(*menu.CallbackData) { app.saveAs() })
	file.AddText("Close Image", keys.Combo("w", keys.CmdOrCtrlKey, keys.ShiftKey), func(*menu.CallbackData) { app.emit("menu-close-image") })
	file.AddText("Delete Image…", nil, func(*menu.CallbackData) { app.emit("menu-delete") })        // (the key D does the same; no accelerator, so that a plain letter is not taken from the menu)
	file.AddText("Close", keys.CmdOrCtrl("w"), func(*menu.CallbackData) { app.emit("menu-close") }) // the picture, or the window when there is none (the UI knows)
	if runtime.GOOS != "darwin" {
		file.AddSeparator()
		file.AddText("Exit", keys.CmdOrCtrl("q"), func(*menu.CallbackData) { wruntime.Quit(app.ctx) })
	}
}

// The Edit menu is mcView's own, not the system's: the system's Copy and Paste work on text in the page, and a picture is not that.
func editMenu(app *App, m *menu.Menu) {
	edit := m.AddSubmenu("Edit")
	edit.AddText("Undo", keys.CmdOrCtrl("z"), func(*menu.CallbackData) { app.emit("menu-undo") })
	edit.AddText("Redo", keys.Combo("z", keys.CmdOrCtrlKey, keys.ShiftKey), func(*menu.CallbackData) { app.emit("menu-redo") })
	edit.AddSeparator()
	edit.AddText("Copy", keys.CmdOrCtrl("c"), func(*menu.CallbackData) { app.emit("menu-copy") })
	edit.AddText("Paste", keys.CmdOrCtrl("v"), func(*menu.CallbackData) { app.pasteImage() })
}
