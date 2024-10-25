package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/png"
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	cellSize            = 16
	sampleRate          = 48000
	defaultWindowWidth  = 1080
	defaultWindowHeight = 720
	spreadFactor        = 2.0 // Mayor número = más dispersión
	centerBias          = 0.5 // 0.5 = centrado, ajustar para desplazar el centro
)

var (
	ErrOutOfBounds       = errors.New("position is out of bounds")
	ErrInvalidDifficulty = errors.New("invalid game difficulty")
	ErrAssetNotFound     = errors.New("game asset not found")
)

type GameState int

const (
	Playing GameState = iota
	Won
	Lost
)

type DificultyLevel int

const (
	Easy DificultyLevel = iota
	Medium
	Hard
)

var difficultyLevels = map[DificultyLevel]GameDifficulty{
	Easy: {
		GridDimensions: GridDimensions{Cols: 9, Rows: 9},
		NumberOfMines:  10,
	},

	Medium: {
		GridDimensions: GridDimensions{Cols: 16, Rows: 16},
		NumberOfMines:  40,
	},

	Hard: {
		GridDimensions: GridDimensions{Cols: 30, Rows: 16},
		NumberOfMines:  99,
	},
}

type AudioManager struct {
	context *audio.Context
	sounds  map[string]*audio.Player
}

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

type Game struct {
	Board         map[Coordinates]CellState
	MinePositions map[Coordinates]bool
	AudioManager  AudioManager
	Sprite        Sprite
	Difficulty    GameDifficulty
	State         GameState
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

func NewAudioManager() (*AudioManager, error) {
	context := audio.NewContext(sampleRate)
	return &AudioManager{
		context: context,
		sounds:  make(map[string]*audio.Player),
	}, nil
}

func (am *AudioManager) LoadSound(name string, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open audio file: %w", err)
	}

	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read audio file: %w", err)
	}

	reader := bytes.NewReader(data)
	decoded, err := mp3.DecodeWithSampleRate(sampleRate, reader)
	if err != nil {
		return fmt.Errorf("failed to decode audio file: %w", err)
	}

	player, err := am.context.NewPlayer(decoded)
	if err != nil {
		return fmt.Errorf("failed to create audio player: %w", err)
	}

	am.sounds[name] = player
	return nil
}

func (am *AudioManager) PlaySound(name string) error {
	player, ok := am.sounds[name]
	if !ok {
		return ErrAssetNotFound
	}

	player.Rewind()
	player.Play()
	return nil
}

func NewGame(level DificultyLevel) (*Game, error) {
	difficulty, ok := difficultyLevels[level]
	if !ok {
		return nil, ErrInvalidDifficulty
	}

	// TODO: change name of the function to NewSprite
	sprite, err := LoadSprite()
	if err != nil {
		return nil, fmt.Errorf("failed to load sprites: %w", err)
	}

	audioManager, err := NewAudioManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize audio: %w", err)
	}

	game := &Game{
		Board:         make(map[Coordinates]CellState),
		MinePositions: make(map[Coordinates]bool),
		Difficulty:    difficulty,
		AudioManager:  *audioManager,
		State:         Playing,
		Sprite:        sprite,
	}

	// TODO: mabye shoudl handle error
	game.InitializeBoard()

	return game, nil
}

func (g *Game) InitializeBoard() {
	// TODO: Implement Error handling
	g.GenerateMinePositions()
	g.createBoard()
}

func (g *Game) createBoard() {
	grid := g.Difficulty.GridDimensions

	for x := 0; x < grid.Cols; x++ {
		for y := 0; y < grid.Rows; y++ {
			pos := Coordinates{X: x, Y: y}
			g.CalculateMinesAround(pos)

		}
	}
}

func (g *Game) RevealAllMines() {
	for minesPos := range g.MinePositions {
		cellState := g.Board[minesPos]
		if !cellState.isFlag && !cellState.isRevealed {
			cellState.isRevealed = true
			g.Board[minesPos] = cellState
		}
	}
}

func (g *Game) HandleMineClicked(pos Coordinates) {
	cellState := g.Board[pos]
	cellState.isMineClicked = true
	cellState.isRevealed = true
	g.Board[pos] = cellState
	g.State = Lost
	g.RevealAllMines()
}

// TODO: Implement Restart
func (g *Game) Restart() {}

func (g *Game) RevealCell(pos Coordinates) error {
	// TODO: maybe this is no necessary here because
	// this should be handle in the Update method
	if g.State != Playing {
		return nil
	}

	// TODO: this should be handle with an error inside of the function
	if g.isOutOfBounds(pos) {
		return nil
	}

	cellState := g.Board[pos]

	if cellState.isRevealed || cellState.isFlag {
		return nil
	}

	if cellState.isMine {
		g.HandleMineClicked(pos)
		return nil
	}

	g.RevealCellChain(pos)

	return nil
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
		g.RevealAllMines()
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
		g.Board[position] = CellState{isMine: true, isFlag: true, isRevealed: false, minesAround: 0}

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

func (g *Game) GenerateMinePositions() {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	// TODO: maybe is not necessary to store this in a variable
	grid := g.Difficulty.GridDimensions
	numberOfMines := g.Difficulty.NumberOfMines
	mines := g.MinePositions

	for len(mines) < numberOfMines {
		x := int(rnd.NormFloat64()*float64(grid.Cols)/spreadFactor +
			float64(grid.Cols)*centerBias)
		y := int(rnd.NormFloat64()*float64(grid.Rows)/spreadFactor +
			float64(grid.Rows)*centerBias)
		pos := Coordinates{X: x, Y: y}

		if !g.isOutOfBounds(pos) && !mines[pos] {
			mines[pos] = true
		}
	}
}

// TODO: maybe should be call ValidatePosition
func (g *Game) isOutOfBounds(position Coordinates) bool {
	return position.X < 0 || position.Y < 0 || position.X >= g.Difficulty.GridDimensions.Rows || position.Y >= g.Difficulty.GridDimensions.Cols
}

func (g *Game) Update() error {
	if g.State != Playing {
		return nil
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		// TODO: Refactor this
		x, y := ebiten.CursorPosition()
		position := Coordinates{x / cellSize, y / cellSize}

		if g.isOutOfBounds(position) {
			return nil
		}

		cellState := g.Board[position]
		if cellState.minesAround == 0 && !cellState.isMine && !cellState.isRevealed && !cellState.isFlag {
			g.AudioManager.PlaySound("totalmenchi")
		}

		g.RevealCell(position)
		g.CheckVictory()
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

func (g *Game) CheckVictory() bool {
	if g.State != Playing {
		return false
	}
	for _, cellState := range g.Board {
		if cellState.isMine {
			if !cellState.isFlag {
				return false
			}
		} else {
			if !cellState.isRevealed {
				return false
			}
		}
	}
	g.State = Won
	// TODO: should collect some statistics
	return true
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

func (g *Game) CalculateScreenDimensions() (width, height int) {
	width = g.Difficulty.GridDimensions.Rows * cellSize
	height = g.Difficulty.GridDimensions.Cols * cellSize
	return width, height
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.CalculateScreenDimensions()
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

func main() {
	ebiten.SetWindowSize(defaultWindowWidth, defaultWindowHeight)
	ebiten.SetWindowTitle("MineSweeper")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game, err := NewGame(Medium)
	if err != nil {
		log.Fatal(err)
	}

	game.AudioManager.LoadSound("totalmenchi", "./assets/sounds/totalmenchi.mp3")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
