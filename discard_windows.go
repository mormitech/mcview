package main

import (
	"golang.org/x/sys/windows"
)

// windowsDiscardChoice is the message box every Windows program shows for unsaved changes: Yes (save), No (do not), Cancel. Wails'
// own message box has Yes and No only.
func windowsDiscardChoice(title, text string) discardChoice {
	t, err1 := windows.UTF16PtrFromString(text)
	c, err2 := windows.UTF16PtrFromString(title)
	if err1 != nil || err2 != nil {
		return choiceCancel
	}
	res, _ := windows.MessageBox(windows.GetForegroundWindow(), t, c, windows.MB_YESNOCANCEL|windows.MB_ICONQUESTION|windows.MB_DEFBUTTON1|windows.MB_SETFOREGROUND)
	const idYes, idNo = 6, 7 // the answers of MessageBox
	switch res {
	case idYes:
		return choiceSave
	case idNo:
		return choiceDiscard
	}
	return choiceCancel
}
