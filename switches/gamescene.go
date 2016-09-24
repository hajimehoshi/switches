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

package switches

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/switches/switches/internal/font"
)

type player struct {
	x         int
	y         int
	z         int
	dir       dir
	moveCount int
}

type gameScene struct {
	game          *Game
	field         *field
	player        *player
	tilesImage    *ebiten.Image
	switchStates  []bool
	selectedTileX int
	selectedTileY int
}

func newGameScene(width, height, depth, switches int, game *Game) (*gameScene, error) {
	f, err := newField(width, height, depth, switches)
	if err != nil {
		return nil, err
	}
	tilesImage, _, err := ebitenutil.NewImageFromFile("tiles.png", ebiten.FilterNearest)
	if err != nil {
		return nil, err
	}
	px, py := f.start()
	s := &gameScene{
		game:         game,
		field:        f,
		player:       &player{x: px, y: py, z: 0},
		tilesImage:   tilesImage,
		switchStates: make([]bool, switches),
	}
	return s, nil
}

func (s *gameScene) Update() error {
	s.updateSelectedTile()
	if s.game.input.IsTriggered() {
		tile, _ := s.field.tile(s.selectedTileX, s.selectedTileY, s.player.z, s.switchStates)
		if !tile.isPassable() {
			return nil
		}
		passable := func(x, y int) bool {
			x0, y0, x1, y1 := s.tileRangeInScreen()
			w, h, _ := s.field.tileSize()
			if x < x0 || x1 <= x || y < y0 || y1 <= y {
				return false
			}
			if x < 0 || w <= x || y < 0 || h <= y {
				return false
			}
			t, _ := s.field.tile(x, y, s.player.z, s.switchStates)
			// Don't go through switches.
			if t == tileSwitch0 || t == tileSwitch1 {
				return x == s.selectedTileX && y == s.selectedTileY
			}
			return t.isPassable()
		}
		path := calcPath(passable, s.player.x, s.player.y, s.selectedTileX, s.selectedTileY)
		if len(path) == 0 {
			return nil
		}
		i := 0
		x, y := s.player.x, s.player.y
		var moveTask task
		s.game.appendTask(func() error {
			if len(path) <= i {
				return taskTerminated
			}
			if moveTask == nil {
				d := path[i]
				switch d {
				case dirLeft:
					x--
				case dirRight:
					x++
				case dirUp:
					y--
				case dirDown:
					y++
				}
				moveTask = s.moveTask(d, x, y)
			}
			if err := moveTask(); err == nil {
				return nil
			} else if err != taskTerminated {
				return err
			}
			moveTask = nil
			i++
			switch t, _ := s.field.tile(x, y, s.player.z, s.switchStates); t {
			case tileSwitch0:
				fallthrough
			case tileSwitch1:
				return taskTerminated
			}
			return nil
		})
		return nil
	}
	// Move the player
	tile, _ := s.field.tile(s.player.x, s.player.y, s.player.z, s.switchStates)
	nx, ny := s.player.x, s.player.y
	w, h, _ := s.field.tileSize()
	var dir dir
	if !tile.oneWay() {
		if ebiten.IsKeyPressed(ebiten.KeyLeft) || tile == tileOneWayLeft {
			nx = max(s.player.x-1, 0)
			dir = dirLeft
		} else if ebiten.IsKeyPressed(ebiten.KeyRight) || tile == tileOneWayRight {
			nx = min(s.player.x+1, w-1)
			dir = dirRight
		} else if ebiten.IsKeyPressed(ebiten.KeyUp) || tile == tileOneWayUp {
			ny = max(s.player.y-1, 0)
			dir = dirUp
		} else if ebiten.IsKeyPressed(ebiten.KeyDown) || tile == tileOneWayDown {
			ny = min(s.player.y+1, h-1)
			dir = dirDown
		}
	}
	if s.player.x == nx && s.player.y == ny {
		return nil
	}
	if t, _ := s.field.tile(nx, ny, s.player.z, s.switchStates); !t.isPassable() {
		return nil
	}
	s.game.appendTask(s.moveTask(dir, nx, ny))
	return nil
}

func (s *gameScene) updateSelectedTile() {
	x, y := ebiten.CursorPosition()
	ox, oy := s.tileOffset()
	x0, y0, _, _ := s.tileRangeInScreen()
	s.selectedTileX = x0 + (x-ox)/gridSize
	s.selectedTileY = y0 + (y-oy)/gridSize
}

func (s *gameScene) moveTask(dir dir, nextX, nextY int) task {
	started := false
	return func() error {
		if !started {
			s.player.dir = dir
			s.player.moveCount = playerMaxMoveCount
			started = true
		}
		if 0 < s.player.moveCount {
			s.player.moveCount--
		}
		if 0 < s.player.moveCount {
			return nil
		}
		s.player.x = nextX
		s.player.y = nextY
		switch t, sw := s.field.tile(nextX, nextY, s.player.z, s.switchStates); t {
		case tileUpstairs:
			fallthrough
		case tileOneWayUpstairs:
			s.player.z -= 1
		case tileDownstairs:
			fallthrough
		case tileOneWayDownstairs:
			s.player.z += 1
		case tileSwitch0:
			fallthrough
		case tileSwitch1:
			wait := 10
			s.game.appendTask(func() error {
				if 0 < wait {
					wait--
					return nil
				}
				s.switchStates[sw] = !s.switchStates[sw]
				return taskTerminated
			})
		}
		return taskTerminated
	}
}

func (s *gameScene) tileRangeInScreen() (int, int, int, int) {
	nx := screenWidth / gridSize
	ny := screenHeight / gridSize
	x0 := s.player.x - nx/2 - 1
	y0 := s.player.y - ny/2 - 1
	x1 := s.player.x + nx/2 + 1
	y1 := s.player.y + ny/2 + 1
	return x0, y0, x1, y1
}

func (s *gameScene) tileOffset() (int, int) {
	ox, oy := -gridSize/2-gridSize, -gridSize/2-gridSize
	if 0 < s.player.moveCount {
		d := gridSize * (playerMaxMoveCount - s.player.moveCount) / playerMaxMoveCount
		switch s.player.dir {
		case dirLeft:
			ox += d
		case dirRight:
			ox -= d
		case dirUp:
			oy += d
		case dirDown:
			oy -= d
		}
	}
	return ox, oy
}

const (
	gridSize           = 16
	playerMaxMoveCount = 4
)

type tilePart struct {
	srcX, srcY int
	dstX, dstY int
}

func (p *tilePart) Len() int { return 1 }
func (p *tilePart) Dst(i int) (int, int, int, int) {
	return p.dstX, p.dstY, p.dstX + gridSize, p.dstY + gridSize
}
func (p *tilePart) Src(i int) (int, int, int, int) {
	return p.srcX, p.srcY, p.srcX + gridSize, p.srcY + gridSize
}

type switchLetter struct {
	letter rune
	color  switchLetterColor
	x      int
	y      int
}

type switchLetterColor color.Color

var (
	switchLetterColor0 switchLetterColor = color.RGBA{0x75, 0x75, 0x75, 0xff}
	switchLetterColor1 switchLetterColor = color.RGBA{0xee, 0xee, 0xee, 0xff}
	switchLetterColor2 switchLetterColor = color.RGBA{0xff, 0xf5, 0x9e, 0xff}
	switchLetterColor3 switchLetterColor = color.RGBA{0x4e, 0x6c, 0xef, 0xff}
)

type tileParts struct {
	scene   *gameScene
	dst     []int
	src     []int
	skips   map[int]struct{}
	letters []*switchLetter
}

func newTileParts(scene *gameScene) *tileParts {
	p := &tileParts{
		scene: scene,
	}
	x0, y0, x1, y1 := scene.tileRangeInScreen()
	sw := x1 - x0 + 1
	l := sw * (y1 - y0 + 1)
	p.dst = make([]int, l*2)
	p.src = make([]int, l*2)
	p.skips = map[int]struct{}{}
	player := p.scene.player
	for i := 0; i < l; i++ {
		x := x0 + i/sw
		y := y0 + i%sw
		if x < 0 || y < 0 {
			p.skips[i] = struct{}{}
			continue
		}
		w, h, _ := p.scene.field.tileSize()
		if w <= x || h <= y {
			p.skips[i] = struct{}{}
			continue
		}
		ox, oy := p.scene.tileOffset()
		dx := (i/sw)*gridSize + ox
		dy := (i%sw)*gridSize + oy
		p.dst[2*i] = dx
		p.dst[2*i+1] = dy
		t, s := p.scene.field.tile(x, y, player.z, p.scene.switchStates)
		switch t {
		case tileNone:
			p.skips[i] = struct{}{}
			continue
		case tileSwitch0:
			fallthrough
		case tileSwitch1:
			clr := switchLetterColor0
			if p.scene.switchStates[s] {
				clr = switchLetterColor1
			}
			p.letters = append(p.letters, &switchLetter{
				letter: 'A' + rune(s),
				color:  clr,
				x:      dx + 4,
				y:      dy + 3,
			})
		case tileSwitchedTileValid:
			fallthrough
		case tileSwitchedTileInvalid:
			clr := switchLetterColor2
			if (p.scene.switchStates[s] && t == tileSwitchedTileValid) ||
				(!p.scene.switchStates[s] && t == tileSwitchedTileInvalid) {
				clr = switchLetterColor3
			}
			p.letters = append(p.letters, &switchLetter{
				letter: 'A' + rune(s),
				color:  clr,
				x:      dx + 4,
				y:      dy + 4,
			})
		}
		type position struct {
			X, Y int
		}
		pos := map[tile]position{
			tileNone:                {0, 0},
			tileRegular:             {1 * gridSize, 0},
			tileUpstairs:            {4 * gridSize, 0},
			tileDownstairs:          {2 * gridSize, 0},
			tileOneWayLeft:          {7 * gridSize, 0},
			tileOneWayRight:         {9 * gridSize, 0},
			tileOneWayUp:            {8 * gridSize, 0},
			tileOneWayDown:          {6 * gridSize, 0},
			tileOneWayUpstairs:      {5 * gridSize, 0},
			tileOneWayDownstairs:    {3 * gridSize, 0},
			tileSwitch0:             {10 * gridSize, 0},
			tileSwitch1:             {11 * gridSize, 0},
			tileSwitchedTileValid:   {1 * gridSize, 0},
			tileSwitchedTileInvalid: {0, 0},
		}[t]
		p.src[2*i] = pos.X
		p.src[2*i+1] = pos.Y
	}
	return p
}

func (p *tileParts) Len() int {
	return len(p.dst) / 2
}

func (p *tileParts) Dst(i int) (int, int, int, int) {
	if _, ok := p.skips[i]; ok {
		return 0, 0, 0, 0
	}
	x, y := p.dst[2*i], p.dst[2*i+1]
	return x, y, x + gridSize, y + gridSize
}

func (p *tileParts) Src(i int) (int, int, int, int) {
	if _, ok := p.skips[i]; ok {
		return 0, 0, 0, 0
	}
	x, y := p.src[2*i], p.src[2*i+1]
	return x, y, x + gridSize, y + gridSize
}

func (p *tileParts) switchLetters() []*switchLetter {
	return p.letters
}

func (s *gameScene) Draw(screen *ebiten.Image) error {
	if err := screen.Fill(backgroundColor); err != nil {
		return err
	}
	op := &ebiten.DrawImageOptions{}
	tileParts := newTileParts(s)
	op.ImageParts = tileParts
	if err := screen.DrawImage(s.tilesImage, op); err != nil {
		return err
	}
	if err := s.drawCursor(screen); err != nil {
		return err
	}
	for _, l := range tileParts.switchLetters() {
		if err := font.ArcadeFont.DrawText(screen, string(l.letter), l.x, l.y, 1, l.color); err != nil {
			return err
		}
	}
	if err := s.drawPlayer(screen); err != nil {
		return err
	}
	if err := s.drawFloorNumber(screen); err != nil {
		return err
	}
	return nil
}

func (s *gameScene) drawCursor(screen *ebiten.Image) error {
	ox, oy := s.tileOffset()
	x0, y0, _, _ := s.tileRangeInScreen()
	dstX := s.selectedTileX*gridSize - x0*gridSize + ox
	dstY := s.selectedTileY*gridSize - y0*gridSize + oy
	op := &ebiten.DrawImageOptions{}
	op.ImageParts = &tilePart{
		srcX: 16,
		srcY: 16,
		dstX: dstX,
		dstY: dstY,
	}
	if err := screen.DrawImage(s.tilesImage, op); err != nil {
		return err
	}
	return nil
}

func (s *gameScene) drawPlayer(screen *ebiten.Image) error {
	dstX := (screenWidth - gridSize) / 2
	dstY := (screenHeight - gridSize) / 2
	op := &ebiten.DrawImageOptions{}
	op.ImageParts = &tilePart{
		srcX: 0,
		srcY: 16,
		dstX: dstX,
		dstY: dstY,
	}
	if err := screen.DrawImage(s.tilesImage, op); err != nil {
		return err
	}
	return nil
}

func (s *gameScene) drawFloorNumber(screen *ebiten.Image) error {
	z := s.player.z
	msg := ""
	if z == 0 {
		msg = "GROUND"
	} else {
		msg = fmt.Sprintf("B%dF", z)
	}
	x := 8
	y := 8
	if err := font.ArcadeFont.DrawTextWithShadow(screen, msg, x, y, 1, color.White); err != nil {
		return err
	}
	return nil
}
