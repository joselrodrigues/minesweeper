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
	screenWidth  = 320
	screenHeight = 240
	gridWidth    = 15 * 16
	gridHeight   = 15 * 16
	cellSize     = 16
)

var easy = GameDifficulty{
	GridDimensions: GridDimensions{Columns: 9, Rows: 9},
	NumberOfMines:  10,
}

var medium = GameDifficulty{
	GridDimensions: GridDimensions{Columns: 16, Rows: 16},
	NumberOfMines:  40,
}

var hard = GameDifficulty{
	GridDimensions: GridDimensions{Columns: 30, Rows: 16},
	NumberOfMines:  99,
}

// type Game struct {
// 	boardSprite *Sprite
// }

type Game struct {
	Dificulty GameDifficulty
}

// var cellImage, _, err = ebitenutil.NewImageFromFile("assets/sprites/board.png")

// func DrawHiddenCell(x int, y int, screen *ebiten.Image) {
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	opts := &ebiten.DrawImageOptions{}
//
// 	tx := float64(x * cellSize)
// 	ty := float64(y * cellSize)
// 	opts.GeoM.Translate(tx, ty)
//
// 	baseSpriteCell := cellImage.SubImage(image.Rect(0, 0, cellSize, cellSize)).(*ebiten.Image)
//
// 	screen.DrawImage(baseSpriteCell, opts)
// }

type Coordinates struct {
	X, Y int
}

type CellState struct {
	isMine      bool
	isFlag      bool
	isRevealed  bool
	minesAround int
}

type Sprite struct {
	image    *ebiten.Image
	position Coordinates
	size     int
}

type GridDimensions struct {
	Columns int
	Rows    int
}

type GameDifficulty struct {
	GridDimensions GridDimensions
	NumberOfMines  int
}

// quizás debería cambiar el nombre de la estructura
type Minesweeper struct {
	Board map[Coordinates]CellState
}

func (m *Minesweeper) createBoard(dificulty GameDifficulty) {
	mines := generateMinePositions(dificulty.GridDimensions, dificulty.NumberOfMines)

	for i := 0; i < dificulty.GridDimensions.Columns; i++ {
		for j := 0; j < dificulty.GridDimensions.Rows; j++ {
			pos := Coordinates{X: i, Y: j}
			if _, exists := mines[pos]; exists {
				m.Board[pos] = CellState{isMine: true, isFlag: false, isRevealed: false, minesAround: 0}
			} else {
				m.Board[pos] = CellState{isMine: false, isFlag: false, isRevealed: false, minesAround: 0}
			}
		}
	}
}

func generateMinePositions(dimension GridDimensions, numberOfMines int) map[Coordinates]bool {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	mines := make(map[Coordinates]bool)

	for len(mines) < numberOfMines {
		x := rnd.Intn(dimension.Columns)
		y := rnd.Intn(dimension.Rows)
		pos := Coordinates{X: x, Y: y}

		if _, exists := mines[pos]; !exists {
			mines[pos] = true
		}
	}
	return mines
}

func RenderBoard() {
	board := Minesweeper{}
	board.createBoard(medium)
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	board := Minesweeper{}
	board.createBoard(g.Dificulty)

	sprite, err := LoadSprite()
	if err != nil {
		log.Fatal(err)
	}

	opts := &ebiten.DrawImageOptions{}
	for boardPosition := range board.Board {
		tx := float64(boardPosition.X * sprite.size)
		ty := float64(boardPosition.Y * sprite.size)
		opts.GeoM.Translate(tx, ty)
	}

	baseSpriteCell := sprite.image.SubImage(image.Rect(0, 0, sprite.size, sprite.size)).(*ebiten.Image)
	screen.DrawImage(baseSpriteCell, opts)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

func LoadSprite() (*Sprite, error) {
	spriteImage, _, err := ebitenutil.NewImageFromFile("assets/sprites/board.png")
	sprite := Sprite{image: spriteImage, position: Coordinates{X: 0, Y: 0}, size: 16}

	return &sprite, err
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Hello, MineSweeper go!")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(&Game{Dificulty: medium}); err != nil {
		log.Fatal(err)
	}
}
