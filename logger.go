// Copyright 2013 Martini Authors
// Copyright 2014 Unknwon
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package bigo

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"time"

	"github.com/fym201/bigo/utl"
)

var ColorLog = true

func init() {
	ColorLog = runtime.GOOS != "windows"
}

// Logger returns a middleware handler that logs the request as it goes in and the response as it goes out.
func ReqLogger() Handler {
	return func(ctx *Context, log *Logger) {
		start := time.Now()

		log.LogInfo("Started %s %s for %s", ctx.Req.Method, ctx.Req.RequestURI, ctx.RemoteAddr())

		rw := ctx.Resp.(ResponseWriter)
		ctx.Next()

		content := fmt.Sprintf("Completed %s %v %s in %v", ctx.Req.RequestURI, rw.Status(), http.StatusText(rw.Status()), time.Since(start))
		if ColorLog {
			switch rw.Status() {
			case 200, 201, 202:
				content = fmt.Sprintf("\033[1;32m[INFO] %s\033[0m", content)
			case 301, 302:
				content = fmt.Sprintf("\033[1;37m[INFO] %s\033[0m", content)
			case 304:
				content = fmt.Sprintf("\033[1;33m[INFO] %s\033[0m", content)
			case 401, 403:
				content = fmt.Sprintf("\033[4;31m[INFO] %s\033[0m", content)
			case 404:
				content = fmt.Sprintf("\033[1;31m[INFO] %s\033[0m", content)
			case 500:
				content = fmt.Sprintf("\033[1;36m[INFO] %s\033[0m", content)
			}
		} else {
			content = fmt.Sprintf("[INFO] %s", content)
		}
		log.Println(content)
	}
}

var _defaultLoggerWriter io.Writer
var _defaultLogger *Logger

type LoggerWriter struct {
	io.Writer
	dir           string
	file          *os.File
	logInterval   time.Duration
	fileBeginTime time.Time
	logChan       chan []byte
}

func NewFileLoggerWriter(dir string, logInterval time.Duration) io.Writer {

	if !utl.IsExist(dir) {
		if err := os.Mkdir(dir, 0x644); err != nil {
			panic(errors.New(fmt.Sprintf("Create log dir error:[%s]", err.Error())))
		}
	}
	if dir[len(dir)-1] != '/' {
		dir += "/"
	}

	now := time.Now()
	file, err := os.OpenFile(dir+now.Format("2006-01-02-15#04#05")+".txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0x644)
	if err != nil {
		panic(errors.New(fmt.Sprintf("Create log file error:[%s]", err.Error())))
	}
	logger := LoggerWriter{dir: dir, file: file, logInterval: logInterval, fileBeginTime: now, logChan: make(chan []byte)}
	return &logger
}

func DefaultLoggerWriter() io.Writer {
	if _defaultLoggerWriter == nil {
		if GetConfig().LogDir != "" {
			_defaultLoggerWriter = NewFileLoggerWriter(GetConfig().LogDir, time.Hour*24)
		} else {
			_defaultLoggerWriter = os.Stdout
		}
	}

	return _defaultLoggerWriter
}

func (log *LoggerWriter) Write(p []byte) (n int, err error) {
	if time.Since(log.fileBeginTime) >= log.logInterval {
		if log.file != nil {
			log.file.Close()
		}
		now := time.Now()
		log.file, err = os.OpenFile(log.dir+now.Format("2006-01-02-15#04#05")+".txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0x644)
		if err != nil {
			return
		}
	}
	return log.file.Write(p)

}

type Logger struct {
	*log.Logger
}

func NewLogger(out io.Writer, prefix string, flag int) *Logger {
	logger := Logger{log.New(out, prefix, flag)}
	return &logger
}

func DefaultLogger() *Logger {
	if _defaultLogger == nil {
		_defaultLogger = NewLogger(DefaultLoggerWriter(), fmt.Sprintf("[%s] ", GetConfig().AppName), 0)
	}
	return _defaultLogger
}

func (l *Logger) Log(flag int, a ...interface{}) {
	alen := len(a)
	if alen == 0 || (l == _defaultLogger && flag < GetConfig().LogLevel) {
		return
	}

	var format string
	args := a
	if reflect.TypeOf(a[0]).Kind() == reflect.String {
		format = a[0].(string)
		args = a[1:]
	}
	content := fmt.Sprintf(format, args...)
	if ColorLog {
		switch flag {
		case LogLevelInfo:
			content = fmt.Sprintf("\033[1;32m[INFO] %s\033[0m", content)
		case LogLevelDebug:
			content = fmt.Sprintf("\033[1;37m[DEBUG] %s\033[0m", content)
		case LogLevelError:
			content = fmt.Sprintf("\033[1;31m[ERROR] %s\033[0m", content)
		}
	} else {
		switch flag {
		case LogLevelInfo:
			content = fmt.Sprintf("[INFO] %s", content)
		case LogLevelDebug:
			content = fmt.Sprintf("[DEBUG] %s", content)
		case LogLevelError:
			content = fmt.Sprintf("[ERROR] %s", content)
		}
	}

	l.Println(content)
}

func (l *Logger) LogInfo(a ...interface{}) {
	l.Log(LogLevelInfo, a...)
}

func (l *Logger) LogDebug(a ...interface{}) {
	l.Log(LogLevelDebug, a...)
}

func (l *Logger) LogError(a ...interface{}) {
	l.Log(LogLevelError, a...)
}
