// Copyright 2016 Hajime Hoshi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package font

import (
	"image"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var (
	ArcadeFont *Font
)

type Font struct {
	image          *ebiten.Image
	offset         int
	charNumPerLine int
	charWidth      int
	charHeight     int
}

func (f *Font) TextWidth(str string) int {
	// TODO: Take care about '\n'
	return f.charWidth * len(str)
}

func (f *Font) TextHeight(str string) int {
	// TODO: Take care about '\n'
	return f.charHeight
}

func init() {
	eimg, _, err := ebitenutil.NewImageFromFile("arcadefont.png")
	if err != nil {
		panic(err)
	}
	ArcadeFont = &Font{eimg, 32, 16, 8, 8}
}

func (f *Font) DrawText(rt *ebiten.Image, str string, ox, oy, scale int, c color.Color) {
	for i := 0; i < len(str); i++ {
		op := &ebiten.DrawImageOptions{}
		dstx := i - strings.LastIndex(str[:i], "\n") - 1
		dstY := strings.Count(str[:i], "\n")
		dstx *= f.charWidth
		dstY *= f.charHeight
		if dstx < 0 {
			continue
		}
		op.GeoM.Translate(float64(dstx), float64(dstY))
		op.GeoM.Scale(float64(scale), float64(scale))
		op.GeoM.Translate(float64(ox), float64(oy))

		op.ColorScale.ScaleWithColor(c)

		code := int(str[i])
		if code == '\n' {
			continue
		}
		srcX := (code % f.charNumPerLine) * f.charWidth
		srcY := ((code - f.offset) / f.charNumPerLine) * f.charHeight
		rt.DrawImage(f.image.SubImage(image.Rect(srcX, srcY, srcX+f.charWidth, srcY+f.charHeight)).(*ebiten.Image), op)
	}
}

func (f *Font) DrawTextWithShadow(rt *ebiten.Image, str string, x, y, scale int, clr color.Color) {
	f.DrawText(rt, str, x+1, y+1, scale, color.RGBA{0, 0, 0, 0x80})
	f.DrawText(rt, str, x, y, scale, clr)
}
