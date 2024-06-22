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
	"errors"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/hajimehoshi/switches/switches/internal/input"
)

type scene interface {
	Update() error
	Draw(screen *ebiten.Image)
}

type task func() error

var (
	taskTerminated = errors.New("switches: task terminated")
)

var (
	backgroundColor = color.RGBA{0x21, 0x21, 0x21, 0xff}
)

type Game struct {
	scene scene
	tasks []task
	input *input.Input
}

func NewGame() (*Game, error) {
	g := &Game{
		input: input.New(),
	}
	g.scene = newTitleScene(g)
	return g, nil
}

const (
	screenWidth  = 256
	screenHeight = 256
)

func (g *Game) appendTask(task task) {
	g.tasks = append(g.tasks, task)
}

func (g *Game) consumeTask() (bool, error) {
	if len(g.tasks) == 0 {
		return false, nil
	}
	t := g.tasks[0]
	if err := t(); err == taskTerminated {
		g.tasks = g.tasks[1:]
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (g *Game) goTo(scene scene) {
	g.scene = scene
}

func (g *Game) Run() error {
	ebiten.SetWindowTitle("Switches")
	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	if err := ebiten.RunGame(g); err != nil {
		panic(err)
	}
	return nil
}

func (g *Game) Update() error {
	g.input.Update()
	if consumed, err := g.consumeTask(); err != nil {
		return err
	} else if !consumed {
		if err := g.scene.Update(); err != nil {
			return err
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.scene.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
