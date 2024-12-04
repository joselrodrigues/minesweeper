package engine

import (
	"fmt"
	m "minesweeper/game"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	BoardOffsetX = 12
	BoardOffsetY = 55
	CellSize     = 16
)

func boardCoordToScreenPos(coord m.Coordinates) (float64, float64) {
	// X: borde izquierdo + (posición * tamaño de celda)
	xpos := float64(BoardOffsetX + (coord.X * CellSize))

	// Y: borde superior + (posición * tamaño de celda)
	ypos := float64(BoardOffsetY + (coord.Y * CellSize))

	return xpos, ypos
}

func RenderBoard(screen *ebiten.Image, g *m.Game) {
	for boardCoord, cellState := range g.Board.Cells {
		var baseSpriteCell *ebiten.Image
		opts := &ebiten.DrawImageOptions{}

		xpos, ypos := boardCoordToScreenPos(boardCoord)

		opts.GeoM.Translate(xpos, ypos)

		switch {
		case cellState.State == m.Flagged:
			baseSpriteCell = g.Sprite.Image["flag"]
		case cellState.State == m.Hidden:
			baseSpriteCell = g.Sprite.Image["hidden"]
		// case cellState.isMineSelected:
		// 	baseSpriteCell = g.Sprite.Image["mineSelected"]
		case cellState.State == m.Revealed:
			if cellState.IsMine {
				baseSpriteCell = g.Sprite.Image["mine"]
			} else if cellState.MinesAround > 0 {
				baseSpriteCell = g.Sprite.Image[fmt.Sprintf("number_%d", cellState.MinesAround)]
			} else {
				baseSpriteCell = g.Sprite.Image["empty"]
			}
		}

		screen.DrawImage(baseSpriteCell, opts)

	}
}
