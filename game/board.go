package minesweeper

func (g *Game) initializeEmptyBoard() {
	for Row := 0; Row < g.Difficulty.BoardDimension.Rows; Row++ {
		for Col := 0; Col < g.Difficulty.BoardDimension.Cols; Col++ {
			g.Board.Cells[Coordinates{X: Row, Y: Col}] = Cell{State: Hidden, IsMine: false, MinesAround: 0}
		}
	}
}

func (g *Game) isOutOfBounds(coor Coordinates) bool {
	return coor.X < 0 || coor.X >= g.Difficulty.BoardDimension.Rows || coor.Y < 0 || coor.Y >= g.Difficulty.BoardDimension.Cols
}
