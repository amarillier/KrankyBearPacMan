package main

import (
	"math"
)

// Player represents Pacman
type Player struct {
	position     Position
	nextPosition Position
	direction    Direction
	nextDirection Direction
	speed        float64
	animFrame    int
}

// NewPlayer creates a new player
func NewPlayer(pos Position) *Player {
	return &Player{
		position:     pos,
		nextPosition: pos,
		direction:    None,
		nextDirection: None,
		speed:        0.15, // Fraction of tile per frame
		animFrame:    0,
	}
}

// Update updates the player position
func (p *Player) Update(g *Game) {
	// Try to change direction if requested
	if p.nextDirection != None && p.canChangeDirection(g, p.nextDirection) {
		p.direction = p.nextDirection
		p.nextDirection = None
	}

	// Move in current direction
	if p.direction != None {
		dx, dy := p.getDirectionVector(p.direction)
		newX := p.position.X + float64(dx)*p.speed
		newY := p.position.Y + float64(dy)*p.speed

		// Handle tunnel wrapping BEFORE movement check (only in tunnel row, around Y=14)
		tunnelRow := 14
		currentY := int(math.Round(newY))
		if currentY == tunnelRow {
			// Allow wrapping at tunnel edges - check if we're moving into tunnel area
			if newX < -0.5 {
				p.position.X = float64(MazeWidth - 1) + (newX + 0.5)
				return
			}
			if newX >= float64(MazeWidth)-0.5 {
				p.position.X = newX - float64(MazeWidth) + 0.5
				return
			}
		}

		// Check if we can move to the next tile
		nextTileX := int(math.Round(p.position.X)) + dx
		nextTileY := int(math.Round(p.position.Y)) + dy

		// For tunnel row, allow movement even if slightly out of bounds
		if currentY == tunnelRow && (nextTileX < 0 || nextTileX >= MazeWidth) {
			// Allow movement in tunnel
			p.position.X = newX
			p.position.Y = newY
		} else if g.IsValidMove(nextTileX, nextTileY) {
			p.position.X = newX
			p.position.Y = newY

			// Snap to grid when reaching center of tile
			gridX := math.Round(p.position.X)
			gridY := math.Round(p.position.Y)
			if math.Abs(p.position.X-gridX) < 0.1 && math.Abs(p.position.Y-gridY) < 0.1 {
				p.position.X = gridX
				p.position.Y = gridY
			}
		} else {
			// Stop if hitting a wall
			p.position.X = math.Round(p.position.X)
			p.position.Y = math.Round(p.position.Y)
		}

		// Final tunnel wrapping check (in case we're already past the edge)
		if currentY == tunnelRow {
			if p.position.X < 0 {
				p.position.X = float64(MazeWidth - 1)
			}
			if p.position.X >= float64(MazeWidth) {
				p.position.X = 0
			}
		}
	}

	p.animFrame++
}

// SetDirection sets the desired direction
func (p *Player) SetDirection(dir Direction) {
	p.nextDirection = dir
}

// GetGridPosition returns the grid position (rounded)
func (p *Player) GetGridPosition() Position {
	return Position{
		X: math.Round(p.position.X),
		Y: math.Round(p.position.Y),
	}
}

// GetPosition returns the exact position
func (p *Player) GetPosition() Position {
	return p.position
}

// canChangeDirection checks if the player can change direction
func (p *Player) canChangeDirection(g *Game, dir Direction) bool {
	gridPos := p.GetGridPosition()
	dx, dy := p.getDirectionVector(dir)
	nextX := int(gridPos.X) + dx
	nextY := int(gridPos.Y) + dy
	return g.IsValidMove(nextX, nextY)
}

// getDirectionVector returns the direction vector
func (p *Player) getDirectionVector(dir Direction) (int, int) {
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

// Reset resets the player to starting position
func (p *Player) Reset() {
	p.position = Position{X: 14.0, Y: 23.0}
	p.nextPosition = p.position
	p.direction = None
	p.nextDirection = None
}

