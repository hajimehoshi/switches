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

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/switches/switches/internal/font"
)

type titleScene struct {
	game        *Game
	gameScene   *gameScene
	gameSceneCh chan error
}

func newTitleScene(game *Game) *titleScene {
	return &titleScene{
		game: game,
	}
}

func (t *titleScene) Update() error {
	if t.gameSceneCh == nil {
		t.gameSceneCh = make(chan error)
		go func() {
			defer close(t.gameSceneCh)
			s, err := newGameScene(t.game)
			if err != nil {
				t.gameSceneCh <- err
				return
			}
			t.gameScene = s
		}()
		return nil
	}
	select {
	case err := <-t.gameSceneCh:
		if err != nil {
			return err
		}
		t.game.goTo(t.gameScene)
	default:
	}
	return nil
}

func (t *titleScene) Draw(screen *ebiten.Image) error {
	if err := font.ArcadeFont.DrawText(screen, "NOW LOADING...", 8, 8, 1, color.White); err != nil {
		return err
	}
	return nil
}
