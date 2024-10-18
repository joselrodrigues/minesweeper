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
	Dificulty   GameDifficulty
	Minesweeper *Minesweeper
	Sprite      Sprite
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
	Image *ebiten.Image
	Size  int
}

type GridDimensions struct {
	Cols int
	Rows int
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
	m.Board = make(map[Coordinates]CellState)
	mines := generateMinePositions(dificulty.GridDimensions, dificulty.NumberOfMines)

	for i := 0; i < dificulty.GridDimensions.Cols; i++ {
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

func (m *Minesweeper) UpdateCellState(position Coordinates, state CellState) {
	m.Board[position] = state
}

func generateMinePositions(dimension GridDimensions, numberOfMines int) map[Coordinates]bool {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	mines := make(map[Coordinates]bool)

	for len(mines) < numberOfMines {
		x := rnd.Intn(dimension.Cols)
		y := rnd.Intn(dimension.Rows)
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
		cellState := g.Minesweeper.Board[position]
		fmt.Println(cellState)
		cellState.isRevealed = true
		g.Minesweeper.UpdateCellState(
			position, cellState)

	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
	}
	return nil
}

func (g *Game) RenderBoard(screen *ebiten.Image) {
	for boardPosition, cellState := range g.Minesweeper.Board {
		var baseSpriteCell *ebiten.Image
		opts := &ebiten.DrawImageOptions{}

		tx := float64(boardPosition.X * g.Sprite.Size)
		ty := float64(boardPosition.Y * g.Sprite.Size)

		opts.GeoM.Translate(tx, ty)

		if !cellState.isRevealed {
			baseSpriteCell = g.Sprite.Image.SubImage(image.Rect(0, 0, g.Sprite.Size, g.Sprite.Size)).(*ebiten.Image)
		} else if cellState.isMine && cellState.isRevealed {
			baseSpriteCell = g.Sprite.Image.SubImage(image.Rect(cellSize*6, 0, g.Sprite.Size*7, g.Sprite.Size)).(*ebiten.Image)
		} else if !cellState.isMine && cellState.minesAround == 0 && cellState.isRevealed {
			baseSpriteCell = g.Sprite.Image.SubImage(image.Rect(cellSize, 0, g.Sprite.Size*2, g.Sprite.Size)).(*ebiten.Image)
		}
		screen.DrawImage(baseSpriteCell, opts)

	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.RenderBoard(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

func LoadSprite() (Sprite, error) {
	spriteImage, _, err := ebitenutil.NewImageFromFile("assets/sprites/board.png")
	sprite := Sprite{Image: spriteImage, Size: cellSize}

	return sprite, err
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Hello, MineSweeper go!")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	game := Minesweeper{}
	game.createBoard(medium)

	sprite, err := LoadSprite()
	if err != nil {
		log.Fatal(err)
	}

	if err := ebiten.RunGame(&Game{Minesweeper: &game, Sprite: sprite}); err != nil {
		log.Fatal(err)
	}
}
