# KrankyBear PacMan

A cross-platform PacMan game built with Go and the Fyne GUI library. Features classic arcade gameplay with the four original ghosts: Blinky, Pinky, Inky, and Clyde.

## Features

- **Classic Gameplay**: Navigate the maze, eat dots, avoid ghosts, and use power pellets to turn the tables
- **Four Original Ghosts**: Each with their own AI behavior
  - **Blinky** (Red): Aggressive chaser, targets player directly
  - **Pinky** (Pink): Ambushes by targeting 4 tiles ahead
  - **Inky** (Cyan): Complex behavior using Blinky's position
  - **Clyde** (Orange): Chases when far, retreats when close
- **Multiple Themes**: Choose your character sprites
- **Sound Effects**: Toggle on/off during gameplay
- **High Scores**: Local leaderboard with top 5 scores
- **Boss Key** (F12): Quickly hide the window when needed
- **System Tray**: Minimize to system tray
- **Cross-Platform**: Works on macOS, Windows, and Linux

## Themes

KrankyBear PacMan supports multiple sprite themes for PacMan and the ghosts. Switch themes anytime using the in-game Theme button or via command line.

| Theme | Description |
|-------|-------------|
| `basic` | Simple colored circles (default) |
| `traditional` | Classic PacMan and ghost artwork |
| `animal` | Wild animal characters (Bear, Lion, Baboon, Moose) |
| `camo` | Camouflage-themed characters |
| `collegefootball` | College football mascots (Ohio State, Michigan, Texas, Notre Dame) |
| `football` | NFL team mascots (Cowboys, Patriots, Eagles, Raiders) |
| `hogwarts` | Harry Potter house themes (Gryffindor, Hufflepuff, Ravenclaw, Slytherin) |
| `hogwarts2` | Harry Potter house themes (alternate artwork) |
| `superheroes` | Superhero characters |

### Command Line Usage

```bash
# Default (basic) theme
./pacman

# Traditional theme
./pacman -theme traditional

# Animal theme
./pacman -theme animal

# Camo theme  
./pacman -theme camo

# College Football theme
./pacman -theme collegefootball

# Football theme
./pacman -theme football

# Hogwarts theme
./pacman -theme hogwarts

# Hogwarts2 theme (alternate)
./pacman -theme hogwarts2

# Superheroes theme
./pacman -theme superheroes

# Show help
./pacman -help
```

### Adding New Themes

The game is designed to support additional themes. This will of course require recompiling the game. To add a new theme:

1. Create sprite images (PNG format, transparent background recommended)
   - `PacMan.png` - The player character
   - `Blinky.png` - Red ghost
   - `Pinky.png` - Pink ghost
   - `Inky.png` - Cyan ghost
   - `Clyde.png` - Orange ghost

2. Place them in `Resources/Images/<theme-name>/`

3. Add the embedded resources to `bundled.go`

4. Update the theme constants and switch statements in `main.go` and `ui.go`

**Potential future themes**: Animals, Pokemon, Football teams, Holidays, etc.

## Controls

| Key | Action |
|-----|--------|
| Arrow Keys | Move PacMan |
| Space | Start game |
| P | Pause/Resume |
| R | Restart (after game over) |
| Z | Buy extra life (-500 points) |
| F12 | Boss key (hide & pause) |

## Building

### Prerequisites

- Go 1.21 or later
- Fyne dependencies (see [Fyne Getting Started](https://developer.fyne.io/started/))

### Build Commands

```bash
# Build for current platform
go build -o pacman .

# Run directly
go run .

# Build with specific theme
go build -o pacman . && ./pacman -theme traditional
```

### Cross-Platform Builds

```bash
# macOS (ARM64)
GOOS=darwin GOARCH=arm64 go build -o bin/pacman-macos-arm64 .

# macOS (AMD64)
GOOS=darwin GOARCH=amd64 go build -o bin/pacman-macos-amd64 .

# Windows
GOOS=windows GOARCH=amd64 go build -o bin/pacman-windows.exe .

# Linux
GOOS=linux GOARCH=amd64 go build -o bin/pacman-linux-amd64 .
```

## Dependencies

- [Fyne](https://fyne.io/) v2.6+ - Cross-platform GUI toolkit
- [beep](https://github.com/gopxl/beep) - Audio playback
- [resize](https://github.com/nfnt/resize) - Image scaling for sprite themes

## Platform Support

| Platform | Requirements |
|----------|-------------|
| macOS | macOS 10.13 (High Sierra) or later |
| Windows | Windows 10 or later |
| Linux | X11 or Wayland with OpenGL support |

## License

This project is provided under the GNU GPL-3.0 license. See [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Ideas for contributions:
- New sprite themes (animals, sports teams, holidays, etc.)
- Additional sound effects
- Game modes (speed runs, maze variants)
- Bug fixes and improvements

Please feel free to submit issues or pull requests.

## Author

Allan Marillier

## Acknowledgments

- Built with [Fyne](https://fyne.io/) - An easy-to-use GUI toolkit for Go
- Inspired by the classic 1980 Namco arcade game
- Ghost AI based on the original PacMan behavior patterns
