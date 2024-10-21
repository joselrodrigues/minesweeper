package main

import (
	"fmt"
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

var PositionNeighbors = []Coordinates{
	{X: -1, Y: -1},
	{X: -1, Y: 0},
	{X: -1, Y: 1},
	{X: 0, Y: -1},
	{X: 0, Y: 1},
	{X: 1, Y: -1},
	{X: 1, Y: 0},
	{X: 1, Y: 1},
}

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
	EndGame       bool
}

type Coordinates struct {
	X, Y int
}

// TODO: keep CellState invariants (keep the struct consistent)
// example: if isMine is true, minesAround should be 0
// example: if isRevealed is true, isFlag should be false
// example: if isFlag is true, isRevealed should be false
type CellState struct {
	minesAround   int
	isMine        bool
	isFlag        bool
	isRevealed    bool
	isMineClicked bool
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

	for x := 0; x < grid.Cols; x++ {
		for y := 0; y < grid.Rows; y++ {
			pos := Coordinates{X: x, Y: y}
			g.CalculateMinesAround(pos)

		}
	}
}

func (g *Game) GameOver() {
	g.EndGame = true
	for minesPos := range g.MinePositions {

		cellState := g.Board[minesPos]
		if cellState.isMine && !cellState.isFlag && !cellState.isRevealed {
			cellState.isRevealed = true
			g.Board[minesPos] = cellState
		}
	}
}

func (g *Game) RevealCellChain(position Coordinates) {
	if g.isOutOfBounds(position) {
		return
	}

	cellState := g.Board[position]

	if cellState.isRevealed || cellState.isFlag {
		return
	}

	if cellState.isMine {
		cellState.isMineClicked = true
		g.Board[position] = cellState
		g.GameOver()
		return
	}

	if cellState.minesAround > 0 {
		cellState.isRevealed = true
		g.Board[position] = cellState
		return
	}

	for _, neighbor := range PositionNeighbors {
		neighborPos := Coordinates{X: position.X + neighbor.X, Y: position.Y + neighbor.Y}

		if g.isOutOfBounds(neighborPos) {
			continue
		}

		cellState.isRevealed = true
		g.Board[position] = cellState
		if cellState.minesAround == 0 && !g.Board[neighborPos].isMine {
			g.RevealCellChain(neighborPos)
		}
	}
}

func (g *Game) CalculateMinesAround(position Coordinates) {
	mines := g.MinePositions

	if mines[position] {
		g.Board[position] = CellState{isMine: true, isFlag: false, isRevealed: false, minesAround: 0}

		for _, neighbor := range PositionNeighbors {
			neighborPos := Coordinates{X: position.X + neighbor.X, Y: position.Y + neighbor.Y}

			if g.isOutOfBounds(neighborPos) {
				continue
			}

			if _, exists := g.Board[neighborPos]; !exists {
				g.Board[neighborPos] = CellState{isMine: false, isFlag: false, isRevealed: false, minesAround: 0}
			}

			if _, exists := mines[neighborPos]; !exists {
				cellState := g.Board[neighborPos]
				cellState.minesAround++
				g.Board[neighborPos] = cellState
			}
		}
	}

	if _, exists := g.Board[position]; !exists {
		g.Board[position] = CellState{isMine: false, isFlag: false, isRevealed: false, minesAround: 0}
		return
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

func (g *Game) isOutOfBounds(position Coordinates) bool {
	return position.X < 0 || position.Y < 0 || position.X >= g.Dificulty.GridDimensions.Rows || position.Y >= g.Dificulty.GridDimensions.Cols
}

func (g *Game) Update() error {
	if g.EndGame {
		return nil
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		// TODO: Refactor this
		x, y := ebiten.CursorPosition()
		position := Coordinates{x / cellSize, y / cellSize}

		if g.isOutOfBounds(position) {
			return nil
		}

		g.RevealCellChain(position)
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		// TODO: Refactor this
		x, y := ebiten.CursorPosition()
		position := Coordinates{x / cellSize, y / cellSize}

		if g.isOutOfBounds(position) {
			return nil
		}

		cellState := g.Board[position]
		cellState.isFlag = !cellState.isFlag
		g.Board[position] = cellState

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
		case cellState.isFlag && !cellState.isRevealed:
			baseSpriteCell = g.Sprite.Image["flag"]
		case !cellState.isRevealed:
			baseSpriteCell = g.Sprite.Image["hidden"]
		case cellState.isMineClicked:
			baseSpriteCell = g.Sprite.Image["mineClicked"]
		case cellState.isMine:
			baseSpriteCell = g.Sprite.Image["mine"]
		case cellState.minesAround > 0:
			baseSpriteCell = g.Sprite.Image[fmt.Sprintf("number_%d", cellState.minesAround)]
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
		"hidden":      spriteSheet.SubImage(image.Rect(0, 0, cellSize, cellSize)).(*ebiten.Image),
		"flag":        spriteSheet.SubImage(image.Rect(cellSize*2, 0, cellSize*3, cellSize)).(*ebiten.Image),
		"mineClicked": spriteSheet.SubImage(image.Rect(cellSize*6, 0, cellSize*7, cellSize)).(*ebiten.Image),
		"mine":        spriteSheet.SubImage(image.Rect(cellSize*5, 0, cellSize*6, cellSize)).(*ebiten.Image),
		"empty":       spriteSheet.SubImage(image.Rect(cellSize, 0, cellSize*2, cellSize)).(*ebiten.Image),
	}

	for spriteNumb := 0; spriteNumb < 8; spriteNumb++ {
		images[fmt.Sprintf("number_%d", spriteNumb+1)] = spriteSheet.SubImage(image.Rect(cellSize*spriteNumb, cellSize, cellSize*(spriteNumb+1), cellSize*2)).(*ebiten.Image)
	}

	return Sprite{Image: images}, err
}

// TODO: Create mines in board using random normal distribution

func InitGame() (Game, error) {
	sprite, err := LoadSprite()
	initialMinePositions := make(map[Coordinates]bool)
	initialBoard := make(map[Coordinates]CellState)

	game := Game{
		Dificulty:     medium,
		Board:         initialBoard,
		Sprite:        sprite,
		EndGame:       false,
		MinePositions: initialMinePositions,
	}

	game.GenerateMinePositions()
	game.createBoard()

	return game, err
}

func main() {
	ebiten.SetWindowSize(1080, 720)
	ebiten.SetWindowTitle("Hello, MineSweeper go!")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game, err := InitGame()
	if err != nil {
		log.Fatal(err)
	}

	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
