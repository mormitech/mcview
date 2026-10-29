package main

import "github.com/wailsapp/wails/v2/pkg/menu"

// The Effects menu: each item tells the page (menu-effect, with the name that EFFECT_MENU in main.js knows), which applies the effect to the part that is
// selected (a polygon or a box), or to all of the picture. The items are off while there is no picture.
func effectsMenu(app *App, m *menu.Menu) {
	fx := m.AddSubmenu("Effects")
	add := func(label, name string) { addEffectItem(app, fx, label, name) }
	add("Clear", "clear") // the selected part is filled with the second color (Delete or Backspace do the same)
	fx.AddSeparator()
	add("Blur", "blur")
	add("Pixelize", "pixelize")
	add("Emboss", "emboss")
	add("Explosion", "explosion")
	add("Censor", "censor")
	hl := fx.AddSubmenu("Highlight") // the selected part stays; the rest is pushed back, a little (light) to a lot (full: blurred and dimmed, with a shadow)
	addEffectItem(app, hl, "Light", "highlight-light")
	addEffectItem(app, hl, "Medium", "highlight-medium")
	addEffectItem(app, hl, "Full", "highlight-full")
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
}

// addEffectItem adds an item of the Effects menu (off while there is no picture).
func addEffectItem(app *App, m *menu.Menu, label, name string) {
	item := m.AddText(label, nil, func(*menu.CallbackData) { app.emit("menu-effect", name) })
	item.Disabled = true
	app.mu.Lock()
	app.effectItems = append(app.effectItems, item)
	app.mu.Unlock()
}

// The Image menu: turning and flipping the whole picture (menu-image, with the name that the page knows). Off while there is no picture, like the Effects menu.
func imageMenu(app *App, m *menu.Menu) {
	im := m.AddSubmenu("Image")
	add := func(label, name string) {
		item := im.AddText(label, nil, func(*menu.CallbackData) { app.emit("menu-image", name) })
		item.Disabled = true
		app.mu.Lock()
		app.effectItems = append(app.effectItems, item)
		app.mu.Unlock()
	}
	add("Rotate Right", "rotate-right")
	add("Rotate Left", "rotate-left")
	add("Rotate by Angle…", "rotate-angle")
	im.AddSeparator()
	add("Start Animation", "anim-start") // the picture becomes frame 1 of an animation (the frame player)
	add("Add Frame", "anim-add")         // a copy of the frame that is shown, after it
	add("Delete Frame", "anim-delete")
	add("Export Frames as PNG…", "anim-export")
	im.AddSeparator()
	add("Flip Horizontal", "flip-h")
	add("Flip Vertical", "flip-v")
}
