package engine

import (
	"embed"
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

//go:embed assets/sprites/*.png assets/sounds/*.mp3
var gameAssets embed.FS

func LoadSprite() (Sprite, error) {
	spriteSheet, _, err := ebitenutil.NewImageFromFileSystem(gameAssets, "assets/sprites/board.png")

	images := map[string]*ebiten.Image{
		"hidden":       spriteSheet.SubImage(image.Rect(0, 0, CellSize, CellSize)).(*ebiten.Image),
		"flag":         spriteSheet.SubImage(image.Rect(CellSize*2, 0, CellSize*3, CellSize)).(*ebiten.Image),
		"mineSelected": spriteSheet.SubImage(image.Rect(CellSize*6, 0, CellSize*7, CellSize)).(*ebiten.Image),
		"mine":         spriteSheet.SubImage(image.Rect(CellSize*5, 0, CellSize*6, CellSize)).(*ebiten.Image),
		"empty":        spriteSheet.SubImage(image.Rect(CellSize, 0, CellSize*2, CellSize)).(*ebiten.Image),
	}

	for spriteNumb := 0; spriteNumb < 8; spriteNumb++ {
		images[fmt.Sprintf("number_%d", spriteNumb+1)] = spriteSheet.SubImage(image.Rect(CellSize*spriteNumb, CellSize, CellSize*(spriteNumb+1), CellSize*2)).(*ebiten.Image)
	}

	return Sprite{Image: images}, err
}
