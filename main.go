package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2/app"
)

const (
	appVersion = "0.1.1" // see FyneApp.toml
	appAuthor  = "Allan Marillier"
)

var appName = "KrankyBear PacMan"
var appCopyright = "Copyright (c) Allan Marillier, 2025-" + strconv.Itoa(time.Now().Year())

// Theme type for character sprites
type Theme string

const (
	ThemeBasic           Theme = "basic"
	ThemeTraditional     Theme = "traditional"
	ThemeCamo            Theme = "camo"
	ThemeAnimal          Theme = "animal"
	ThemeFootball        Theme = "football"
	ThemeHogwarts        Theme = "hogwarts"
	ThemeCollegeFootball Theme = "collegefootball"
	ThemeHogwarts2       Theme = "hogwarts2"
	ThemeSuperheroes     Theme = "superheroes"
)

// currentTheme holds the selected theme
var currentTheme Theme = ThemeBasic

func main() {
	// Parse command line arguments
	themeFlag := flag.String("theme", "basic", "Character theme: basic, traditional, camo, animal, football, hogwarts, collegefootball, hogwarts2, superheroes")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nThemes:\n")
		fmt.Fprintf(os.Stderr, "  basic           - Simple colored shapes (default)\n")
		fmt.Fprintf(os.Stderr, "  traditional     - Classic PacMan and ghost images\n")
		fmt.Fprintf(os.Stderr, "  camo            - Camouflage-themed characters\n")
		fmt.Fprintf(os.Stderr, "  animal          - Wild animal characters\n")
		fmt.Fprintf(os.Stderr, "  football        - NFL team mascots\n")
		fmt.Fprintf(os.Stderr, "  hogwarts        - Harry Potter house themes\n")
		fmt.Fprintf(os.Stderr, "  collegefootball - College football team mascots\n")
		fmt.Fprintf(os.Stderr, "  hogwarts2       - Harry Potter house themes (alternate)\n")
		fmt.Fprintf(os.Stderr, "  superheroes     - Superhero characters\n")
	}
	flag.Parse()

	// Validate and set theme
	switch strings.ToLower(*themeFlag) {
	case "basic":
		currentTheme = ThemeBasic
	case "traditional":
		currentTheme = ThemeTraditional
	case "camo":
		currentTheme = ThemeCamo
	case "animal":
		currentTheme = ThemeAnimal
	case "football":
		currentTheme = ThemeFootball
	case "hogwarts":
		currentTheme = ThemeHogwarts
	case "collegefootball":
		currentTheme = ThemeCollegeFootball
	case "hogwarts2":
		currentTheme = ThemeHogwarts2
	case "superheroes":
		currentTheme = ThemeSuperheroes
	default:
		fmt.Fprintf(os.Stderr, "Invalid theme: %s. Valid options are: basic, traditional, camo, animal, football, hogwarts, collegefootball, hogwarts2, superheroes\n", *themeFlag)
		os.Exit(1)
	}
	// Create Fyne app
	myApp := app.NewWithID("com.github.amarillier.KrankyBearPacMan")
	
	// Set app icon
	myApp.SetIcon(resourceKrankyBearTrapperRedPlaidPng)
	
	// Set theme
	myApp.Settings().SetTheme(newAppTheme())
	
	// Create game instance
	game := NewGame()
	
	// Create UI
	ui := NewGameUI(myApp, game)
	
	// Set up game loop
	go func() {
		ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
		defer ticker.Stop()
		
		for range ticker.C {
			ui.Update()
		}
	}()
	
	// Show window and run app
	ui.ShowAndRun()
}

// "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
