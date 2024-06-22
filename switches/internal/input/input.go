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

package input

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Input struct {
	mouseState int
}

func New() *Input {
	return &Input{}
}

func (i *Input) Update() {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		i.mouseState++
	} else {
		i.mouseState = 0
	}
}

func (i *Input) IsTriggered() bool {
	return i.mouseState == 1
}
