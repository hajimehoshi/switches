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
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type dir int

const (
	dirLeft dir = iota
	dirRight
	dirUp
	dirDown
	dirUpstairs
	dirDownstairs
)

func (d dir) opposite() dir {
	switch d {
	case dirRight:
		return dirLeft
	case dirLeft:
		return dirRight
	case dirDown:
		return dirUp
	case dirUp:
		return dirDown
	case dirDownstairs:
		return dirUpstairs
	case dirUpstairs:
		return dirDownstairs
	}
	panic("not reach")
}

type passageSwitchType int

const (
	passageSwitchTypeDontCare passageSwitchType = iota
	passageSwitchTypeNeedFalse
	passageSwitchTypeNeedTrue
)

type passage struct {
	switches []passageSwitchType
}

func newPassage(switches int) *passage {
	return &passage{
		switches: make([]passageSwitchType, switches),
	}
}

func (p *passage) initRandomly(switches int, switchBits int) {
	for i := 0; i < switches; i++ {
		if switches > 1 && rand.Intn(switches) == 0 {
			continue
		}
		if (switchBits>>uint(i))&1 == 0 {
			p.switches[i] = passageSwitchTypeNeedFalse
		} else {
			p.switches[i] = passageSwitchTypeNeedTrue
		}
	}
}

func (p *passage) allow(switches int, switchBits int) {
	for i := 0; i < switches; i++ {
		if (switchBits>>uint(i))&1 == 0 {
			if p.switches[i] == passageSwitchTypeNeedTrue {
				p.switches[i] = passageSwitchTypeDontCare
			}
		} else {
			if p.switches[i] == passageSwitchTypeNeedFalse {
				p.switches[i] = passageSwitchTypeDontCare
			}
		}
	}
}

func (p *passage) dontCareNum(switches int, switchBits int) int {
	n := 0
	for i := 0; i < switches; i++ {
		if p.switches[i] == passageSwitchTypeDontCare {
			n++
			continue
		}
		if (switchBits>>uint(i))&1 == 0 {
			if p.switches[i] == passageSwitchTypeNeedTrue {
				n++
			}
		} else {
			if p.switches[i] == passageSwitchTypeNeedFalse {
				n++
			}
		}
	}
	return n
}

type room struct {
	x, y, z  int
	dirs     [6]*passage
	switches []bool
}

type field struct {
	rooms    []*room
	width    int
	height   int
	depth    int
	switches int
}

func newField(width, height, depth, switches int) (*field, error) {
	f := &field{
		width:    width,
		height:   height,
		depth:    depth,
		switches: switches,
	}
	for !f.makeRoughStructure() {
	}
	return f, nil
}

func (f *field) index(x, y, z int) int {
	return x + y*f.width + z*f.width*(f.height+1)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func (f *field) newRoom(x, y, z int) *room {
	r := &room{
		x:        x,
		y:        y,
		z:        z,
		switches: make([]bool, f.switches),
	}
	return r
}

func (f *field) makeRoughStructure() bool {
	f.rooms = make([]*room, f.width*(f.height+1)*f.depth)
	type position struct {
		X, Y, Z, SwitchBits int
	}
	start := position{0, 0, 0, 0}
	goal := position{f.width - 1, f.height - 1, f.depth - 1, (1 << uint(f.switches)) - 1}
	f.rooms[f.index(start.X, start.Y, start.Z)] = f.newRoom(start.X, start.Y, start.Z)
	current := start
	continued := 0
	for current != goal {
		// TODO: Calc candidate first!
		if 10 < continued {
			return false
		}
		nx, ny, nz, ns := current.X, current.Y, current.Z, current.SwitchBits
		var d dir
		changeSwitch := f.switches > 0 && rand.Intn(4) == 0
		changedSwitch := 0
		if changeSwitch {
			changedSwitch = rand.Intn(f.switches)
			ns ^= 1 << uint(changedSwitch)
		} else {
			d = dir(rand.Intn(6))
			switch d {
			case dirRight:
				nx = min(current.X+1, f.width-1)
			case dirLeft:
				nx = max(current.X-1, 0)
			case dirDown:
				ny = min(current.Y+1, f.height-1)
			case dirUp:
				ny = max(current.Y-1, 0)
			case dirDownstairs:
				nz = min(current.Z+1, f.depth-1)
			case dirUpstairs:
				nz = max(current.Z-1, 0)
			}
			if nx == current.X && ny == current.Y && nz == current.Z {
				continue
			}
		}
		prevRoom := f.rooms[f.index(current.X, current.Y, current.Z)]
		if changeSwitch {
			n := 0
			for i, b := range prevRoom.switches {
				if b || i == changedSwitch {
					n++
				}
			}
			if max(1, f.switches/2) < n {
				continued++
				continue
			}
			prevRoom.switches[changedSwitch] = true
		} else {
			if prevRoom.dirs[d] == nil {
				p := newPassage(f.switches)
				p.initRandomly(f.switches, ns)
				prevRoom.dirs[d] = p
				nextRoom := f.rooms[f.index(nx, ny, nz)]
				if nextRoom == nil {
					nextRoom = f.newRoom(nx, ny, nz)
					f.rooms[f.index(nx, ny, nz)] = nextRoom
				}
				nextRoom.dirs[d.opposite()] = p
			} else {
				p := prevRoom.dirs[d]
				if max(0, f.switches-1) < p.dontCareNum(f.switches, ns) {
					continued++
					continue
				}
				p.allow(f.switches, ns)
			}
		}
		continued = 0
		current = position{nx, ny, nz, ns}
	}
	lastRoom := f.newRoom(f.width - 1, f.height, f.depth - 1)
	f.rooms[f.index(f.width - 1, f.height, f.depth - 1)] = lastRoom
	lastPassage := newPassage(f.switches)
	for i := 0; i < f.switches; i++ {
		lastPassage.switches[i] = passageSwitchTypeNeedTrue
	}
	f.rooms[f.index(f.width - 1, f.height - 1, f.depth - 1)].dirs[dirDown] = lastPassage
	lastRoom.dirs[dirUp] = lastPassage
	return true
}

type tile int

const (
	tileNone tile = iota
	tileRegular
	tileDownstairs
	tileUpstairs
	tileOneWayLeft
	tileOneWayRight
	tileOneWayUp
	tileOneWayDown
	tileOneWayDownstairs
	tileOneWayUpstairs
	tileSwitch0
	tileSwitch1
	tileSwitchedTileValid
	tileSwitchedTileInvalid
)

func (t tile) oneWay() bool {
	switch t {
	case tileOneWayLeft:
		return true
	case tileOneWayRight:
		return true
	case tileOneWayUp:
		return true
	case tileOneWayDown:
		return true
	case tileOneWayDownstairs:
		return true
	case tileOneWayUpstairs:
		return true
	}
	return false
}

func (t tile) isPassable() bool {
	if t == tileNone {
		return false
	}
	if t == tileSwitchedTileInvalid {
		return false
	}
	return true
}

func (f *field) start() (int, int) {
	_, h := f.roomSize()
	return 2, h - 1
}

func (f *field) roomSize() (int, int) {
	return 5 + 2*f.switches, 4 + f.switches
}

func (f *field) tileSize() (int, int, int) {
	w, h := f.roomSize()
	return f.width * w, (f.height + 1) * h, f.depth
}

func switchedTile(passageSwitchType passageSwitchType, state bool) tile {
	switch passageSwitchType {
	case passageSwitchTypeDontCare:
		return tileRegular
	case passageSwitchTypeNeedFalse:
		if state {
			return tileSwitchedTileInvalid
		} else {
			return tileSwitchedTileValid
		}
	case passageSwitchTypeNeedTrue:
		if state {
			return tileSwitchedTileValid
		} else {
			return tileSwitchedTileInvalid
		}
	}
	panic("not reach")
}

func (f *field) tile(x, y, z int, switchStates []bool) (tile, int) {
	// 7x5
	//     ^^
	// ST  []  ST
	// BL  BL  BL
	// []  []SW[]
	// [][][][][]BL>>

	// 9x6
	//     ^^
	// ST  []    ST
	// BL  BL    BL
	// BL  BL    BL
	// []  []SWSW[]
	// [][][][][][]BLBL>>

	w, h := f.roomSize()
	rx, ry, rz := x/w, y/h, z
	room := f.rooms[f.index(rx, ry, rz)]
	if room == nil {
		return tileNone, 0
	}
	mx := x % w
	my := y % h
	cx, cy := 2, h-1
	if mx == cx && my == cy {
		return tileRegular, 0
	}
	hasUpstairsLeft := false
	hasDownstairsLeft := false
	hasUpstairsRight := false
	hasDownstairsRight := false
	hasSwitch := false
	for i := 0; i < f.switches; i++ {
		if room.switches[i] {
			hasSwitch = true
			break
		}
	}
	if z%2 == 0 {
		if room.dirs[dirUpstairs] != nil {
			hasUpstairsLeft = true
		}
		if room.dirs[dirDownstairs] != nil {
			hasDownstairsRight = true
		}
	}
	if z%2 == 1 {
		if room.dirs[dirDownstairs] != nil {
			hasDownstairsLeft = true
		}
		if room.dirs[dirUpstairs] != nil {
			hasUpstairsRight = true
		}
	}

	if my == cy {
		switch {
		case mx < cx:
			if hasDownstairsLeft || hasUpstairsLeft || room.dirs[dirLeft] != nil {
				return tileRegular, 0
			}
		case cx < mx && mx <= cx+f.switches+1:
			if hasDownstairsRight || hasUpstairsRight || room.dirs[dirRight] != nil || hasSwitch {
				return tileRegular, 0
			}
		case cx+f.switches+1 < mx && mx < w-1:
			p := room.dirs[dirRight]
			if p == nil {
				return tileNone, 0
			}
			i := mx - (cx + f.switches + 2)
			return switchedTile(p.switches[i], switchStates[i]), i
		case mx == w-1:
			if room.dirs[dirRight] != nil {
				return tileRegular, 0
			}
		}
		return tileNone, 0
	}
	if my == cy-1 && cx+1 <= mx && mx <= cx+f.switches {
		i := mx - (cx + 1)
		if room.switches[i] {
			tile := tileSwitch0
			if switchStates[i] {
				tile = tileSwitch1
			}
			return tile, i
		}
		return tileNone, 0
	}
	switch {
	case mx == 0:
		if my == 0 {
			return tileNone, 0
		}
		if hasUpstairsLeft {
			switch {
			case my == 1:
				return tileUpstairs, 0
			case 1 < my && my < f.switches+2:
				p := room.dirs[dirUpstairs]
				i := my - 2
				return switchedTile(p.switches[i], switchStates[i]), i
			}
			return tileRegular, 0
		}
		if hasDownstairsLeft {
			switch {
			case my == 1:
				return tileDownstairs, 0
			case 1 < my && my < f.switches+2:
				p := room.dirs[dirDownstairs]
				i := my - 2
				return switchedTile(p.switches[i], switchStates[i]), i
			}
			return tileRegular, 0
		}
	case mx == cx:
		if 1 < my && my < f.switches+2 {
			p := room.dirs[dirUp]
			if p == nil {
				return tileNone, 0
			}
			i := my - 2
			return switchedTile(p.switches[i], switchStates[i]), i
		}
		if room.dirs[dirUp] != nil {
			return tileRegular, 0
		}
	case mx == 3+f.switches:
		if my == 0 {
			return tileNone, 0
		}
		if hasDownstairsRight {
			switch {
			case my == 1:
				return tileDownstairs, 0
			case 1 < my && my < f.switches+2:
				p := room.dirs[dirDownstairs]
				i := my - 2
				return switchedTile(p.switches[i], switchStates[i]), i
			}
			return tileRegular, 0
		}
		if hasUpstairsRight {
			switch {
			case my == 1:
				return tileUpstairs, 0
			case 1 < my && my < f.switches+2:
				p := room.dirs[dirUpstairs]
				i := my - 2
				return switchedTile(p.switches[i], switchStates[i]), i
			}
			return tileRegular, 0
		}
	}
	return tileNone, 0
}
