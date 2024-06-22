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
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

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
	goal          bool
}

func newGameScene(width, height, depth, switches int, game *Game) (*gameScene, error) {
	f, err := newField(width, height, depth, switches)
	if err != nil {
		return nil, err
	}
	tilesImage, _, err := ebitenutil.NewImageFromFile("tiles.png")
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
	tile, _ := s.field.tile(s.player.x, s.player.y, s.player.z, s.switchStates)
	if tile == tileGoal {
		s.goal = true
		if s.game.input.IsTriggered() {
			s.game.goTo(newTitleScene(s.game))
		}
		return nil
	}
	s.updateSelectedTile()
	if s.game.input.IsTriggered() {
		w, h, _ := s.field.tileSize()
		if s.selectedTileX < 0 || w <= s.selectedTileX || s.selectedTileY < 0 || h <= s.selectedTileY {
			return nil
		}
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
			tileGoal:                {12 * gridSize, 0},
		}[t]
		p.src[2*i] = pos.X
		p.src[2*i+1] = pos.Y
	}
	return p
}

func (p *tileParts) draw(screen *ebiten.Image, tilesImage *ebiten.Image) {
	for i := 0; i < len(p.dst)/2; i++ {
		if _, ok := p.skips[i]; ok {
			continue
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(p.dst[2*i]), float64(p.dst[2*i+1]))
		x, y := p.src[2*i], p.src[2*i+1]
		screen.DrawImage(tilesImage.SubImage(image.Rect(x, y, x+gridSize, y+gridSize)).(*ebiten.Image), op)
	}
}

func (p *tileParts) switchLetters() []*switchLetter {
	return p.letters
}

func (s *gameScene) Draw(screen *ebiten.Image) {
	screen.Fill(backgroundColor)
	tileParts := newTileParts(s)
	tileParts.draw(screen, s.tilesImage)
	s.drawCursor(screen)
	for _, l := range tileParts.switchLetters() {
		font.ArcadeFont.DrawText(screen, string(l.letter), l.x, l.y, 1, l.color)
	}
	s.drawPlayer(screen)
	s.drawFloorNumber(screen)
	if s.goal {
		s.drawGoalMessage(screen)
	}
}

func (s *gameScene) drawCursor(screen *ebiten.Image) {
	ox, oy := s.tileOffset()
	x0, y0, _, _ := s.tileRangeInScreen()
	dstX := s.selectedTileX*gridSize - x0*gridSize + ox
	dstY := s.selectedTileY*gridSize - y0*gridSize + oy
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(dstX), float64(dstY))
	screen.DrawImage(s.tilesImage.SubImage(image.Rect(16, 16, 16+gridSize, 16+gridSize)).(*ebiten.Image), op)
}

func (s *gameScene) drawPlayer(screen *ebiten.Image) {
	dstX := (screenWidth - gridSize) / 2
	dstY := (screenHeight - gridSize) / 2
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(dstX), float64(dstY))
	screen.DrawImage(s.tilesImage.SubImage(image.Rect(0, 16, 0+gridSize, 16+gridSize)).(*ebiten.Image), op)
}

func (s *gameScene) drawFloorNumber(screen *ebiten.Image) {
	z := s.player.z
	msg := ""
	if z == 0 {
		msg = "GROUND"
	} else {
		msg = fmt.Sprintf("B%dF", z)
	}
	x := 8
	y := 8
	font.ArcadeFont.DrawTextWithShadow(screen, msg, x, y, 1, color.White)
}

var emptyImage *ebiten.Image

func (s *gameScene) drawGoalMessage(screen *ebiten.Image) {
	if emptyImage == nil {
		emptyImage = ebiten.NewImage(screenWidth, screenHeight)
		emptyImage.Fill(color.RGBA{0, 0, 0, 0x80})
	}
	screen.DrawImage(emptyImage, nil)
	msg := "GOAL!"
	w := font.ArcadeFont.TextWidth(msg)
	x := (screenWidth - w*2) / 2
	y := 64
	font.ArcadeFont.DrawTextWithShadow(screen, msg, x, y, 2, color.White)
}
