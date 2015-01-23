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
	"runtime"
	"time"

	"github.com/fym201/bigo/utl"
)

var ColorLog = true

func init() {
	ColorLog = runtime.GOOS != "windows"
}

// Logger returns a middleware handler that logs the request as it goes in and the response as it goes out.
func Logger() Handler {
	return func(ctx *Context, log *log.Logger) {
		start := time.Now()

		log.Printf("Started %s %s for %s", ctx.Req.Method, ctx.Req.RequestURI, ctx.RemoteAddr())

		rw := ctx.Resp.(ResponseWriter)
		ctx.Next()

		content := fmt.Sprintf("Completed %s %v %s in %v", ctx.Req.RequestURI, rw.Status(), http.StatusText(rw.Status()), time.Since(start))
		if ColorLog {
			switch rw.Status() {
			case 200, 201, 202:
				content = fmt.Sprintf("\033[1;32m%s\033[0m", content)
			case 301, 302:
				content = fmt.Sprintf("\033[1;37m%s\033[0m", content)
			case 304:
				content = fmt.Sprintf("\033[1;33m%s\033[0m", content)
			case 401, 403:
				content = fmt.Sprintf("\033[4;31m%s\033[0m", content)
			case 404:
				content = fmt.Sprintf("\033[1;31m%s\033[0m", content)
			case 500:
				content = fmt.Sprintf("\033[1;36m%s\033[0m", content)
			}
		}
		log.Println(content)
	}
}

type LoggerWriter struct {
	io.Writer
	dir           string
	file          *os.File
	logInterval   time.Duration
	fileBeginTime time.Time
	logChan       chan []byte
}

func NewLoggerWriter(dir string, logInterval time.Duration) *LoggerWriter {

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
