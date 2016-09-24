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

func calcPath(passable func(x, y int) bool, startX, startY, goalX, goalY int) []dir {
	type pos struct {
		X, Y int
	}
	current := []pos{{startX, startY}}
	parents := map[pos]pos{}
	for 0 < len(current) {
		next := []pos{}
		for _, p := range current {
			successors := []pos{
				{p.X + 1, p.Y},
				{p.X - 1, p.Y},
				{p.X, p.Y + 1},
				{p.X, p.Y - 1},
			}
			for _, s := range successors {
				if !passable(s.X, s.Y) {
					continue
				}
				if _, ok := parents[s]; ok {
					continue
				}
				parents[s] = p
				if s.X == goalX && s.Y == goalY {
					break
				}
				next = append(next, s)
			}
		}
		current = next
	}
	p := pos{goalX, goalY}
	dirs := []dir{}
	for p.X != startX || p.Y != startY {
		parent, ok := parents[p]
		// There is no path.
		if !ok {
			return nil
		}
		switch {
		case parent.X == p.X - 1:
			dirs = append(dirs, dirRight)
		case parent.X == p.X + 1:
			dirs = append(dirs, dirLeft)
		case parent.Y == p.Y - 1:
			dirs = append(dirs, dirDown)
		case parent.Y == p.Y + 1:
			dirs = append(dirs, dirUp)
		default:
			panic("not reach")
		}
		p = parent
	}
	path := make([]dir, len(dirs))
	for i, d := range dirs {
		path[len(dirs) - i - 1] = d
	}
	return path
}
