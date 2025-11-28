package main

import (
	"math"
	"math/rand"
	"time"
)

// Game constants
const (
	TileSize       = 20
	MazeWidth      = 28
	MazeHeight     = 31
	WindowWidth    = MazeWidth * TileSize
	WindowHeight   = MazeHeight * TileSize + 60 // Extra space for score
	FrightenedTime = 6 * time.Second
	ScatterTime    = 7 * time.Second
	ChaseTime      = 20 * time.Second
)

// Cell types in the maze
type CellType int

const (
	Empty CellType = iota
	Wall
	Dot
	PowerPellet
	GhostHouse
)

// Direction
type Direction int

const (
	None Direction = iota
	Up
	Down
	Left
	Right
)

// Game state
type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StatePaused
	StateGameOver
	StateWin
)

// Position on the maze grid (can be float for smooth movement)
type Position struct {
	X, Y float64
}

// GridPosition returns integer grid coordinates
func (p Position) GridX() int {
	return int(p.X)
}

func (p Position) GridY() int {
	return int(p.Y)
}

// Classic Pacman maze layout (simplified version)
var classicMaze = [MazeHeight][MazeWidth]CellType{
	{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	{1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 1},
	{1, 2, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 2, 1, 1, 2, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 2, 1},
	{1, 3, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 2, 1, 1, 2, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 3, 1},
	{1, 2, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 2, 1, 1, 2, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 2, 1},
	{1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 1},
	{1, 2, 1, 1, 1, 1, 2, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 1, 2, 1, 1, 2, 1, 1, 1, 1, 2, 1},
	{1, 2, 1, 1, 1, 1, 2, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 1, 2, 1, 1, 2, 1, 1, 1, 1, 2, 1},
	{1, 2, 2, 2, 2, 2, 2, 1, 1, 2, 2, 2, 2, 1, 1, 2, 2, 2, 2, 1, 1, 2, 2, 2, 2, 2, 2, 1},
	{1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1},
	{0, 0, 0, 0, 0, 1, 2, 1, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 1, 1, 1, 2, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 2, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 2, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 2, 1, 1, 0, 1, 1, 1, 4, 4, 1, 1, 1, 0, 1, 1, 2, 1, 0, 0, 0, 0, 0},
	{1, 1, 1, 1, 1, 1, 2, 1, 1, 0, 1, 4, 4, 4, 4, 4, 4, 1, 0, 1, 1, 2, 1, 1, 1, 1, 1, 1},
	{0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 1, 4, 4, 4, 4, 4, 4, 1, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0},
	{1, 1, 1, 1, 1, 1, 2, 1, 1, 0, 1, 4, 4, 4, 4, 4, 4, 1, 0, 1, 1, 2, 1, 1, 1, 1, 1, 1},
	{0, 0, 0, 0, 0, 1, 2, 1, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1, 0, 1, 1, 2, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 2, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 2, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 1, 2, 1, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1, 0, 1, 1, 2, 1, 0, 0, 0, 0, 0},
	{1, 1, 1, 1, 1, 1, 2, 1, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1, 0, 1, 1, 2, 1, 1, 1, 1, 1, 1},
	{1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 1},
	{1, 2, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 2, 1, 1, 2, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 2, 1},
	{1, 2, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 2, 1, 1, 2, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 2, 1},
	{1, 3, 2, 2, 1, 1, 2, 2, 2, 2, 2, 2, 2, 0, 0, 2, 2, 2, 2, 2, 2, 2, 1, 1, 2, 2, 3, 1},
	{1, 1, 1, 2, 1, 1, 2, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 1, 2, 1, 1, 2, 1, 1, 2, 1, 1, 1},
	{1, 1, 1, 2, 1, 1, 2, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 1, 2, 1, 1, 2, 1, 1, 2, 1, 1, 1},
	{1, 2, 2, 2, 2, 2, 2, 1, 1, 2, 2, 2, 2, 1, 1, 2, 2, 2, 2, 1, 1, 2, 2, 2, 2, 2, 2, 1},
	{1, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 1},
	{1, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 1},
	{1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 1},
	{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
}

// Game struct
type Game struct {
	maze          [MazeHeight][MazeWidth]CellType
	player        *Player
	ghosts        []*Ghost
	score         int
	lives         int
	level         int
	state         GameState
	dotsRemaining int
	frightenedEnd time.Time
	modeStartTime time.Time
	currentMode   GhostMode
	rand          *rand.Rand
	soundCallback func(SoundEvent)
}

// GhostMode represents the current AI mode
type GhostMode int

const (
	ModeScatter GhostMode = iota
	ModeChase
	ModeFrightened
)

// NewGame creates a new game instance
func NewGame() *Game {
	g := &Game{
		maze:          classicMaze,
		score:         0,
		lives:         3,
		level:         1,
		state:         StateMenu,
		frightenedEnd: time.Time{},
		modeStartTime: time.Now(),
		currentMode:   ModeScatter,
		rand:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	// Count dots
	g.countDots()

	// Initialize player
	g.player = NewPlayer(Position{X: 14.0, Y: 23.0})

	// Initialize ghosts with their classic starting positions
	g.ghosts = []*Ghost{
		NewGhost("Blinky", Position{X: 14.0, Y: 11.0}, Red, BlinkyAI),
		NewGhost("Pinky", Position{X: 14.0, Y: 14.0}, Pink, PinkyAI),
		NewGhost("Inky", Position{X: 12.0, Y: 14.0}, Cyan, InkyAI),
		NewGhost("Clyde", Position{X: 16.0, Y: 14.0}, Orange, ClydeAI),
	}

	return g
}

// SetSoundCallback sets the callback function for sound events
func (g *Game) SetSoundCallback(callback func(SoundEvent)) {
	g.soundCallback = callback
}

// Count dots in the maze
func (g *Game) countDots() {
	g.dotsRemaining = 0
	for y := 0; y < MazeHeight; y++ {
		for x := 0; x < MazeWidth; x++ {
			if g.maze[y][x] == Dot || g.maze[y][x] == PowerPellet {
				g.dotsRemaining++
			}
		}
	}
}

// Start begins the game
func (g *Game) Start() {
	g.state = StatePlaying
	g.modeStartTime = time.Now()
	g.currentMode = ModeScatter
}

// resetLevel resets the maze and entities for the next level
func (g *Game) resetLevel() {
	// Reset maze to original state
	g.maze = classicMaze
	
	// Count dots
	g.countDots()
	
	// Reset player position
	g.player = NewPlayer(Position{X: 14.0, Y: 23.0})
	
	// Reset ghosts to starting positions
	for _, ghost := range g.ghosts {
		ghost.Reset(g)
	}
	
	// Reset game mode timing
	g.modeStartTime = time.Now()
	g.currentMode = ModeScatter
	g.frightenedEnd = time.Time{}
}

// Update updates the game state
func (g *Game) Update() {
	if g.state != StatePlaying {
		return
	}

	// Update ghost AI mode
	g.updateGhostMode()

	// Update player
	g.player.Update(g)

	// Update ghosts
	for _, ghost := range g.ghosts {
		ghost.Update(g)
	}

	// Check collisions
	g.checkCollisions()

	// Check level completion - advance to next level instead of ending
	if g.dotsRemaining == 0 {
		g.level++
		g.resetLevel()
	}
}

// Update ghost AI mode (scatter/chase cycle)
func (g *Game) updateGhostMode() {
	now := time.Now()
	elapsed := now.Sub(g.modeStartTime)

	// If frightened mode is active, don't change scatter/chase
	if !g.frightenedEnd.IsZero() && now.Before(g.frightenedEnd) {
		return
	}

	// Clear frightened mode if it expired
	if !g.frightenedEnd.IsZero() && now.After(g.frightenedEnd) {
		g.frightenedEnd = time.Time{}
		for _, ghost := range g.ghosts {
			if ghost.mode == ModeFrightened {
				ghost.mode = ModeChase
			}
		}
	}

	// Classic Pacman mode cycle:
	// Scatter 7s, Chase 20s, Scatter 7s, Chase 20s, Scatter 5s, Chase 20s, Scatter 5s, Chase indefinite
	var nextMode GhostMode

	if g.currentMode == ModeScatter {
		nextMode = ModeChase
		if elapsed >= ScatterTime {
			g.currentMode = nextMode
			g.modeStartTime = now
		}
	} else {
		nextMode = ModeScatter
		if elapsed >= ChaseTime {
			g.currentMode = nextMode
			g.modeStartTime = now
		}
	}
}

// Check collisions between player and ghosts/dots
func (g *Game) checkCollisions() {
	playerPos := g.player.GetPosition()
	// Get the grid cell the player is currently in
	px := int(math.Floor(playerPos.X + 0.5)) // Round to nearest integer
	py := int(math.Floor(playerPos.Y + 0.5))

	// Check the cell the player is in for dots/pellets
	if py >= 0 && py < MazeHeight && px >= 0 && px < MazeWidth {
		if g.maze[py][px] == Dot {
			g.maze[py][px] = Empty
			g.score += 10
			g.dotsRemaining--
			if g.soundCallback != nil {
				g.soundCallback(SoundEventDot)
			}
		} else if g.maze[py][px] == PowerPellet {
			g.maze[py][px] = Empty
			g.score += 50
			g.dotsRemaining--
			// Activate frightened mode
			g.frightenedEnd = time.Now().Add(FrightenedTime)
			for _, ghost := range g.ghosts {
				if ghost.mode != ModeFrightened {
					ghost.mode = ModeFrightened
				}
			}
			if g.soundCallback != nil {
				g.soundCallback(SoundEventPowerPellet)
			}
		}
	}

	// Check ghost collisions
	for _, ghost := range g.ghosts {
		ghostPos := ghost.GetGridPosition()
		gx, gy := ghostPos.GridX(), ghostPos.GridY()
		if px == gx && py == gy {
			if ghost.mode == ModeFrightened {
				// Player eats ghost
				g.score += 200 * (1 << ghost.eatenCount) // 200, 400, 800, 1600
				ghost.eatenCount++
				ghost.Reset(g)
				if g.soundCallback != nil {
					g.soundCallback(SoundEventEatGhost)
				}
			} else {
				// Ghost catches player
				g.lives--
				if g.lives <= 0 {
					g.state = StateGameOver
					if g.soundCallback != nil {
						g.soundCallback(SoundEventGameOver)
					}
				} else {
					// Reset positions
					g.player.Reset()
					for _, ghost := range g.ghosts {
						ghost.Reset(g)
					}
				}
			}
		}
	}
}

// GetCell returns the cell type at the given position
func (g *Game) GetCell(x, y int) CellType {
	if x < 0 || x >= MazeWidth || y < 0 || y >= MazeHeight {
		return Wall
	}
	return g.maze[y][x]
}

// GetCellFloat returns the cell type at the given float position
func (g *Game) GetCellFloat(x, y float64) CellType {
	return g.GetCell(int(x), int(y))
}

// IsValidMove checks if a position is valid (not a wall)
func (g *Game) IsValidMove(x, y int) bool {
	cell := g.GetCell(x, y)
	return cell != Wall
}

// Distance calculates Manhattan distance between two positions
func Distance(p1, p2 Position) float64 {
	return math.Abs(p1.X-p2.X) + math.Abs(p1.Y-p2.Y)
}

// WrapPosition handles tunnel wrapping (left/right edges)
func WrapPosition(pos Position) Position {
	if pos.X < 0 {
		pos.X = float64(MazeWidth - 1)
	}
	if pos.X >= float64(MazeWidth) {
		pos.X = 0
	}
	return pos
}

