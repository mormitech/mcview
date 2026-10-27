package main

import "github.com/wailsapp/wails/v2/pkg/menu"

// The Effects menu: each item tells the page (menu-effect, with the name that EFFECT_MENU in main.js knows), which applies the effect to the part that is
// selected (a polygon or a box), or to all of the picture. The items are off while there is no picture.
func effectsMenu(app *App, m *menu.Menu) {
	fx := m.AddSubmenu("Effects")
	add := func(label, name string) {
		item := fx.AddText(label, nil, func(*menu.CallbackData) { app.emit("menu-effect", name) })
		item.Disabled = true
		app.mu.Lock()
		app.effectItems = append(app.effectItems, item)
		app.mu.Unlock()
	}
	add("Clear", "clear") // the selected part is filled with the second color (Delete or Backspace do the same)
	fx.AddSeparator()
	add("Blur", "blur")
	add("Pixelize", "pixelize")
	add("Emboss", "emboss")
	add("Explosion", "explosion")
	add("Censor", "censor")
	add("Sharpen", "sharpen")
	fx.AddSeparator()
	add("Grayscale", "grayscale")
	add("Negative", "negative")
	add("Sepia", "sepia")
	fx.AddSeparator()
	add("Brighter", "brightness+")
	add("Darker", "brightness-")
	add("More Contrast", "contrast+")
	add("Less Contrast", "contrast-")
	add("More Color", "saturation+")
	add("Less Color", "saturation-")
	add("Warmer", "warmer")
	add("Cooler", "cooler")
	add("Auto Levels", "auto-levels")
	fx.AddSeparator()
	add("Fill with Foreground Color", "fill-fg")
	fx.AddSeparator()
	add("Rotate by Angle…", "rotate-angle")
}
