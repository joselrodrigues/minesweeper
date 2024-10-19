package main

import (
	"image"
	_ "image/png"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	cellSize = 16
)

var easy = GameDifficulty{
	GridDimensions: GridDimensions{Cols: 9, Rows: 9},
	NumberOfMines:  10,
}

var medium = GameDifficulty{
	GridDimensions: GridDimensions{Cols: 16, Rows: 16},
	NumberOfMines:  40,
}

var hard = GameDifficulty{
	GridDimensions: GridDimensions{Cols: 30, Rows: 16},
	NumberOfMines:  99,
}

type Game struct {
	Board         map[Coordinates]CellState
	MinePositions map[Coordinates]bool
	Sprite        Sprite
	Dificulty     GameDifficulty
}

type Coordinates struct {
	X, Y int
}

type CellState struct {
	minesAround int
	isMine      bool
	isFlag      bool
	isRevealed  bool
}

type Sprite struct {
	Image map[string]*ebiten.Image
}

type GridDimensions struct {
	Cols int
	Rows int
}

type GameDifficulty struct {
	NumberOfMines  int
	GridDimensions GridDimensions
}

func (g *Game) createBoard() {
	grid := g.Dificulty.GridDimensions
	mines := g.GenerateMinePositions()

	for x := 0; x < grid.Cols; x++ {
		for y := 0; y < grid.Rows; y++ {
			pos := Coordinates{X: x, Y: y}
			if _, exists := mines[pos]; exists {
				g.Board[pos] = CellState{isMine: true, isFlag: false, isRevealed: false, minesAround: 0}
			} else {
				g.Board[pos] = CellState{isMine: false, isFlag: false, isRevealed: false, minesAround: 0}
			}
		}
	}
}

func (g *Game) GenerateMinePositions() map[Coordinates]bool {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	grid := g.Dificulty.GridDimensions
	numberOfMines := g.Dificulty.NumberOfMines
	mines := g.MinePositions

	for len(mines) < numberOfMines {
		x := rnd.Intn(grid.Cols)
		y := rnd.Intn(grid.Rows)
		pos := Coordinates{X: x, Y: y}

		if _, exists := mines[pos]; !exists {
			mines[pos] = true
		}
	}
	return mines
}

func (g *Game) Update() error {
	x, y := ebiten.CursorPosition()
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		position := Coordinates{x / cellSize, y / cellSize}
		cellState := g.Board[position]
		cellState.isRevealed = true
		g.Board[position] = cellState
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
	}
	return nil
}

func (g *Game) RenderBoard(screen *ebiten.Image) {
	for boardPosition, cellState := range g.Board {
		var baseSpriteCell *ebiten.Image
		opts := &ebiten.DrawImageOptions{}

		tx := float64(boardPosition.X * cellSize)
		ty := float64(boardPosition.Y * cellSize)

		opts.GeoM.Translate(tx, ty)

		switch {
		case !cellState.isRevealed:
			baseSpriteCell = g.Sprite.Image["hidden"]
		case cellState.isMine:
			baseSpriteCell = g.Sprite.Image["mine"]
		default:
			baseSpriteCell = g.Sprite.Image["empty"]
		}

		screen.DrawImage(baseSpriteCell, opts)

	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.RenderBoard(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 720, 480
}

func LoadSprite() (Sprite, error) {
	spriteSheet, _, err := ebitenutil.NewImageFromFile("assets/sprites/board.png")

	images := map[string]*ebiten.Image{
		"hidden": spriteSheet.SubImage(image.Rect(0, 0, cellSize, cellSize)).(*ebiten.Image),
		"flag":   spriteSheet.SubImage(image.Rect(cellSize*2, 0, cellSize*3, cellSize)).(*ebiten.Image),
		"mine":   spriteSheet.SubImage(image.Rect(cellSize*6, 0, cellSize*7, cellSize)).(*ebiten.Image),
		"empty":  spriteSheet.SubImage(image.Rect(cellSize, 0, cellSize*2, cellSize)).(*ebiten.Image),
	}

	return Sprite{Image: images}, err
}

func main() {
	ebiten.SetWindowSize(1080, 720)
	ebiten.SetWindowTitle("Hello, MineSweeper go!")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	sprite, err := LoadSprite()
	minePositions := make(map[Coordinates]bool)
	game := Game{
		Dificulty:     medium,
		Board:         make(map[Coordinates]CellState),
		Sprite:        sprite,
		MinePositions: minePositions,
	}
	game.createBoard()

	if err != nil {
		log.Fatal(err)
	}

	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
