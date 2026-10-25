package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

// wails.json is the single source of the version: it ends up in the macOS Info.plist and the Windows file properties,
// and is read from here for the window title and the UI.
//
//go:embed wails.json
var projectJSON []byte

func version() string {
	var p struct {
		Info struct {
			ProductVersion string `json:"productVersion"`
		} `json:"info"`
	}
	if json.Unmarshal(projectJSON, &p) != nil || p.Info.ProductVersion == "" {
		return "dev"
	}
	return p.Info.ProductVersion
}

// windowFlags reads --about (show the About window instead of the viewer), --theme, --cascade, and the file to open (`mcview picture.jpg`,
// which is also how Windows and Linux hand over a double-clicked file). The app starts itself with --about, because Wails v2 has
// one window per process.
func windowFlags(args []string) (about bool, theme, file string, cascade int) {
	fs := flag.NewFlagSet("mcview", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.BoolVar(&about, "about", false, "show the About window")
	fs.StringVar(&theme, "theme", "", "light or dark, instead of the system's setting (for screenshots)")
	fs.IntVar(&cascade, "cascade", 0, "open this many steps down and to the right of the centre (used for the windows a window starts)")
	fs.Parse(args) // other arguments, such as the ones macOS adds when it launches an app, are ignored
	return about, theme, fileArg(fs.Args()), cascade
}

func main() {
	trace("main")
	app := NewApp()
	title, width, height := "mcView v"+version(), 1024, 640
	about, theme, file, cascade := windowFlags(os.Args[1:])
	if about {
		app.mode, title, width, height = "about", "About mcView", 440, 520
	} else if file != "" {
		app.child = true
		if err := app.setImage(file); err != nil {
			fmt.Fprintln(os.Stderr, "mcview:", err)
		}
	}
	if !about {
		if st, ok := loadWinState(); ok { // the window comes up the way it was left
			width, height = st.W, st.H
			app.saved = &st
		}
	}
	app.theme, app.cascade = theme, cascade
	app.winW, app.winH = width, height

	var appMenuBar *menu.Menu
	if !about { // the About window has no menu of its own
		appMenuBar = appMenu(app)
	}

	trace("wails.Run")
	err := wails.Run(&options.App{
		Menu:          appMenuBar,
		Title:         title,
		Width:         width,
		Height:        height,
		DisableResize: about,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: app.imageHandler(),
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		// Without Mac options Wails leaves the green zoom (maximise) button disabled. OnFileOpen is how macOS hands over the
		// files opened with the app.
		Mac:        macOptions(theme, app),
		Windows:    &windows.Options{Theme: windowsTheme(theme)},
		OnStartup:  app.startup,
		OnDomReady: func(context.Context) { trace("dom ready"); app.reveal() },
		OnShutdown: app.shutdown,
		OnBeforeClose: func(context.Context) bool {
			if !app.discardOK("close the window") {
				return true // the window stays open
			}
			app.SaveWindow()
			return false
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

// The window follows the system's light or dark appearance, unless --theme says otherwise (used for screenshots).
func macAppearance(theme string) mac.AppearanceType {
	switch theme {
	case "dark":
		return mac.NSAppearanceNameDarkAqua
	case "light":
		return mac.NSAppearanceNameAqua
	}
	return mac.DefaultAppearance
}

func windowsTheme(theme string) windows.Theme {
	switch theme {
	case "dark":
		return windows.Dark
	case "light":
		return windows.Light
	}
	return windows.SystemDefault
}

// macOptions: OnFileOpen is how macOS hands over the files opened with the app. MCVIEW_NO_FILEOPEN is for tracking down slow
// starts: it leaves the handler out.
func macOptions(theme string, app *App) *mac.Options {
	o := &mac.Options{Appearance: macAppearance(theme)}
	if os.Getenv("MCVIEW_NO_FILEOPEN") == "" {
		o.OnFileOpen = app.onFileOpen
	}
	return o
}
