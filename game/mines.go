package minesweeper

import (
	"math/rand"
	"time"
)

const (
	spreadFactor = 2.0
	centerBias   = 0.5
)

var MineNeighbors = []Coordinates{
	{X: -1, Y: -1},
	{X: -1, Y: 0},
	{X: -1, Y: 1},
	{X: 0, Y: -1},
	{X: 0, Y: 1},
	{X: 1, Y: -1},
	{X: 1, Y: 0},
	{X: 1, Y: 1},
}

type MineGenerator struct {
	rnd    *rand.Rand
	grid   BoardDimension
	bias   float64
	spread float64
}

func newMineGenerator(grid BoardDimension) *MineGenerator {
	return &MineGenerator{
		rnd:    rand.New(rand.NewSource(time.Now().UnixNano())),
		grid:   grid,
		bias:   centerBias,
		spread: spreadFactor,
	}
}

func (m *MineGenerator) generateMinePosition() Coordinates {
	x := int(m.rnd.NormFloat64()*float64(m.grid.Cols)/m.spread + float64(m.grid.Cols)*m.bias)
	y := int(m.rnd.NormFloat64()*float64(m.grid.Rows)/m.spread + float64(m.grid.Rows)*m.bias)
	return Coordinates{X: x, Y: y}
}

func (g *Game) initializeMineLocations() {
	numberOfMines := g.Difficulty.NumberOfMines
	mines := g.Board.Mines
	gen := newMineGenerator(g.Difficulty.BoardDimension)
	for len(mines) < numberOfMines {
		coord := gen.generateMinePosition()
		isNotFirstMove := g.Statistics.MoveTracker.Count != 0 || coord != g.Statistics.MoveTracker.FirstPosition
		isNotRepeatedMineCoord := !mines[coord]
		if !g.isOutOfBounds(coord) && isNotRepeatedMineCoord && isNotFirstMove {
			mines[coord] = true
		}
	}
}

func (g *Game) calculateMinesAround(coord Coordinates) {
	mines := g.Board.Mines

	if !mines[coord] {
		return
	}

	cell := g.Board.Cells[coord]
	cell.IsMine = true
	g.Board.Cells[coord] = cell

	for _, neighborCoord := range MineNeighbors {
		neighborPos := Coordinates{X: coord.X + neighborCoord.X, Y: coord.Y + neighborCoord.Y}

		if g.isOutOfBounds(neighborPos) {
			continue
		}

		cellState := g.Board.Cells[neighborPos]
		cellState.MinesAround++
		g.Board.Cells[neighborPos] = cellState
	}
}
