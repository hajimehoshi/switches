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
)

type task func() error

var (
	taskTerminated = errors.New("switches: task terminated")
	tasks          []task
)

func appendTask(task task) {
	tasks = append(tasks, task)
}

func consumeTask() (bool, error) {
	if len(tasks) == 0 {
		return false, nil
	}
	t := tasks[0]
	if err := t(); err == taskTerminated {
		tasks = tasks[1:]
	} else if err != nil {
		return false, err
	}
	return true, nil
}
