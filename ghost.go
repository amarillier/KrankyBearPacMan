package main

import (
	"math"
	"time"
)

// GhostColor represents ghost colors
type GhostColor int

const (
	Red GhostColor = iota
	Pink
	Cyan
	Orange
)

// GhostAI is a function type for ghost AI behavior
type GhostAI func(*Ghost, *Game) Position

// Ghost represents a ghost enemy
type Ghost struct {
	name          string
	color         GhostColor
	position      Position
	nextPosition  Position
	direction     Direction
	mode          GhostMode
	ai            GhostAI
	eatenCount    int
	speed         float64
	inHouse       bool
	hasExitedHouse bool  // True once ghost has left house area
	houseTimer    time.Time
}

// NewGhost creates a new ghost
func NewGhost(name string, pos Position, color GhostColor, ai GhostAI) *Ghost {
	now := time.Now()
	inHouse := true
	var houseTimer time.Time
	
	// All ghosts except Blinky start in house and release after 2s
	switch name {
	case "Blinky":
		inHouse = false // Already outside house
		houseTimer = now
	default:
		inHouse = true
		houseTimer = now // All release after 2s
	}
	
	return &Ghost{
		name:           name,
		color:          color,
		position:       pos,
		nextPosition:   pos,
		direction:      Up,
		mode:           ModeScatter,
		ai:             ai,
		speed:          0.15,
		inHouse:        inHouse,
		hasExitedHouse: name == "Blinky", // Blinky starts outside
		houseTimer:     houseTimer,
	}
}

// Update updates the ghost
func (g *Ghost) Update(game *Game) {
	// Handle house exit - ghosts wait in house for staggered release
	if g.inHouse {
		if time.Since(g.houseTimer) > 2*time.Second {
			g.inHouse = false
		} else {
			return // Still in house, don't move yet
		}
	}

	// Update mode based on game state
	if !game.frightenedEnd.IsZero() && time.Now().Before(game.frightenedEnd) {
		g.mode = ModeFrightened
	} else {
		g.mode = game.currentMode
	}

	// Get current grid position
	gridX := int(math.Round(g.position.X))
	gridY := int(math.Round(g.position.Y))
	
	// Check if in house area
	inHouseArea := gridY >= 12 && gridY <= 15 && gridX >= 11 && gridX <= 17
	
	// Mark as exited once outside house area
	if !inHouseArea && !g.hasExitedHouse {
		g.hasExitedHouse = true
	}
	
	// Snap to grid when close
	if math.Abs(g.position.X-float64(gridX)) < 0.1 {
		g.position.X = float64(gridX)
	}
	if math.Abs(g.position.Y-float64(gridY)) < 0.1 {
		g.position.Y = float64(gridY)
	}
	
	// At grid center?
	atCenter := g.position.X == float64(gridX) && g.position.Y == float64(gridY)
	
	// Get current movement vector
	dx, dy := g.getDirectionVector(g.direction)
	
	// Check if blocked
	blocked := g.direction == None || !game.IsValidMove(gridX+dx, gridY+dy)
	
	// Count available exits (for intersection detection)
	numExits := 0
	for _, dir := range []Direction{Up, Down, Left, Right} {
		tdx, tdy := g.getDirectionVector(dir)
		if game.IsValidMove(gridX+tdx, gridY+tdy) {
			numExits++
		}
	}
	
	// An intersection has 3+ exits (can turn, not just corridor)
	isIntersection := numExits >= 3
	
	// Choose new direction when:
	// 1. Blocked (must turn)
	// 2. At an intersection (can choose to turn)
	needsNewDirection := blocked || (atCenter && isIntersection)
	
	if atCenter && needsNewDirection {
		if inHouseArea && !g.hasExitedHouse {
			// Still exiting house - navigate toward exit
			if gridX < 14 {
				g.direction = Right
			} else if gridX > 14 {
				g.direction = Left
			} else {
				g.direction = Up
			}
		} else {
			// Normal maze navigation - use AI
			target := g.ai(g, game)
			g.direction = g.findBestDirection(game, target, gridX, gridY)
		}
		
		// Verify direction is valid, find fallback if not
		dx, dy = g.getDirectionVector(g.direction)
		if g.direction == None || !game.IsValidMove(gridX+dx, gridY+dy) {
			// Try all directions
			for _, dir := range []Direction{Up, Left, Right, Down} {
				tdx, tdy := g.getDirectionVector(dir)
				if game.IsValidMove(gridX+tdx, gridY+tdy) {
					g.direction = dir
					dx, dy = tdx, tdy
					break
				}
			}
		}
	}
	
	// Move in current direction
	if g.direction != None {
		dx, dy = g.getDirectionVector(g.direction)
		newX := g.position.X + float64(dx)*g.speed
		newY := g.position.Y + float64(dy)*g.speed
		
		// Handle tunnel wrapping (row 14, edges of maze)
		if gridY == 14 {
			if newX < 0 {
				newX = float64(MazeWidth - 1)
			} else if newX >= float64(MazeWidth) {
				newX = 0
			}
		}
		
		// Calculate next tile
		nextTileX := int(math.Round(newX))
		nextTileY := int(math.Round(newY))
		
		// Move if valid
		if game.IsValidMove(nextTileX, nextTileY) {
			g.position.X = newX
			g.position.Y = newY
		}
	}
}

// findBestDirection finds the best direction toward target without reversing
func (g *Ghost) findBestDirection(game *Game, target Position, gridX, gridY int) Direction {
	bestDir := g.direction
	bestDist := math.MaxFloat64
	
	for _, dir := range []Direction{Up, Down, Left, Right} {
		// Can't reverse (except in frightened mode)
		if g.mode != ModeFrightened && isOpposite(dir, g.direction) {
			continue
		}
		
		dx, dy := g.getDirectionVector(dir)
		nextX := gridX + dx
		nextY := gridY + dy
		
		if game.IsValidMove(nextX, nextY) {
			dist := Distance(Position{X: float64(nextX), Y: float64(nextY)}, target)
			if dist < bestDist {
				bestDist = dist
				bestDir = dir
			}
		}
	}
	
	// If no valid direction found (all blocked except reverse), allow reverse
	if bestDir == g.direction || bestDir == None {
		for _, dir := range []Direction{Up, Down, Left, Right} {
			dx, dy := g.getDirectionVector(dir)
			if game.IsValidMove(gridX+dx, gridY+dy) {
				return dir
			}
		}
	}
	
	return bestDir
}

// findAnyValidDirection finds any valid direction
func (g *Ghost) findAnyValidDirection(game *Game, gridX, gridY int) Direction {
	for _, dir := range []Direction{Up, Down, Left, Right} {
		dx, dy := g.getDirectionVector(dir)
		if game.IsValidMove(gridX+dx, gridY+dy) {
			return dir
		}
	}
	return None
}


// isOpposite checks if two directions are opposite
func isOpposite(d1, d2 Direction) bool {
	return (d1 == Up && d2 == Down) ||
		(d1 == Down && d2 == Up) ||
		(d1 == Left && d2 == Right) ||
		(d1 == Right && d2 == Left)
}

// getDirectionVector returns the direction vector
func (g *Ghost) getDirectionVector(dir Direction) (int, int) {
	switch dir {
	case Up:
		return 0, -1
	case Down:
		return 0, 1
	case Left:
		return -1, 0
	case Right:
		return 1, 0
	default:
		return 0, 0
	}
}

// GetGridPosition returns the grid position
func (g *Ghost) GetGridPosition() Position {
	return Position{
		X: math.Round(g.position.X),
		Y: math.Round(g.position.Y),
	}
}

// Reset resets the ghost to starting position
func (g *Ghost) Reset(game *Game) {
	now := time.Now()
	switch g.name {
	case "Blinky":
		g.position = Position{X: 14.0, Y: 11.0}
		g.inHouse = false
		g.hasExitedHouse = true // Already outside
		g.houseTimer = now
	case "Pinky":
		g.position = Position{X: 14.0, Y: 14.0}
		g.inHouse = true
		g.hasExitedHouse = false
		g.houseTimer = now // Releases after 2s
	case "Inky":
		g.position = Position{X: 12.0, Y: 14.0}
		g.inHouse = true
		g.hasExitedHouse = false
		g.houseTimer = now // Releases after 2s (same as Pinky for simplicity)
	case "Clyde":
		g.position = Position{X: 16.0, Y: 14.0}
		g.inHouse = true
		g.hasExitedHouse = false
		g.houseTimer = now // Releases after 2s
	default:
		g.position = Position{X: 14.0, Y: 14.0}
		g.inHouse = true
		g.hasExitedHouse = false
		g.houseTimer = now
	}
	g.direction = Up
	g.mode = game.currentMode
	g.eatenCount = 0
}

// ============================================================================
// Ghost AI Behaviors - Classic Pacman AI
// ============================================================================

// BlinkyAI - Red ghost: Aggressive chaser, targets player directly
func BlinkyAI(ghost *Ghost, game *Game) Position {
	if ghost.mode == ModeFrightened {
		return getRandomTarget(game)
	}
	if ghost.mode == ModeScatter {
		return Position{X: float64(MazeWidth - 2), Y: 0} // Top right corner
	}
	// Chase: Target player directly
	return game.player.GetGridPosition()
}

// PinkyAI - Pink ghost: Ambushes by targeting 4 tiles ahead of player
func PinkyAI(ghost *Ghost, game *Game) Position {
	if ghost.mode == ModeFrightened {
		return getRandomTarget(game)
	}
	if ghost.mode == ModeScatter {
		return Position{X: 0, Y: 0} // Top left corner
	}
	// Chase: Target 4 tiles ahead of player
	playerPos := game.player.GetGridPosition()
	dx, dy := game.player.getDirectionVector(game.player.direction)
	target := Position{
		X: playerPos.X + float64(dx*4),
		Y: playerPos.Y + float64(dy*4),
	}
	return target
}

// InkyAI - Cyan ghost: Complex behavior using Blinky's position
func InkyAI(ghost *Ghost, game *Game) Position {
	if ghost.mode == ModeFrightened {
		return getRandomTarget(game)
	}
	if ghost.mode == ModeScatter {
		return Position{X: float64(MazeWidth - 2), Y: float64(MazeHeight - 1)} // Bottom right corner
	}
	// Chase: Complex behavior
	playerPos := game.player.GetGridPosition()
	dx, dy := game.player.getDirectionVector(game.player.direction)
	// Target 2 tiles ahead of player
	intermediate := Position{
		X: playerPos.X + float64(dx*2),
		Y: playerPos.Y + float64(dy*2),
	}
	// Find Blinky
	var blinkyPos Position
	for _, g := range game.ghosts {
		if g.name == "Blinky" {
			blinkyPos = g.GetGridPosition()
			break
		}
	}
	// Target is 2x the vector from Blinky to intermediate
	target := Position{
		X: intermediate.X + (intermediate.X - blinkyPos.X),
		Y: intermediate.Y + (intermediate.Y - blinkyPos.Y),
	}
	return target
}

// ClydeAI - Orange ghost: Chases when far, retreats when close
func ClydeAI(ghost *Ghost, game *Game) Position {
	if ghost.mode == ModeFrightened {
		return getRandomTarget(game)
	}
	if ghost.mode == ModeScatter {
		return Position{X: 0, Y: float64(MazeHeight - 1)} // Bottom left corner
	}
	// Chase: If far from player, chase. If close, scatter to corner
	playerPos := game.player.GetGridPosition()
	ghostPos := ghost.GetGridPosition()
	dist := Distance(ghostPos, playerPos)

	if dist > 8 {
		// Far: chase player
		return playerPos
	} else {
		// Close: scatter to bottom left
		return Position{X: 0, Y: float64(MazeHeight - 1)}
	}
}

// getRandomTarget returns a random valid position (for frightened mode)
func getRandomTarget(game *Game) Position {
	for i := 0; i < 100; i++ {
		x := game.rand.Intn(MazeWidth)
		y := game.rand.Intn(MazeHeight)
		if game.IsValidMove(x, y) {
			return Position{X: float64(x), Y: float64(y)}
		}
	}
	return Position{X: float64(MazeWidth / 2), Y: float64(MazeHeight / 2)}
}

