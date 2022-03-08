// Copyright (c) 2022  The Go-Enjin Authors
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

package run

import (
	"bytes"
	"os"
	"os/exec"
)

type PipeCmd struct {
	Name string
	Args []string
}

func NewPipe(name string, args ...string) PipeCmd {
	return PipeCmd{
		Name: name,
		Args: args,
	}
}

func ExePipe(inputs ...PipeCmd) (status int, err error) {
	chain := make([]*exec.Cmd, len(inputs))
	for idx, input := range inputs {
		chain[idx] = exec.Command(input.Name, input.Args...)
		chain[idx].Stderr = os.Stderr
		if idx > 0 {
			prev := chain[idx-1]
			chain[idx].Stdin, _ = prev.StdoutPipe()
		}
	}
	first := chain[0]
	last := chain[len(chain)-1]
	last.Stdout = os.Stdout
	for _, link := range chain[1:] {
		_ = link.Start()
	}
	if err = first.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			status = exitError.ExitCode()
		}
	}
	for _, link := range chain[1:] {
		_ = link.Wait()
	}
	return
}

func Cmd(name string, argv ...string) (stdout, stderr string, status int, err error) {
	cmd := exec.Command(name, argv...)
	var ob, eb bytes.Buffer
	cmd.Stdout = &ob
	cmd.Stderr = &eb
	if err = cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			status = exitError.ExitCode()
		}
		return
	}
	stdout = ob.String()
	stderr = eb.String()
	return
}

func Exe(name string, argv ...string) (status int, err error) {
	cmd := exec.Command(name, argv...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			status = exitError.ExitCode()
		}
		return
	}
	return
}