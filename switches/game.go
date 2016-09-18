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
	"github.com/hajimehoshi/ebiten"
)

type scene interface {
	Update() error
	Draw(screen *ebiten.Image) error
}

type Game struct {
	scene scene
	tasks []task
}

func NewGame() (*Game, error) {
	s, err := newGameScene()
	if err != nil {
		return nil, err
	}
	g := &Game{
		scene: s,
	}
	return g, nil
}

const (
	screenWidth  = 256
	screenHeight = 256
)

func (g *Game) Run() error {
	f := func(screen *ebiten.Image) error {
		if consumed, err := consumeTask(); err != nil {
			return err
		} else if !consumed {
			if err := g.update(); err != nil {
				return err
			}
		}
		if ebiten.IsRunningSlowly() {
			return nil
		}
		if err := g.draw(screen); err != nil {
			return err
		}
		return nil
	}
	if err := ebiten.Run(f, screenWidth, screenHeight, 2, "Switches"); err != nil {
		panic(err)
	}
	return nil
}

func (g *Game) update() error {
	if err := g.scene.Update(); err != nil {
		return err
	}
	return nil
}

func (g *Game) draw(screen *ebiten.Image) error {
	if err := g.scene.Draw(screen); err != nil {
		return err
	}
	return nil
}
