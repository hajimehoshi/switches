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
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/hajimehoshi/switches/switches/internal/font"
)

type mode struct {
	text      string
	fieldSize int
	x         int
	y         int
}

func (m *mode) size() (int, int) {
	return font.ArcadeFont.TextWidth(m.text), font.ArcadeFont.TextHeight(m.text)
}

type titleScene struct {
	game         *Game
	gameScene    *gameScene
	modes        []*mode
	selectedMode *mode
	loadingCh    chan error
}

func newTitleScene(game *Game) *titleScene {
	y := screenHeight - 32 - 64
	modes := []*mode{
		{"EASY", 2, 0, y},
		{"NORMAL", 4, 0, y + 16},
		{"HARD", 6, 0, y + 32},
		{"EXTREME", 8, 0, y + 48},
	}
	maxWidth := 0
	for _, m := range modes {
		w, _ := m.size()
		if w > maxWidth {
			maxWidth = w
		}
	}
	for _, m := range modes {
		m.x = (screenWidth - maxWidth) / 2
	}
	return &titleScene{
		game:  game,
		modes: modes,
	}
}

func (t *titleScene) Update() error {
	if t.loadingCh == nil {
		t.selectedMode = nil
		x, y := ebiten.CursorPosition()
		for _, m := range t.modes {
			w, h := m.size()
			if m.x <= x && x < m.x+w && m.y <= y && y < m.y+h {
				t.selectedMode = m
				break
			}
		}
	}
	if t.game.input.IsTriggered() && t.selectedMode != nil && t.loadingCh == nil {
		t.loadingCh = make(chan error)
		m := t.selectedMode
		go func() {
			defer close(t.loadingCh)
			s, err := newGameScene(m.fieldSize, m.fieldSize, m.fieldSize, m.fieldSize, t.game)
			if err != nil {
				t.loadingCh <- err
				return
			}
			t.gameScene = s
		}()
		return nil
	}
	select {
	case err := <-t.loadingCh:
		if err != nil {
			return err
		}
		t.game.goTo(t.gameScene)
	default:
	}
	return nil
}

func (t *titleScene) Draw(screen *ebiten.Image) {
	screen.Fill(backgroundColor)
	if t.loadingCh == nil {
		title := "SWITCHES"
		w := font.ArcadeFont.TextWidth(title)
		x := (screenWidth - w*2) / 2
		font.ArcadeFont.DrawText(screen, title, x, 64, 2, color.White)
		for _, m := range t.modes {
			clr := color.Color(color.White)
			if t.selectedMode == m {
				clr = color.RGBA{0xff, 0xee, 0x58, 0xff}
			}
			font.ArcadeFont.DrawText(screen, m.text, m.x, m.y, 1, clr)
		}
		return
	}
	font.ArcadeFont.DrawText(screen, "NOW LOADING...", 8, 8, 1, color.White)
}
