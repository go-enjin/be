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
	"strconv"

	"github.com/go-enjin/be/pkg/log/filelogwriter"
	clPath "github.com/go-corelibs/path"
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

func Cmd(name string, argv ...string) (stdout, stderr string, status int, err error) {
	stdout, stderr, status, err = CmdWith(&Options{Name: name, Argv: argv})
	return
}

func CmdWith(options *Options) (stdout, stderr string, status int, err error) {
	cmd := exec.Command(options.Name, options.Argv...)
	cmd.Stdin = nil
	cmd.Dir = options.Path
	cmd.Env = options.Environ

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
	exeBin := clPath.Which(name)
	if exeBin == "" || !clPath.IsFile(exeBin) {
		err = fmt.Errorf("%v not found", exeBin)
	} else if stdout, stderr, status, err = Cmd(exeBin, argv...); err == nil && status != 0 {
		err = fmt.Errorf("%v exited with status: %d", exeBin, status)
	}
	return
}

func CheckExe(name string, argv ...string) (err error) {
	var status int
	exeBin := clPath.Which(name)
	if exeBin == "" || !clPath.IsFile(exeBin) {
		err = fmt.Errorf("%v not found", exeBin)
	} else if status, err = Exe(exeBin, argv...); err == nil && status != 0 {
		err = fmt.Errorf("%v exited with status: %d", exeBin, status)
	}
	return
}

func Exe(name string, argv ...string) (status int, err error) {
	if err = ExeWith(&Options{Name: name, Argv: argv}); err != nil {
		status = 1
	}
	return
}

func ExeWith(options *Options) (err error) {
	cmd := exec.Command(options.Name, options.Argv...)
	cmd.Stdin = nil
	cmd.Dir = options.Path
	cmd.Env = options.Environ

	var o, e io.ReadCloser
	var outFH, errFH *filelogwriter.FileLogWriter

	if o, err = cmd.StdoutPipe(); err != nil {
		return
	}
	if options.Stdout != "" {
		if outFH, err = filelogwriter.NewFileLogWriter(options.Stdout); err != nil {
			return
		}
	}

	if e, err = cmd.StderrPipe(); err != nil {
		return
	}
	if options.Stderr != "" {
		if errFH, err = filelogwriter.NewFileLogWriter(options.Stderr); err != nil {
			return
		}
	}

	if err = cmd.Start(); err != nil {
		err = fmt.Errorf("CMD - %v", err)
		return
	}

	if options.PidFile != "" {
		if ee := os.WriteFile(options.PidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0644); ee != nil {
			_, _ = errFH.WriteString(fmt.Sprintf("error writing pid file: %v - %v\n", options.PidFile, ee))
		}
	}

	go func() {
		so := bufio.NewScanner(o)
		for so.Scan() {
			if options.Stdout != "" {
				_, _ = outFH.WriteString(so.Text() + "\n")
			} else {
				_, _ = os.Stdout.WriteString(CustomExeIndent + so.Text() + "\n")
			}
		}
	}()

	go func() {
		se := bufio.NewScanner(e)
		for se.Scan() {
			if options.Stderr != "" {
				_, _ = errFH.WriteString(se.Text() + "\n")
			} else {
				_, _ = os.Stderr.WriteString(CustomExeIndent + se.Text() + "\n")
			}
		}
	}()

	if err = cmd.Wait(); err != nil {
		err = fmt.Errorf("CMD - %v", err)
	}
	return
}

func BackgroundWith(options *Options) (pid int, err error) {
	cmd := exec.Command(options.Name, options.Argv...)
	cmd.Dir = options.Path
	cmd.Stdin = nil
	cmd.Env = options.Environ

	var o, e io.ReadCloser
	var outFH, errFH *filelogwriter.FileLogWriter

	if options.Stdout != "" {
		if o, err = cmd.StdoutPipe(); err != nil {
			return
		}
		if outFH, err = filelogwriter.NewFileLogWriter(options.Stdout); err != nil {
			return
		}
	}

	if options.Stderr != "" {
		if e, err = cmd.StderrPipe(); err != nil {
			return
		}
		if errFH, err = filelogwriter.NewFileLogWriter(options.Stderr); err != nil {
			return
		}
	}

	if err = cmd.Start(); err != nil {
		err = fmt.Errorf("CMD - %v", err)
		return
	}

	if options.PidFile != "" {
		if ee := os.WriteFile(options.PidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0644); ee != nil {
			_, _ = errFH.WriteString(fmt.Sprintf("error writing pid file: %v - %v\n", options.PidFile, ee))
		}
	}

	if options.Stdout != "" {
		go func() {
			so := bufio.NewScanner(o)
			for so.Scan() {
				_, _ = outFH.WriteString(so.Text() + "\n")
			}
		}()
	}

	if options.Stderr != "" {
		go func() {
			se := bufio.NewScanner(e)
			for se.Scan() {
				_, _ = errFH.WriteString(se.Text() + "\n")
			}
		}()
	}

	pid = cmd.Process.Pid

	go func() {
		if err = cmd.Wait(); err != nil {
			_, _ = errFH.WriteString(fmt.Sprintf("CMD - %v\n", err))
		}
	}()
	return
}

func Interactive(environ []string, dir, name string, argv ...string) (err error) {
	if len(environ) == 0 {
		environ = os.Environ()
	}
	cmd := exec.Command(name, argv...)
	cmd.Env = environ
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return
}
