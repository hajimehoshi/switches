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
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
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
	eimg, _, err := ebitenutil.NewImageFromFile("arcadefont.png", ebiten.FilterNearest)
	if err != nil {
		panic(err)
	}
	ArcadeFont = &Font{eimg, 32, 16, 8, 8}
}

type fontImageParts struct {
	str  string
	font *Font
}

func (f *fontImageParts) Len() int {
	return len(f.str)
}

func (f *fontImageParts) Dst(i int) (x0, y0, x1, y1 int) {
	x := i - strings.LastIndex(f.str[:i], "\n") - 1
	y := strings.Count(f.str[:i], "\n")
	x *= f.font.charWidth
	y *= f.font.charHeight
	if x < 0 {
		return 0, 0, 0, 0
	}
	return x, y, x + f.font.charWidth, y + f.font.charHeight
}

func (f *fontImageParts) Src(i int) (x0, y0, x1, y1 int) {
	code := int(f.str[i])
	if code == '\n' {
		return 0, 0, 0, 0
	}
	x := (code % f.font.charNumPerLine) * f.font.charWidth
	y := ((code - f.font.offset) / f.font.charNumPerLine) * f.font.charHeight
	return x, y, x + f.font.charWidth, y + f.font.charHeight
}

func (f *Font) DrawText(rt *ebiten.Image, str string, ox, oy, scale int, c color.Color) error {
	op := &ebiten.DrawImageOptions{
		ImageParts: &fontImageParts{str, f},
	}
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(float64(ox), float64(oy))

	ur, ug, ub, ua := c.RGBA()
	const max = math.MaxUint16
	r := float64(ur) / max
	g := float64(ug) / max
	b := float64(ub) / max
	a := float64(ua) / max
	if 0 < a {
		r /= a
		g /= a
		b /= a
	}
	op.ColorM.Scale(r, g, b, a)

	return rt.DrawImage(f.image, op)
}

func (f *Font) DrawTextWithShadow(rt *ebiten.Image, str string, x, y, scale int, clr color.Color) error {
	if err := f.DrawText(rt, str, x+1, y+1, scale, color.RGBA{0, 0, 0, 0x80}); err != nil {
		return err
	}
	if err := f.DrawText(rt, str, x, y, scale, clr); err != nil {
		return err
	}
	return nil
}
