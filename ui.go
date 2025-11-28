package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	updatechecker "github.com/amarillier/go-update-checker"
	"github.com/nfnt/resize"
)

// ShortcutBossKey describes the boss key shortcut (F12) - pause and hide window
type ShortcutBossKey struct{}

var _ fyne.KeyboardShortcut = (*ShortcutBossKey)(nil)

func (s *ShortcutBossKey) Key() fyne.KeyName {
	return fyne.KeyF12
}

func (s *ShortcutBossKey) Mod() fyne.KeyModifier {
	return fyne.KeyModifierShortcutDefault
}

func (s *ShortcutBossKey) ShortcutName() string {
	return "BossKey"
}

// GameUI manages the game's user interface
type GameUI struct {
	game           *Game
	app            fyne.App
	window         fyne.Window
	gameCanvas     *canvas.Raster
	scoreLabel     *widget.Label
	livesLabel     *widget.Label
	levelLabel     *widget.Label
	statusLabel    *widget.Label
	highScoreLabel *widget.Label
	highScoreList  *widget.Label
	startButton    *widget.Button
	pauseButton    *widget.Button
	soundButton    *widget.Button
	themeButton    *widget.Button
	highScoreMgr   *HighScoreManager
	soundMgr       *SoundManager
	resetDialog    fyne.Window
	aboutDialog    fyne.Window
	helpDialog     fyne.Window
	updateDialog   fyne.Window
	lastGameState  GameState
	scoreAdded     bool
	showMenuItem   *fyne.MenuItem
	hideMenuItem   *fyne.MenuItem
	windowVisible  bool
	versionStatus  string
	lastUpdate     time.Time
	// Theme-related cached images (scaled for game display)
	pacmanImage    image.Image
	blinkyImage    image.Image
	pinkyImage     image.Image
	inkyImage      image.Image
	clydeImage     image.Image
	frightenedImg  image.Image // Blue ghost for frightened mode
	// Cached help dialog label for current theme
	helpCurrentThemeLabel *widget.Label
	// Cached game frame buffer to reduce allocations and flicker
	frameBuffer   *image.RGBA
	frameBufferW  int
	frameBufferH  int
}

const lifeCheatCost = 500

// NewGameUI creates a new game UI
func NewGameUI(app fyne.App, game *Game) *GameUI {
	ui := &GameUI{
		game:          game,
		app:           app,
		window:        app.NewWindow(appName),
		highScoreMgr:  NewHighScoreManager(app),
		soundMgr:      NewSoundManager(),
		windowVisible: true,
		lastUpdate:    time.Now(),
	}

	// Load theme preference (command line takes precedence, otherwise use saved preference)
	// Only load from preferences if command line was default (basic)
	if currentTheme == ThemeBasic {
		savedTheme := app.Preferences().StringWithFallback("theme", "basic")
		switch savedTheme {
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
		}
	}

	// Load theme images
	ui.loadThemeImages()

	// Load sound preference (default to true if not set)
	soundEnabled := app.Preferences().BoolWithFallback("sound_enabled", true)
	ui.soundMgr.SetEnabled(soundEnabled)

	ui.setupUI()
	ui.setupKeyboard()
	ui.setupMenu()

	// Set up sound callback
	game.SetSoundCallback(ui.handleSoundEvent)

	// Set close intercept to close all dialogs and quit app
	ui.window.SetCloseIntercept(func() {
		// Close all dialog windows (use Close() to fully destroy them)
		if ui.helpDialog != nil {
			ui.helpDialog.Close()
			ui.helpDialog = nil
			ui.helpCurrentThemeLabel = nil
		}
		if ui.aboutDialog != nil {
			ui.aboutDialog.Close()
		}
		if ui.updateDialog != nil {
			ui.updateDialog.Close()
		}
		if ui.resetDialog != nil {
			ui.resetDialog.Close()
		}
		// Quit the application
		ui.app.Quit()
	})

	// Check for updates on startup
	ui.checkForUpdates()

	// Setup system tray
	ui.setupSystemTray()

	return ui
}

// spriteSize is the size for themed sprites (larger than basic circles)
const spriteSize = 54 // 3x original size for better visibility

// loadThemeImages loads and scales theme images for game display
func (ui *GameUI) loadThemeImages() {
	if currentTheme == ThemeBasic {
		// Clear images for basic theme
		ui.pacmanImage = nil
		ui.blinkyImage = nil
		ui.pinkyImage = nil
		ui.inkyImage = nil
		ui.clydeImage = nil
		ui.frightenedImg = nil
		return
	}

	var pacmanRes, blinkyRes, pinkyRes, inkyRes, clydeRes *fyne.StaticResource

	switch currentTheme {
	case ThemeTraditional:
		pacmanRes = resourceTraditionalPacManPng
		blinkyRes = resourceTraditionalBlinkyPng
		pinkyRes = resourceTraditionalPinkyPng
		inkyRes = resourceTraditionalInkyPng
		clydeRes = resourceTraditionalClydePng
	case ThemeCamo:
		pacmanRes = resourceCamoPacManPng
		blinkyRes = resourceCamoBlinkyPng
		pinkyRes = resourceCamoPinkyPng
		inkyRes = resourceCamoInkyPng
		clydeRes = resourceCamoClydePng
	case ThemeAnimal:
		pacmanRes = resourceAnimalPacManPng
		blinkyRes = resourceAnimalBlinkyPng
		pinkyRes = resourceAnimalPinkyPng
		inkyRes = resourceAnimalInkyPng
		clydeRes = resourceAnimalClydePng
	case ThemeFootball:
		pacmanRes = resourceFootballPacManPng
		blinkyRes = resourceFootballBlinkyPng
		pinkyRes = resourceFootballPinkyPng
		inkyRes = resourceFootballInkyPng
		clydeRes = resourceFootballClydePng
	case ThemeHogwarts:
		pacmanRes = resourceHogwartsPacManPng
		blinkyRes = resourceHogwartsBlinkyPng
		pinkyRes = resourceHogwartsPinkyPng
		inkyRes = resourceHogwartsInkyPng
		clydeRes = resourceHogwartsClydePng
	case ThemeCollegeFootball:
		pacmanRes = resourceCollegeFootballPacManPng
		blinkyRes = resourceCollegeFootballBlinkyPng
		pinkyRes = resourceCollegeFootballPinkyPng
		inkyRes = resourceCollegeFootballInkyPng
		clydeRes = resourceCollegeFootballClydePng
	case ThemeHogwarts2:
		pacmanRes = resourceHogwarts2PacManPng
		blinkyRes = resourceHogwarts2BlinkyPng
		pinkyRes = resourceHogwarts2PinkyPng
		inkyRes = resourceHogwarts2InkyPng
		clydeRes = resourceHogwarts2ClydePng
	case ThemeSuperheroes:
		pacmanRes = resourceSuperheroesPacManPng
		blinkyRes = resourceSuperheroesBlinkyPng
		pinkyRes = resourceSuperheroesPinkyPng
		inkyRes = resourceSuperheroesInkyPng
		clydeRes = resourceSuperheroesClydePng
	}

	ui.pacmanImage = loadAndScaleImage(pacmanRes, spriteSize)
	ui.blinkyImage = loadAndScaleImage(blinkyRes, spriteSize)
	ui.pinkyImage = loadAndScaleImage(pinkyRes, spriteSize)
	ui.inkyImage = loadAndScaleImage(inkyRes, spriteSize)
	ui.clydeImage = loadAndScaleImage(clydeRes, spriteSize)

	// Create a blue-tinted version of pacman for frightened ghosts
	ui.frightenedImg = createFrightenedImage(spriteSize)
}

// cycleTheme cycles through available themes
func (ui *GameUI) cycleTheme() {
	// Order: basic, traditional, then alphabetical (animal, camo, collegefootball, football, hogwarts, hogwarts2, superheroes)
	switch currentTheme {
	case ThemeBasic:
		currentTheme = ThemeTraditional
	case ThemeTraditional:
		currentTheme = ThemeAnimal
	case ThemeAnimal:
		currentTheme = ThemeCamo
	case ThemeCamo:
		currentTheme = ThemeCollegeFootball
	case ThemeCollegeFootball:
		currentTheme = ThemeFootball
	case ThemeFootball:
		currentTheme = ThemeHogwarts
	case ThemeHogwarts:
		currentTheme = ThemeHogwarts2
	case ThemeHogwarts2:
		currentTheme = ThemeSuperheroes
	case ThemeSuperheroes:
		currentTheme = ThemeBasic
	}

	// Reload theme images
	ui.loadThemeImages()

	// Update button text
	ui.themeButton.SetText("Theme: " + string(currentTheme))

	// Save theme preference
	ui.app.Preferences().SetString("theme", string(currentTheme))

	// Refresh the game canvas to show new sprites
	canvas.Refresh(ui.gameCanvas)
}

// loadAndScaleImage loads a PNG resource and scales it to the given size
func loadAndScaleImage(res *fyne.StaticResource, size int) image.Image {
	img, err := png.Decode(bytes.NewReader(res.StaticContent))
	if err != nil {
		return nil
	}
	// Use Lanczos3 for best quality scaling
	return resize.Resize(uint(size), uint(size), img, resize.Lanczos3)
}

// createFrightenedImage creates a blue circle for frightened mode
func createFrightenedImage(size int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	blueColor := color.RGBA{0, 0, 255, 255}
	radius := size / 2
	cx, cy := size/2, size/2

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := x - cx
			dy := y - cy
			if dx*dx+dy*dy <= radius*radius {
				img.Set(x, y, blueColor)
			}
		}
	}
	return img
}

// drawImageAt draws a scaled image centered at the given pixel coordinates
func drawImageAt(dst *image.RGBA, src image.Image, cx, cy int) {
	if src == nil {
		return
	}
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	startX := cx - w/2
	startY := cy - h/2

	// Use draw.Draw for proper alpha blending
	draw.Draw(dst, image.Rect(startX, startY, startX+w, startY+h), src, bounds.Min, draw.Over)
}

// setupUI creates the UI elements
func (ui *GameUI) setupUI() {
	// Create game canvas
	ui.gameCanvas = canvas.NewRaster(ui.drawGame)
	ui.gameCanvas.Resize(fyne.NewSize(WindowWidth, WindowHeight))

	// Create labels
	ui.scoreLabel = widget.NewLabel("Score: 0")
	ui.scoreLabel.TextStyle.Bold = true
	ui.levelLabel = widget.NewLabel("Level: 1")
	ui.levelLabel.TextStyle.Bold = true

	ui.livesLabel = widget.NewLabel("Lives: 3")
	ui.livesLabel.TextStyle.Bold = true

	ui.statusLabel = widget.NewLabel("Press SPACE to Start")
	ui.statusLabel.TextStyle.Bold = true
	ui.statusLabel.Wrapping = fyne.TextWrapWord

	ui.highScoreLabel = widget.NewLabel("Best: 0")
	ui.highScoreLabel.TextStyle.Bold = true

	ui.highScoreList = widget.NewLabel("No scores yet.")
	ui.highScoreList.Wrapping = fyne.TextWrapWord

	// Create buttons
	ui.startButton = widget.NewButton("Start", func() {
		if ui.game != nil && ui.game.state == StateMenu {
			ui.game.Start()
			canvas.Refresh(ui.gameCanvas)
		}
	})

	ui.pauseButton = widget.NewButton("Pause", func() {
		if ui.game.state == StatePlaying {
			ui.game.state = StatePaused
		} else if ui.game.state == StatePaused {
			ui.game.state = StatePlaying
		}
	})

	// Initialize sound button with correct state from preferences
	soundText := "Sound: ON"
	if !ui.soundMgr.IsEnabled() {
		soundText = "Sound: OFF"
	}
	ui.soundButton = widget.NewButton(soundText, func() {
		ui.soundMgr.Toggle()
		if ui.soundMgr.IsEnabled() {
			ui.soundButton.SetText("Sound: ON")
		} else {
			ui.soundButton.SetText("Sound: OFF")
		}
		// Save sound preference
		ui.app.Preferences().SetBool("sound_enabled", ui.soundMgr.IsEnabled())
	})

	// Initialize theme button
	ui.themeButton = widget.NewButton("Theme: "+string(currentTheme), func() {
		ui.cycleTheme()
	})

	resetScoresButton := widget.NewButton("Reset High Scores", ui.showResetHighScoresDialog)

	controlsTitle := widget.NewLabel("Controls")
	controlsTitle.TextStyle = fyne.TextStyle{Bold: true}

	instructions := widget.NewLabel("Arrow Keys: move Pacman\nSpace: start game\nP: pause/resume\nR: restart (after game over)\nZ: +1 life (-500 pts)\nF12: boss key (hide and pause)")
	instructions.Wrapping = fyne.TextWrapWord

	statusTitle := widget.NewLabel("Status")
	statusTitle.TextStyle = fyne.TextStyle{Bold: true}

	buttonColumn := container.NewVBox(
		ui.startButton,
		ui.pauseButton,
		ui.soundButton,
		ui.themeButton,
	)

	buttonGroupTitle := widget.NewLabel("Actions")
	buttonGroupTitle.TextStyle = fyne.TextStyle{Bold: true}

	highScoreTitle := widget.NewLabel("High Scores (Top 5)")
	highScoreTitle.TextStyle = fyne.TextStyle{Bold: true}

	controlPanel := container.NewVBox(
		statusTitle,
		ui.statusLabel,
		widget.NewSeparator(),
		ui.scoreLabel,
		ui.livesLabel,
		ui.levelLabel,
		widget.NewSeparator(),
		buttonGroupTitle,
		buttonColumn,
		widget.NewSeparator(),
		controlsTitle,
		instructions,
		widget.NewSeparator(),
		ui.highScoreLabel,
		highScoreTitle,
		ui.highScoreList,
		resetScoresButton,
		layout.NewSpacer(),
	)

	controlScroll := container.NewVScroll(controlPanel)
	controlScroll.SetMinSize(fyne.NewSize(260, WindowHeight))

	mainSplit := container.NewHSplit(
		container.NewMax(ui.gameCanvas),
		controlScroll,
	)
	mainSplit.SetOffset(0.72)

	ui.window.SetContent(mainSplit)
	ui.window.Resize(fyne.NewSize(WindowWidth+320, WindowHeight+80))
	ui.window.CenterOnScreen()
	ui.window.RequestFocus() // Ensure window has focus for keyboard input

	ui.refreshHighScoreDisplay()
}

// setupKeyboard sets up keyboard shortcuts
func (ui *GameUI) setupKeyboard() {
	// Arrow keys for movement
	ui.window.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyUp,
		Modifier: 0,
	}, func(shortcut fyne.Shortcut) {
		if ui.game.state == StatePlaying {
			ui.game.player.SetDirection(Up)
		}
	})

	ui.window.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyDown,
		Modifier: 0,
	}, func(shortcut fyne.Shortcut) {
		if ui.game.state == StatePlaying {
			ui.game.player.SetDirection(Down)
		}
	})

	ui.window.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyLeft,
		Modifier: 0,
	}, func(shortcut fyne.Shortcut) {
		if ui.game.state == StatePlaying {
			ui.game.player.SetDirection(Left)
		}
	})

	ui.window.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyRight,
		Modifier: 0,
	}, func(shortcut fyne.Shortcut) {
		if ui.game.state == StatePlaying {
			ui.game.player.SetDirection(Right)
		}
	})

	// Space to start
	ui.window.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeySpace,
		Modifier: 0,
	}, func(shortcut fyne.Shortcut) {
		if ui.game != nil && ui.game.state == StateMenu {
			ui.game.Start()
			canvas.Refresh(ui.gameCanvas)
		}
	})

	// P to pause
	ui.window.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyP,
		Modifier: 0,
	}, func(shortcut fyne.Shortcut) {
		if ui.game != nil {
			if ui.game.state == StatePlaying {
				ui.game.state = StatePaused
				canvas.Refresh(ui.gameCanvas)
			} else if ui.game.state == StatePaused {
				ui.game.state = StatePlaying
				canvas.Refresh(ui.gameCanvas)
			}
		}
	})

	// R to restart
	ui.window.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyR,
		Modifier: 0,
	}, func(shortcut fyne.Shortcut) {
		if ui.game.state == StateGameOver || ui.game.state == StateWin {
			ui.game = NewGame()
			ui.game.SetSoundCallback(ui.handleSoundEvent)
		}
	})

	// Handle all keyboard input via SetOnTypedKey
	ui.window.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
		// Boss key (F12)
		if ev.Name == fyne.KeyF12 {
			ui.toggleBossKey()
			return
		}
		// Space to start
		if ev.Name == fyne.KeySpace {
			if ui.game != nil && ui.game.state == StateMenu {
				ui.game.Start()
				canvas.Refresh(ui.gameCanvas)
			}
			return
		}
		// P to pause/resume
		if ev.Name == fyne.KeyP {
			if ui.game != nil {
				if ui.game.state == StatePlaying {
					ui.game.state = StatePaused
					canvas.Refresh(ui.gameCanvas)
				} else if ui.game.state == StatePaused {
					ui.game.state = StatePlaying
					canvas.Refresh(ui.gameCanvas)
				}
			}
			return
		}
		// Z cheat key
		if ev.Name == fyne.KeyZ {
			ui.applyLifeCheat()
			return
		}
		// Arrow keys for movement
		if ui.game != nil && ui.game.state == StatePlaying {
			switch ev.Name {
			case fyne.KeyUp:
				ui.game.player.SetDirection(Up)
			case fyne.KeyDown:
				ui.game.player.SetDirection(Down)
			case fyne.KeyLeft:
				ui.game.player.SetDirection(Left)
			case fyne.KeyRight:
				ui.game.player.SetDirection(Right)
			}
		}
		// R to restart
		if ev.Name == fyne.KeyR {
			if ui.game != nil && (ui.game.state == StateGameOver || ui.game.state == StateWin) {
				ui.game = NewGame()
				ui.game.SetSoundCallback(ui.handleSoundEvent)
			}
			return
		}
	})

	// Also register shortcuts for consistency (though SetOnTypedKey handles them)
	ui.window.Canvas().AddShortcut(&ShortcutBossKey{}, func(shortcut fyne.Shortcut) {
		ui.toggleBossKey()
	})

	// Cheat key (Z)
	ui.window.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyZ,
		Modifier: 0,
	}, func(shortcut fyne.Shortcut) {
		ui.applyLifeCheat()
	})
}

// setupMenu sets up the menu bar
func (ui *GameUI) setupMenu() {
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("New Game", func() {
			ui.game = NewGame()
			ui.game.SetSoundCallback(ui.handleSoundEvent)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() {
			ui.app.Quit()
		}),
	)

	viewMenu := fyne.NewMenu("View",
		fyne.NewMenuItem("Show Window", func() {
			fyne.Do(ui.showWindow)
		}),
		fyne.NewMenuItem("Hide Window", func() {
			fyne.Do(ui.hideWindow)
		}),
	)

	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Help", func() {
			fyne.Do(ui.showHelp)
		}),
		fyne.NewMenuItem("About", func() {
			fyne.Do(ui.showAbout)
		}),
		fyne.NewMenuItem("Check for Updates", func() {
			ui.showUpdateDialog()
		}),
	)

	ui.showMenuItem = viewMenu.Items[0]
	ui.hideMenuItem = viewMenu.Items[1]

	mainMenu := fyne.NewMainMenu(fileMenu, viewMenu, helpMenu)
	ui.window.SetMainMenu(mainMenu)

	// Set window icon
	ui.window.SetIcon(resourceKrankyBearTrapperRedPlaidPng)
}

// setupSystemTray sets up the system tray icon and menu using desktop.App interface
func (ui *GameUI) setupSystemTray() {
	// Check if app supports desktop features (system tray)
	if desk, ok := ui.app.(desktop.App); ok {
		// Create menu items
		ui.showMenuItem = fyne.NewMenuItem("Show", func() {
			fyne.Do(ui.showWindow)
		})
		ui.hideMenuItem = fyne.NewMenuItem("Hide", func() {
			fyne.Do(ui.hideWindow)
		})
		helpTray := fyne.NewMenuItem("Help", func() {
			fyne.Do(ui.showHelp)
		})
		aboutTray := fyne.NewMenuItem("About", func() {
			fyne.Do(ui.showAbout)
		})
		updateTray := fyne.NewMenuItem("Check for Update", func() {
			ui.showUpdateDialog()
		})
		quitTray := fyne.NewMenuItem("Quit", func() {
			fyne.Do(ui.app.Quit)
		})

		// Create menu
		menu := fyne.NewMenu(appName,
			ui.showMenuItem,
			ui.hideMenuItem,
			fyne.NewMenuItemSeparator(),
			aboutTray,
			helpTray,
			updateTray,
			fyne.NewMenuItemSeparator(),
			quitTray,
		)

		// Set system tray menu and icon
		desk.SetSystemTrayMenu(menu)
		desk.SetSystemTrayIcon(resourceKrankyBearTrapperRedPlaidPng)

		// Update menu state based on window visibility
		ui.updateTrayMenuState()
	}
}

// updateTrayMenuState updates the tray menu items based on window visibility
func (ui *GameUI) updateTrayMenuState() {
	// Recreate the menu with updated state
	if desk, ok := ui.app.(desktop.App); ok {
		// Create menu items
		aboutTray := fyne.NewMenuItem("About", func() {
			fyne.Do(ui.showAbout)
		})
		helpTray := fyne.NewMenuItem("Help", func() {
			fyne.Do(ui.showHelp)
		})
		updateTray := fyne.NewMenuItem("Check for Update", func() {
			ui.showUpdateDialog()
		})

		// Create show/hide menu items - always create both, but we'll handle logic in callbacks
		ui.showMenuItem = fyne.NewMenuItem("Show", func() {
			fyne.Do(func() {
				if !ui.windowVisible {
					ui.showWindow()
				}
			})
		})
		ui.hideMenuItem = fyne.NewMenuItem("Hide", func() {
			fyne.Do(func() {
				if ui.windowVisible {
					ui.hideWindow()
				}
			})
		})

		quitTray := fyne.NewMenuItem("Quit", func() {
			fyne.Do(ui.app.Quit)
		})

		// Create menu
		menu := fyne.NewMenu(appName,
			aboutTray,
			helpTray,
			updateTray,
			fyne.NewMenuItemSeparator(),
			ui.showMenuItem,
			ui.hideMenuItem,
			fyne.NewMenuItemSeparator(),
			quitTray,
		)

		// Update system tray menu
		desk.SetSystemTrayMenu(menu)
	}
}

// showWindow shows the main window
func (ui *GameUI) showWindow() {
	ui.window.Show()
	ui.windowVisible = true
	ui.updateTrayMenuState()
}

// hideWindow hides the main window
func (ui *GameUI) hideWindow() {
	ui.window.Hide()
	ui.windowVisible = false
	if ui.game.state == StatePlaying {
		ui.game.state = StatePaused
	}
	ui.updateTrayMenuState()
}

// drawGame draws the game
func (ui *GameUI) drawGame(w, h int) image.Image {
	// Reuse frame buffer if dimensions match, otherwise create new one
	if ui.frameBuffer == nil || ui.frameBufferW != w || ui.frameBufferH != h {
		ui.frameBuffer = image.NewRGBA(image.Rect(0, 0, w, h))
		ui.frameBufferW = w
		ui.frameBufferH = h
	}

	img := ui.frameBuffer

	// Fill background with black using fast draw.Draw
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{0, 0, 0, 255}}, image.Point{}, draw.Src)

	// Draw maze
	ui.drawMaze(img, w, h)

	// Draw dots and power pellets
	ui.drawDots(img, w, h)

	// Draw player
	ui.drawPlayer(img, w, h)

	// Draw ghosts
	ui.drawGhosts(img, w, h)

	return img
}

// drawMaze draws the maze walls
func (ui *GameUI) drawMaze(img *image.RGBA, w, h int) {
	wallColor := color.RGBA{0, 0, 255, 255}

	for y := 0; y < MazeHeight; y++ {
		for x := 0; x < MazeWidth; x++ {
			if ui.game.maze[y][x] == Wall {
				drawTile(img, x, y, wallColor, w, h)
			}
		}
	}
}

// drawDots draws dots and power pellets
func (ui *GameUI) drawDots(img *image.RGBA, w, h int) {
	dotColor := color.RGBA{255, 255, 255, 255}
	powerPelletColor := color.RGBA{0, 255, 0, 255} // Bright green for power pellets

	for y := 0; y < MazeHeight; y++ {
		for x := 0; x < MazeWidth; x++ {
			if ui.game.maze[y][x] == Dot {
				drawDot(img, x, y, dotColor, w, h)
			} else if ui.game.maze[y][x] == PowerPellet {
				drawPowerPellet(img, x, y, powerPelletColor, w, h)
			}
		}
	}
}

// drawPlayer draws Pacman
func (ui *GameUI) drawPlayer(img *image.RGBA, w, h int) {
	playerPos := ui.game.player.GetPosition()

	// Calculate tile size in pixels (use integer division like dots)
	tileW := w / MazeWidth
	tileH := h / MazeHeight

	// Use the same centering logic as dots - center in the tile
	// Convert float position to pixel position within tile, then center
	gridX := int(math.Round(playerPos.X))
	gridY := int(math.Round(playerPos.Y))

	// Calculate offset within the tile (0.0 to 1.0)
	offsetX := playerPos.X - float64(gridX)
	offsetY := playerPos.Y - float64(gridY)

	// Center of the tile
	centerX := gridX*tileW + tileW/2
	centerY := gridY*tileH + tileH/2

	// Add offset within tile
	px := centerX + int(offsetX*float64(tileW))
	py := centerY + int(offsetY*float64(tileH))

	// Draw themed image or basic circle
	if currentTheme != ThemeBasic && ui.pacmanImage != nil {
		drawImageAt(img, ui.pacmanImage, px, py)
	} else {
		playerColor := color.RGBA{255, 255, 0, 255}
		size := int(float64(TileSize) * 0.7) // 70% of tile (14)
		drawCircle(img, px, py, size, playerColor)
	}
}

// drawGhosts draws all ghosts
func (ui *GameUI) drawGhosts(img *image.RGBA, w, h int) {
	for _, ghost := range ui.game.ghosts {
		ui.drawGhost(img, ghost, w, h)
	}
}

// drawGhost draws a single ghost
func (ui *GameUI) drawGhost(img *image.RGBA, ghost *Ghost, w, h int) {
	ghostPos := ghost.position

	// Calculate tile size in pixels (use integer division like dots)
	tileW := w / MazeWidth
	tileH := h / MazeHeight

	// Use the same centering logic as dots - center in the tile
	// Convert float position to pixel position within tile, then center
	gridX := int(math.Round(ghostPos.X))
	gridY := int(math.Round(ghostPos.Y))

	// Calculate offset within the tile (0.0 to 1.0)
	offsetX := ghostPos.X - float64(gridX)
	offsetY := ghostPos.Y - float64(gridY)

	// Center of the tile
	centerX := gridX*tileW + tileW/2
	centerY := gridY*tileH + tileH/2

	// Add offset within tile
	px := centerX + int(offsetX*float64(tileW))
	py := centerY + int(offsetY*float64(tileH))

	// Draw themed image or basic circle
	if currentTheme != ThemeBasic {
		var ghostImg image.Image
		if ghost.mode == ModeFrightened {
			ghostImg = ui.frightenedImg
		} else {
			switch ghost.color {
			case Red:
				ghostImg = ui.blinkyImage
			case Pink:
				ghostImg = ui.pinkyImage
			case Cyan:
				ghostImg = ui.inkyImage
			case Orange:
				ghostImg = ui.clydeImage
			}
		}
		if ghostImg != nil {
			drawImageAt(img, ghostImg, px, py)
			return
		}
	}

	// Fallback to basic circles
	var ghostColor color.RGBA
	if ghost.mode == ModeFrightened {
		ghostColor = color.RGBA{0, 0, 255, 255}
	} else {
		switch ghost.color {
		case Red:
			ghostColor = color.RGBA{255, 0, 0, 255}
		case Pink:
			ghostColor = color.RGBA{255, 192, 203, 255}
		case Cyan:
			ghostColor = color.RGBA{0, 255, 255, 255}
		case Orange:
			ghostColor = color.RGBA{255, 165, 0, 255}
		}
	}
	size := int(float64(TileSize) * 0.7) // 70% of tile (14)
	drawCircle(img, px, py, size, ghostColor)
}

// drawTile draws a wall tile
func drawTile(img *image.RGBA, x, y int, c color.RGBA, w, h int) {
	tileW := w / MazeWidth
	tileH := h / MazeHeight
	startX := x * tileW
	startY := y * tileH

	for dy := 0; dy < tileH; dy++ {
		for dx := 0; dx < tileW; dx++ {
			if startX+dx < w && startY+dy < h {
				img.Set(startX+dx, startY+dy, c)
			}
		}
	}
}

// drawDot draws a small dot
func drawDot(img *image.RGBA, x, y int, c color.RGBA, w, h int) {
	tileW := w / MazeWidth
	tileH := h / MazeHeight
	centerX := x*tileW + tileW/2
	centerY := y*tileH + tileH/2
	radius := 2

	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= radius*radius {
				px := centerX + dx
				py := centerY + dy
				if px >= 0 && px < w && py >= 0 && py < h {
					img.Set(px, py, c)
				}
			}
		}
	}
}

// drawPowerPellet draws a power pellet
func drawPowerPellet(img *image.RGBA, x, y int, c color.RGBA, w, h int) {
	tileW := w / MazeWidth
	tileH := h / MazeHeight
	centerX := x*tileW + tileW/2
	centerY := y*tileH + tileH/2
	radius := 6

	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= radius*radius {
				px := centerX + dx
				py := centerY + dy
				if px >= 0 && px < w && py >= 0 && py < h {
					img.Set(px, py, c)
				}
			}
		}
	}
}

// drawCircle draws a filled circle
func drawCircle(img *image.RGBA, cx, cy, radius int, c color.RGBA) {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= radius*radius {
				px := cx + dx
				py := cy + dy
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					img.Set(px, py, c)
				}
			}
		}
	}
}

// Update updates the UI
func (ui *GameUI) Update() {
	now := time.Now()
	elapsed := now.Sub(ui.lastUpdate)

	// Update game logic at 60 FPS
	gameUpdated := false
	if elapsed >= time.Second/60 {
		if ui.game != nil {
			ui.game.Update()
			gameUpdated = true
		}
		ui.lastUpdate = now
	}

	// Wrap all UI updates in fyne.Do() for thread-safety when called from goroutine
	fyne.Do(func() {
		if ui.game == nil {
			return
		}

		// Check if state changed
		stateChanged := ui.lastGameState != ui.game.state

		// Always update labels to show current score/lives/level (score changes happen in game.Update)
		ui.scoreLabel.SetText("Score: " + strconv.Itoa(ui.game.score))
		ui.livesLabel.SetText("Lives: " + strconv.Itoa(ui.game.lives))
		ui.levelLabel.SetText("Level: " + strconv.Itoa(ui.game.level))

		// Update status label and high scores (only when game state changes)
		if gameUpdated {
			ui.refreshHighScoreDisplay()

			if stateChanged {
				ui.scoreAdded = false
			}

			// Update status label
			switch ui.game.state {
			case StateMenu:
				ui.statusLabel.SetText("Press SPACE to Start")
				ui.scoreAdded = false // Reset when returning to menu
			case StatePlaying:
				ui.statusLabel.SetText("")
			case StatePaused:
				ui.statusLabel.SetText("PAUSED - Press P to Resume")
			case StateGameOver:
				ui.statusLabel.SetText("GAME OVER - Press R to Restart")
				if !ui.scoreAdded {
					if ui.highScoreMgr.AddScore(ui.game.score) {
						ui.soundMgr.PlaySound(SoundEventHighScore)
						ui.refreshHighScoreDisplay()
					}
					ui.scoreAdded = true
				}
			case StateWin:
				ui.statusLabel.SetText("YOU WIN! - Press R to Restart")
				if !ui.scoreAdded {
					if ui.highScoreMgr.AddScore(ui.game.score) {
						ui.soundMgr.PlaySound(SoundEventHighScore)
						ui.refreshHighScoreDisplay()
					}
					ui.scoreAdded = true
				}
			}

			ui.lastGameState = ui.game.state
		}

		// Only refresh canvas when game is actively playing or state just changed
		// This reduces flicker during menu/pause states
		if ui.game.state == StatePlaying || stateChanged {
			canvas.Refresh(ui.gameCanvas)
		}
	})
}

// ShowAndRun shows the window and runs the app
func (ui *GameUI) ShowAndRun() {
	ui.window.ShowAndRun()
}

// Show shows the window
func (ui *GameUI) Show() {
	ui.window.Show()
}

// handleSoundEvent handles sound events from the game
func (ui *GameUI) handleSoundEvent(event SoundEvent) {
	ui.soundMgr.PlaySound(event)
}

// toggleBossKey toggles boss key (hide/show window)
func (ui *GameUI) toggleBossKey() {
	// Pause the game if it's currently playing
	if ui.game.state == StatePlaying {
		ui.game.state = StatePaused
	}

	if ui.windowVisible {
		ui.hideWindow()
	} else {
		ui.showWindow()
	}
}

// positionDialogRelativeToMain positions a dialog window relative to the main window
// Uses the same pattern as KrankyBearTetris: show first, then center
// This ensures the dialog appears on the same display as the main window
func (ui *GameUI) positionDialogRelativeToMain(dialogWindow fyne.Window) {
	// Center the dialog on the screen (window should already be shown before calling this)
	dialogWindow.CenterOnScreen()
}

func (ui *GameUI) refreshHighScoreDisplay() {
	if ui.highScoreMgr == nil || ui.highScoreLabel == nil || ui.highScoreList == nil {
		return
	}

	scores := ui.highScoreMgr.GetHighScores()
	if len(scores) == 0 {
		ui.highScoreLabel.SetText("Best: 0")
		ui.highScoreList.SetText("No scores yet. Finish a round to record one.")
		return
	}

	ui.highScoreLabel.SetText(fmt.Sprintf("Best: %d", scores[0]))

	var builder strings.Builder
	for i, score := range scores {
		fmt.Fprintf(&builder, "%d. %d\n", i+1, score)
	}
	ui.highScoreList.SetText(strings.TrimSpace(builder.String()))
}

func (ui *GameUI) showResetHighScoresDialog() {
	if ui.resetDialog != nil {
		ui.resetDialog.RequestFocus()
		ui.resetDialog.Show()
		return
	}

	dialogWindow := ui.app.NewWindow("Reset High Scores")
	dialogWindow.Resize(fyne.NewSize(420, 200))
	dialogWindow.SetFixedSize(true)
	ui.resetDialog = dialogWindow
	dialogWindow.SetOnClosed(func() {
		ui.resetDialog = nil
	})

	message := widget.NewLabel("Type \"KrankyBear\" to confirm resetting the leaderboard. This cannot be undone.")
	message.Wrapping = fyne.TextWrapWord

	entry := widget.NewEntry()
	entry.SetPlaceHolder("KrankyBear")

	confirmButton := widget.NewButton("Reset", func() {
		ui.highScoreMgr.ResetHighScores()
		ui.refreshHighScoreDisplay()
		dialogWindow.Close()
	})
	confirmButton.Disable()

	cancelButton := widget.NewButton("Cancel", func() {
		dialogWindow.Close()
	})

	entry.OnChanged = func(text string) {
		if text == "KrankyBear" {
			confirmButton.Enable()
		} else {
			confirmButton.Disable()
		}
	}

	content := container.NewVBox(
		message,
		entry,
		container.NewHBox(
			layout.NewSpacer(),
			cancelButton,
			confirmButton,
		),
	)

	dialogWindow.SetContent(content)
	ui.positionDialogRelativeToMain(dialogWindow)
}

func (ui *GameUI) applyLifeCheat() {
	if ui.game == nil {
		return
	}
	if ui.game.state != StatePlaying && ui.game.state != StatePaused {
		return
	}
	if ui.game.score < lifeCheatCost {
		ui.statusLabel.SetText(fmt.Sprintf("Need %d points for an extra life.", lifeCheatCost))
		return
	}
	ui.game.score -= lifeCheatCost
	ui.game.lives++
	ui.statusLabel.SetText(fmt.Sprintf("Cheat: +1 life, -%d points.", lifeCheatCost))
}

// showHelp shows the help dialog
func (ui *GameUI) showHelp() {
	// If dialog already exists, update theme label and bring it to front
	if ui.helpDialog != nil {
		if ui.helpCurrentThemeLabel != nil {
			ui.helpCurrentThemeLabel.SetText("Current theme: " + string(currentTheme))
		}
		ui.helpDialog.RequestFocus()
		ui.helpDialog.Show()
		return
	}

	// Create a new window for the dialog
	dialogWindow := ui.app.NewWindow("Help")
	dialogWindow.Resize(fyne.NewSize(550, 700)) // Sized for scrollable content with large sprites
	dialogWindow.SetFixedSize(true)

	// Track the dialog window
	ui.helpDialog = dialogWindow

	// Hide instead of destroying when closed (for faster reopening)
	dialogWindow.SetCloseIntercept(func() {
		dialogWindow.Hide()
	})

	// Create image - TrapperRedPlaid on left side
	iconImage := canvas.NewImageFromResource(resourceKrankyBearTrapperRedPlaidPng)
	iconImage.FillMode = canvas.ImageFillContain
	iconImage.SetMinSize(fyne.NewSize(120, 120))

	// Create help text - use simple label instead of RichText to avoid crashes
	helpText := widget.NewLabel(`How to Play Pacman

Controls:
- Arrow Keys - Move Pacman
- Space - Start Game
- P - Pause/Resume
- R - Restart (after game over)
- F12 - Boss Key (hide window)

Gameplay:
Eat all dots to win!
Avoid ghosts unless you've eaten a power pellet.
Power pellets make ghosts vulnerable for a short time.

Ghosts:
- Blinky (Red): Aggressive chaser
- Pinky (Pink): Ambushes ahead
- Inky (Cyan): Complex behavior
- Clyde (Orange): Chases when far, retreats when close`)
	helpText.Wrapping = fyne.TextWrapWord

	// Create theme section with images
	themeTitle := widget.NewLabel("Themes (use -theme flag or Theme button):")
	themeTitle.TextStyle = fyne.TextStyle{Bold: true}

	// Helper to create themed image row
	createThemeRow := func(name string, pacmanRes, blinkyRes, pinkyRes, inkyRes, clydeRes *fyne.StaticResource) fyne.CanvasObject {
		label := widget.NewLabel(name + ":")
		label.TextStyle = fyne.TextStyle{Bold: true}

		imgSize := fyne.NewSize(64, 64) // Double size for better visibility

		pacmanImg := canvas.NewImageFromResource(pacmanRes)
		pacmanImg.FillMode = canvas.ImageFillContain
		pacmanImg.SetMinSize(imgSize)

		blinkyImg := canvas.NewImageFromResource(blinkyRes)
		blinkyImg.FillMode = canvas.ImageFillContain
		blinkyImg.SetMinSize(imgSize)

		pinkyImg := canvas.NewImageFromResource(pinkyRes)
		pinkyImg.FillMode = canvas.ImageFillContain
		pinkyImg.SetMinSize(imgSize)

		inkyImg := canvas.NewImageFromResource(inkyRes)
		inkyImg.FillMode = canvas.ImageFillContain
		inkyImg.SetMinSize(imgSize)

		clydeImg := canvas.NewImageFromResource(clydeRes)
		clydeImg.FillMode = canvas.ImageFillContain
		clydeImg.SetMinSize(imgSize)

		return container.NewHBox(
			label,
			pacmanImg,
			blinkyImg,
			pinkyImg,
			inkyImg,
			clydeImg,
		)
	}

	// Create basic theme row (with colored circles description)
	basicLabel := widget.NewLabel("basic:")
	basicLabel.TextStyle = fyne.TextStyle{Bold: true}
	basicDesc := widget.NewLabel("Simple colored circles (yellow PacMan, colored ghosts)")
	basicRow := container.NewHBox(basicLabel, basicDesc)

	// Traditional theme row
	traditionalRow := createThemeRow("traditional",
		resourceTraditionalPacManPng,
		resourceTraditionalBlinkyPng,
		resourceTraditionalPinkyPng,
		resourceTraditionalInkyPng,
		resourceTraditionalClydePng,
	)

	// Camo theme row
	camoRow := createThemeRow("camo",
		resourceCamoPacManPng,
		resourceCamoBlinkyPng,
		resourceCamoPinkyPng,
		resourceCamoInkyPng,
		resourceCamoClydePng,
	)

	// Animal theme row
	animalRow := createThemeRow("animal",
		resourceAnimalPacManPng,
		resourceAnimalBlinkyPng,
		resourceAnimalPinkyPng,
		resourceAnimalInkyPng,
		resourceAnimalClydePng,
	)

	// Football theme row
	footballRow := createThemeRow("football",
		resourceFootballPacManPng,
		resourceFootballBlinkyPng,
		resourceFootballPinkyPng,
		resourceFootballInkyPng,
		resourceFootballClydePng,
	)

	// Hogwarts theme row
	hogwartsRow := createThemeRow("hogwarts",
		resourceHogwartsPacManPng,
		resourceHogwartsBlinkyPng,
		resourceHogwartsPinkyPng,
		resourceHogwartsInkyPng,
		resourceHogwartsClydePng,
	)

	// College Football theme row
	collegeFootballRow := createThemeRow("collegefootball",
		resourceCollegeFootballPacManPng,
		resourceCollegeFootballBlinkyPng,
		resourceCollegeFootballPinkyPng,
		resourceCollegeFootballInkyPng,
		resourceCollegeFootballClydePng,
	)

	// Hogwarts2 theme row
	hogwarts2Row := createThemeRow("hogwarts2",
		resourceHogwarts2PacManPng,
		resourceHogwarts2BlinkyPng,
		resourceHogwarts2PinkyPng,
		resourceHogwarts2InkyPng,
		resourceHogwarts2ClydePng,
	)

	// Superheroes theme row
	superheroesRow := createThemeRow("superheroes",
		resourceSuperheroesPacManPng,
		resourceSuperheroesBlinkyPng,
		resourceSuperheroesPinkyPng,
		resourceSuperheroesInkyPng,
		resourceSuperheroesClydePng,
	)

	// Current theme indicator (cached for updates)
	ui.helpCurrentThemeLabel = widget.NewLabel("Current theme: " + string(currentTheme))
	ui.helpCurrentThemeLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Order: basic, traditional, then alphabetical
	themeSection := container.NewVBox(
		widget.NewSeparator(),
		themeTitle,
		basicRow,
		traditionalRow,
		animalRow,
		camoRow,
		collegeFootballRow,
		footballRow,
		hogwartsRow,
		hogwarts2Row,
		superheroesRow,
		ui.helpCurrentThemeLabel,
	)

	// Combine help text and theme section
	helpContent := container.NewVBox(
		helpText,
		themeSection,
	)

	// Wrap text in scroll container
	textScroll := container.NewVScroll(helpContent)
	textScroll.SetMinSize(fyne.NewSize(380, 550))

	// Layout: image on left, text on right
	textContainer := container.NewStack(
		container.NewPadded(textScroll),
	)
	content := container.NewBorder(
		nil,
		container.NewCenter(widget.NewButton("Close", func() {
			dialogWindow.Hide()
		})),
		nil,
		nil,
		container.NewHBox(
			container.NewPadded(iconImage),
			textContainer,
		),
	)

	dialogWindow.SetContent(content)
	dialogWindow.Show()

	// Position dialog relative to main window (after showing)
	ui.positionDialogRelativeToMain(dialogWindow)
}

// showAbout shows the about dialog
func (ui *GameUI) showAbout() {
	// If dialog already exists, bring it to front
	if ui.aboutDialog != nil {
		ui.aboutDialog.RequestFocus()
		ui.aboutDialog.Show()
		return
	}

	// Create a new window for the dialog
	dialogWindow := ui.app.NewWindow("About")
	dialogWindow.Resize(fyne.NewSize(500, 300))
	dialogWindow.SetFixedSize(true)

	// Track the dialog window
	ui.aboutDialog = dialogWindow

	// Reset tracking when window closes
	dialogWindow.SetOnClosed(func() {
		ui.aboutDialog = nil
	})

	// Create image - TrapperRedPlaid on left side
	iconImage := canvas.NewImageFromResource(resourceKrankyBearTrapperRedPlaidPng)
	iconImage.FillMode = canvas.ImageFillContain
	iconImage.SetMinSize(fyne.NewSize(150, 150))

	// Create GitHub URL
	githubLink, err := url.Parse("https://github.com/amarillier/KrankyBearPacMan")
	if err != nil {
		fyne.LogError("Could not parse URL", err)
	}
	githubHyperlink := widget.NewHyperlink("https://github.com/amarillier/KrankyBearPacMan", githubLink)
	githubHyperlink.Alignment = fyne.TextAlignLeading

	// Create text content
	aboutText := widget.NewRichTextFromMarkdown(`# ` + appName + `

**Version:** ` + appVersion + `

**Author:** ` + appAuthor + `

**Copyright:** ` + appCopyright + `

A classic Pacman game with the four original ghosts:
Blinky, Pinky, Inky, and Clyde.

Enjoy the game!`)

	aboutText.Wrapping = fyne.TextWrapWord

	// Layout: image on left, text on right
	textContainer := container.NewVBox(
		container.NewPadded(aboutText),
		githubHyperlink,
	)
	content := container.NewBorder(
		nil,
		container.NewCenter(widget.NewButton("Close", func() {
			dialogWindow.Close()
		})),
		nil,
		nil,
		container.NewHBox(
			container.NewPadded(iconImage),
			textContainer,
		),
	)

	dialogWindow.SetContent(content)
	dialogWindow.Show()

	// Position dialog relative to main window (after showing)
	ui.positionDialogRelativeToMain(dialogWindow)
}

// updateChecker checks for version updates
func (ui *GameUI) updateChecker() (string, bool) {
	uc := updatechecker.New("amarillier", "KrankyBearPacMan", appName, "https://github.com/amarillier/KrankyBearPacMan/releases/latest", 0, false)
	uc.CheckForUpdate(appVersion)
	return uc.Message, uc.UpdateAvailable
}

// checkForUpdates checks for version updates
func (ui *GameUI) checkForUpdates() {
	updtmsg, updateAvailable := ui.updateChecker()
	// Check if we're running a newer version than released
	if strings.Contains(updtmsg, "running a newer version") || strings.Contains(updtmsg, "newer than") {
		ui.versionStatus = "newer" // We're running newer than released
	} else if strings.Contains(updtmsg, "running the latest") || strings.Contains(updtmsg, "up to date") {
		ui.versionStatus = "current" // Version matches released
	} else if updateAvailable {
		ui.versionStatus = "update" // Update available
	} else {
		ui.versionStatus = "unknown"
	}
}

// showUpdateDialog shows the Update Check dialog with version status and appropriate image
func (ui *GameUI) showUpdateDialog() {
	// Check if window already exists
	if ui.updateDialog != nil {
		ui.updateDialog.RequestFocus()
		return
	}

	// Run update check in goroutine to avoid blocking UI
	go func() {
		updtmsg, _ := ui.updateChecker()
		fyne.Do(func() {
			ui.updateAlert(updtmsg)
		})
	}()
}

// updateAlert displays the update check result in a dialog window
func (ui *GameUI) updateAlert(updtmsg string) {
	// Check if window already exists
	if ui.updateDialog != nil {
		ui.updateDialog.RequestFocus()
		return
	}

	// Create release link
	releaselink, rerr := url.Parse("https://github.com/amarillier/KrankyBearPacMan/releases/latest")
	if rerr != nil {
		fyne.LogError("Could not parse URL", rerr)
	}
	myreleaselink := widget.NewHyperlink("https://github.com/amarillier/KrankyBearPacMan/releases/latest", releaselink)
	myreleaselink.Alignment = fyne.TextAlignLeading

	// Create release notes link
	releasenoteslink, rnerr := url.Parse("https://github.com/amarillier/KrankyBearPacMan/blob/allanm/ReleaseNotes.txt")
	if rnerr != nil {
		fyne.LogError("Could not parse URL", rnerr)
	}
	myreleasenoteslink := widget.NewHyperlink("https://github.com/amarillier/KrankyBearPacMan/blob/allanm/ReleaseNotes.txt", releasenoteslink)
	myreleasenoteslink.Alignment = fyne.TextAlignLeading

	// Create image based on update message
	// Hard Hat (yellow): running newer version than released (developer/beta)
	// TrapperRedPlaid: current version or update available (normal user)
	var kbimg *canvas.Image
	if strings.Contains(updtmsg, "newer version") {
		kbimg = canvas.NewImageFromResource(resourceKrankyBearHardHatPng)
		kbimg.FillMode = canvas.ImageFillOriginal
	} else {
		// For current version, update available, or unknown status, show TrapperRedPlaid
		kbimg = canvas.NewImageFromResource(resourceKrankyBearTrapperRedPlaidPng)
		kbimg.FillMode = canvas.ImageFillOriginal
	}

	// Create version label
	versionLabel := widget.NewLabel("Current Version: " + appVersion)
	versionLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Create text label with update message
	text := widget.NewLabel(updtmsg)
	text.Wrapping = fyne.TextWrapWord

	// Create content: image, version, update message, release link, release notes link
	content := container.NewVBox(kbimg, versionLabel, text, myreleaselink, myreleasenoteslink)

	// Create window
	ui.updateDialog = ui.app.NewWindow(appName + ": Update Check")
	ui.updateDialog.SetIcon(resourceKrankyBearTrapperRedPlaidPng)
	ui.updateDialog.Resize(fyne.NewSize(600, 320))
	ui.updateDialog.SetContent(content)
	ui.updateDialog.SetCloseIntercept(func() {
		ui.updateDialog.Close()
		ui.updateDialog = nil
	})
	ui.updateDialog.Show()

	ui.positionDialogRelativeToMain(ui.updateDialog)
}
