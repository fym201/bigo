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

package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	. "github.com/fym201/bigo"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_Version(t *testing.T) {
	Convey("Get version", t, func() {
		So(Version(), ShouldEqual, "0.1.0.0122")
	})
}

func Test_New(t *testing.T) {
	Convey("Initialize a new instance", t, func() {
		So(New(), ShouldNotBeNil)
	})

	Convey("Just test that Run doesn't bomb", t, func() {
		go New().Run()
		time.Sleep(1 * time.Second)
		os.Setenv("PORT", "4001")
		go New().Run("0.0.0.0")
		go New().Run(4002)
		go New().Run("0.0.0.0", 4003)
	})
}

func Test_Bigo_Before(t *testing.T) {
	Convey("Register before handlers", t, func() {
		m := New()
		m.Before(func(rw http.ResponseWriter, req *http.Request) bool {
			return false
		})
		m.Before(func(rw http.ResponseWriter, req *http.Request) bool {
			return true
		})
		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/", nil)
		So(err, ShouldBeNil)
		m.ServeHTTP(resp, req)
	})
}

func Test_Bigo_ServeHTTP(t *testing.T) {
	Convey("Serve HTTP requests", t, func() {
		result := ""
		m := New()
		m.Use(func(c *Context) {
			result += "foo"
			c.Next()
			result += "ban"
		})
		m.Use(func(c *Context) {
			result += "bar"
			c.Next()
			result += "baz"
		})
		m.Get("/", func() {})
		m.Action(func(res http.ResponseWriter, req *http.Request) {
			result += "bat"
			res.WriteHeader(http.StatusBadRequest)
		})

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/", nil)
		So(err, ShouldBeNil)
		m.ServeHTTP(resp, req)
		So(result, ShouldEqual, "foobarbatbazban")
		So(resp.Code, ShouldEqual, http.StatusBadRequest)
	})
}

func Test_Bigo_Handlers(t *testing.T) {
	Convey("Add custom handlers", t, func() {
		result := ""
		batman := func(c *Context) {
			result += "batman!"
		}

		m := New()
		m.Use(func(c *Context) {
			result += "foo"
			c.Next()
			result += "ban"
		})
		m.Handlers(
			batman,
			batman,
			batman,
		)

		Convey("Add not callable function", func() {
			defer func() {
				So(recover(), ShouldNotBeNil)
			}()
			m.Use("shit")
		})

		m.Get("/", func() {})
		m.Action(func(res http.ResponseWriter, req *http.Request) {
			result += "bat"
			res.WriteHeader(http.StatusBadRequest)
		})

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/", nil)
		So(err, ShouldBeNil)
		m.ServeHTTP(resp, req)
		So(result, ShouldEqual, "batman!batman!batman!bat")
		So(resp.Code, ShouldEqual, http.StatusBadRequest)
	})
}

func Test_Bigo_EarlyWrite(t *testing.T) {
	Convey("Write early content to response", t, func() {
		result := ""
		m := New()
		m.Use(func(res http.ResponseWriter) {
			result += "foobar"
			res.Write([]byte("Hello world"))
		})
		m.Use(func() {
			result += "bat"
		})
		m.Get("/", func() {})
		m.Action(func(res http.ResponseWriter) {
			result += "baz"
			res.WriteHeader(http.StatusBadRequest)
		})

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/", nil)
		So(err, ShouldBeNil)
		m.ServeHTTP(resp, req)
		So(result, ShouldEqual, "foobar")
		So(resp.Code, ShouldEqual, http.StatusOK)
	})
}
