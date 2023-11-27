// Copyright (c) 2023  The Go-Enjin Authors
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

package filelogwriter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var _ io.Writer = (*FileLogWriter)(nil)

// FileLogWriter is an io.Writer that does not keep any file handles open any
// longer than necessary.
type FileLogWriter struct {
	file string
	flag int
	mode os.FileMode
}

// NewFileLogWriter constructs a new FileLogWriter instance with the settings
// given.
//
// `file` is the local filesystem output destination
// `flag` is the os.Flags setting, default is os.O_CREATE|os.O_WRONLY|os.O_APPEND
// `mode` is the file mode setting, default is 0644
func NewFileLogWriter(file string) (flw *FileLogWriter, err error) {
	if file == "" {
		err = fmt.Errorf("NewFileLogWriter() requires the file argument to not be empty")
		return
	}
	if file, err = filepath.Abs(file); err != nil {
		return
	}
	flw = &FileLogWriter{
		file: file,
		flag: os.O_CREATE | os.O_WRONLY | os.O_APPEND,
		mode: 0644,
	}
	return
}

// SetFlag is a chainable method for setting the file flags used to open a new
// file handle each time Write is called
func (flw *FileLogWriter) SetFlag(flag int) *FileLogWriter {
	flw.flag = flag
	return flw
}

// SetMode is a chainable method for setting the file mode used to open a new
// file handle each time Write is called
func (flw *FileLogWriter) SetMode(mode os.FileMode) *FileLogWriter {
	flw.mode = mode
	return flw
}

// Write opens the log file, writes the data given and returns the bytes written
// and the error state after closing the open file handle.
func (flw *FileLogWriter) Write(p []byte) (n int, err error) {
	var fh *os.File
	if fh, err = os.OpenFile(flw.file, flw.flag, flw.mode); err != nil {
		return
	}
	defer func() {
		_ = fh.Close()
	}()
	n, err = fh.Write(p)
	return
}

// WriteString is a convenience wrapper around Write()
func (flw *FileLogWriter) WriteString(s string) (n int, err error) {
	n, err = flw.Write([]byte(s))
	return
}
