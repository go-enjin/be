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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

	bePath "github.com/go-enjin/be/pkg/path"
)

var CustomExeIndent string

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

type CmdFn = func(argv ...string) (stdout string, stderr string, err error)

func Cmd(name string, argv ...string) (stdout, stderr string, status int, err error) {
	cmd := exec.Command(name, argv...)
	cmd.Env = os.Environ()

	var ob, eb bytes.Buffer
	cmd.Stdout = &ob
	cmd.Stderr = &eb

	err = cmd.Run()

	stdout = ob.String()
	stderr = eb.String()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			status = exitError.ExitCode()
		}
	}
	return
}

func CheckCmd(name string, argv ...string) (stdout string, stderr string, err error) {
	var status int
	exeBin := bePath.Which(name)
	if exeBin == "" || !bePath.IsFile(exeBin) {
		err = fmt.Errorf("%v not found", exeBin)
	} else if stdout, stderr, status, err = Cmd(exeBin, argv...); err == nil && status != 0 {
		err = fmt.Errorf("%v exited with status: %d", exeBin, status)
	}
	return
}

func MakeCmdFn(exeName string) CmdFn {
	return func(argv ...string) (stdout string, stderr string, err error) {
		var status int
		exeBin := bePath.Which(exeName)
		if exeBin == "" || !bePath.IsFile(exeBin) {
			err = fmt.Errorf("%v not found", exeBin)
		} else if stdout, stderr, status, err = Cmd(exeBin, argv...); err == nil && status != 0 {
			err = fmt.Errorf("%v exited with status: %d", exeBin, status)
		}
		return
	}
}

type ExeFn = func(argv ...string) (err error)

func Exe(name string, argv ...string) (status int, err error) {
	status, err = ExeWithChdir("", name, argv...)
	return
}

func ExeWithChdir(path, name string, argv ...string) (status int, err error) {
	cmd := exec.Command(name, argv...)
	cmd.Dir = path
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	var o, e io.ReadCloser
	if o, err = cmd.StdoutPipe(); err != nil {
		return
	}
	if e, err = cmd.StderrPipe(); err != nil {
		return
	}

	if err = cmd.Start(); err != nil {
		err = fmt.Errorf("exe start error: %v", err)
		return
	}

	so := bufio.NewScanner(o)
	se := bufio.NewScanner(e)

	go func() {
		for so.Scan() {
			_, _ = os.Stdout.WriteString(CustomExeIndent + so.Text() + "\n")
		}
	}()

	go func() {
		for se.Scan() {
			_, _ = os.Stderr.WriteString(CustomExeIndent + se.Text() + "\n")
		}
	}()

	if err = cmd.Wait(); err != nil {
		err = fmt.Errorf("exe wait error: %v", err)
	}
	return
}

func ExeWithChdirAndLog(path, name string, argv []string, stdout, stderr string, environ []string) (status int, err error) {
	cmd := exec.Command(name, argv...)
	cmd.Dir = path
	cmd.Stdin = nil
	cmd.Env = environ

	var o, e io.ReadCloser
	if o, err = cmd.StdoutPipe(); err != nil {
		return
	}
	if e, err = cmd.StderrPipe(); err != nil {
		return
	}

	var outfh, errfh *os.File
	if outfh, err = os.OpenFile(stdout, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660); err != nil {
		return
	}
	defer outfh.Close()
	if errfh, err = os.OpenFile(stderr, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660); err != nil {
		return
	}
	defer errfh.Close()

	if err = cmd.Start(); err != nil {
		err = fmt.Errorf("exe start error: %v", err)
		return
	}

	so := bufio.NewScanner(o)
	se := bufio.NewScanner(e)

	go func() {
		for so.Scan() {
			_, _ = outfh.WriteString(so.Text() + "\n")
		}
	}()

	go func() {
		for se.Scan() {
			_, _ = errfh.WriteString(se.Text() + "\n")
		}
	}()

	if err = cmd.Wait(); err != nil {
		err = fmt.Errorf("exe wait error: %v", err)
	}
	return
}

func Daemonize(path, name string, argv []string, stdout, stderr string, environ []string) (pid int, err error) {
	cmd := exec.Command(name, argv...)
	cmd.Dir = path
	cmd.Stdin = nil
	cmd.Env = environ

	var o, e io.ReadCloser
	if o, err = cmd.StdoutPipe(); err != nil {
		return
	}
	if e, err = cmd.StderrPipe(); err != nil {
		return
	}

	var outfh, errfh *os.File
	if outfh, err = os.OpenFile(stdout, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660); err != nil {
		return
	}
	if errfh, err = os.OpenFile(stderr, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660); err != nil {
		return
	}

	if err = cmd.Start(); err != nil {
		_ = outfh.Close()
		_ = errfh.Close()
		err = fmt.Errorf("exe start error: %v", err)
		return
	}

	so := bufio.NewScanner(o)
	se := bufio.NewScanner(e)

	go func() {
		for so.Scan() {
			_, _ = outfh.WriteString(so.Text() + "\n")
		}
	}()

	go func() {
		for se.Scan() {
			_, _ = errfh.WriteString(se.Text() + "\n")
		}
	}()

	pid = cmd.Process.Pid

	go func() {
		_ = cmd.Wait()
		_ = outfh.Close()
		_ = errfh.Close()
	}()
	return
}

func CheckExe(name string, argv ...string) (err error) {
	var status int
	exeBin := bePath.Which(name)
	if exeBin == "" || !bePath.IsFile(exeBin) {
		err = fmt.Errorf("%v not found", exeBin)
	} else if status, err = Exe(exeBin, argv...); err == nil && status != 0 {
		err = fmt.Errorf("%v exited with status: %d", exeBin, status)
	}
	return
}

func MakeExeFn(exeName string) ExeFn {
	return func(argv ...string) (err error) {
		var status int
		exeBin := bePath.Which(exeName)
		if exeBin == "" || !bePath.IsFile(exeBin) {
			err = fmt.Errorf("%v not found", exeBin)
		} else if status, err = Exe(exeBin, argv...); err == nil && status != 0 {
			err = fmt.Errorf("%v exited with status: %d", exeBin, status)
		}
		return
	}
}