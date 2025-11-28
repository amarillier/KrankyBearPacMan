//go:generate fyne bundle -o bundled.go -a Resources/Sounds/zip.mp3 Resources/Sounds/wheeHoo.mp3 Resources/Sounds/boing.mp3 Resources/Sounds/uhOh.mp3 Resources/Images/KrankyBearTrapperRedPlaid.png Resources/Images/KrankyBearHardHat.png

package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type appTheme struct {
	fyne.Theme
}

func (a *appTheme) Size(n fyne.ThemeSizeName) float32 {
	if n == theme.SizeNameHeadingText {
		return a.Theme.Size(n) * 1.5
	}

	return a.Theme.Size(n)
}

func newAppTheme() fyne.Theme {
	return &appTheme{Theme: theme.DefaultTheme()}
}

// "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
