// Copyright 2014 bigo
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

// Package bigo is a high productive and modular design web framework in Go.
package bigo

import (
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/fym201/bigo/utl"

	"github.com/fym201/bigo/inject"
)

const _VERSION = "0.1.0.0122"

func Version() string {
	return _VERSION
}

// Handler can be any callable function.
// Bigo attempts to inject services into the handler's argument list,
// and panics if an argument could not be fullfilled via dependency injection.
type Handler interface{}

// validateHandler makes sure a handler is a callable function,
// and panics if it is not.
func validateHandler(h Handler) {
	if reflect.TypeOf(h).Kind() != reflect.Func {
		panic("Bigo handler must be a callable function")
	}
}

// validateHandlers makes sure handlers are callable functions,
// and panics if any of them is not.
func validateHandlers(handlers []Handler) {
	for _, h := range handlers {
		validateHandler(h)
	}
}

// Bigo represents the top level web application.
// inject.Injector methods can be invoked to map services on a global level.
type Bigo struct {
	inject.Injector
	befores  []BeforeHandler
	handlers []Handler
	action   Handler

	urlPrefix string // For suburl support.
	*Router

	logger *Logger
}

// NewWithLogger creates a bare bones Bigo instance.
// Use this method if you want to have full control over the middleware that is used.
// You can specify logger output writer with this function.
func NewWithLogger(logger *Logger) *Bigo {
	m := &Bigo{
		Injector: inject.New(),
		action:   func() {},
		Router:   NewRouter(),
		logger:   logger,
	}
	m.Router.m = m
	m.Map(m.logger)
	m.Map(defaultReturnHandler())
	m.notFound = func(resp http.ResponseWriter, req *http.Request) {
		c := m.createContext(resp, req)
		c.handlers = append(c.handlers, http.NotFound)
		c.run()
	}
	return m
}

// New creates a bare bones Bigo instance.
// Use this method if you want to have full control over the middleware that is used.
func New() *Bigo {
	return NewWithLogger(DefaultLogger())
}

// Classic creates a classic Bigo with some basic default middleware:
// mocaron.Logger, mocaron.Recovery and mocaron.Static.
func Classic() *Bigo {
	conf := GetConfig()
	var m *Bigo
	if conf.LogLevel != LogLevelNone {
		m = New()
		m.Use(ReqLogger())
	}

	m.Use(Recovery())

	if conf.EnableGzip {
		m.Use(Gziper())
	}

	for i := 0; i < len(conf.Statics); i++ {
		opt := conf.Statics[i]
		mopt := StaticOptions{opt.Prefix, opt.SkipLogging, opt.IndexFile, nil, nil}
		m.Use(Static(opt.Path, mopt))
	}

	if conf.Tmpl != nil && conf.Tmpl.Enable {
		m.Use(Renderer(RenderOptions{
			Directory:       conf.Tmpl.Directory,
			Extensions:      conf.Tmpl.Extensions,
			Delims:          Delims{conf.Tmpl.Delims[0], conf.Tmpl.Delims[1]},
			Charset:         conf.Tmpl.Charset,
			IndentJSON:      conf.Tmpl.IndentJSON,
			IndentXML:       conf.Tmpl.IndentXML,
			HTMLContentType: conf.Tmpl.HTMLContentType,
		}))
	}

	if conf.I18n != nil && conf.I18n.Enable {
		m.Use(I18n())
	}
	return m
}

// Handlers sets the entire middleware stack with the given Handlers.
// This will clear any current middleware handlers,
// and panics if any of the handlers is not a callable function
func (m *Bigo) Handlers(handlers ...Handler) {
	m.handlers = make([]Handler, 0)
	for _, handler := range handlers {
		m.Use(handler)
	}
}

// Action sets the handler that will be called after all the middleware has been invoked.
// This is set to macaron.Router in a macaron.Classic().
func (m *Bigo) Action(handler Handler) {
	validateHandler(handler)
	m.action = handler
}

// BeforeHandler represents a handler executes at beginning of every request.
// Bigo stops future process when it returns true.
type BeforeHandler func(rw http.ResponseWriter, req *http.Request) bool

func (m *Bigo) Before(handler BeforeHandler) {
	m.befores = append(m.befores, handler)
}

// Use adds a middleware Handler to the stack,
// and panics if the handler is not a callable func.
// Middleware Handlers are invoked in the order that they are added.
func (m *Bigo) Use(handler Handler) {
	validateHandler(handler)
	m.handlers = append(m.handlers, handler)
}

func (m *Bigo) createContext(rw http.ResponseWriter, req *http.Request) *Context {
	c := &Context{
		Injector: inject.New(),
		handlers: m.handlers,
		action:   m.action,
		index:    0,
		Router:   m.Router,
		Req:      Request{req},
		Resp:     NewResponseWriter(rw),
		Data:     make(map[string]interface{}),
	}
	c.SetParent(m)
	c.Map(c)
	c.MapTo(c.Resp, (*http.ResponseWriter)(nil))
	c.Map(req)
	return c
}

// ServeHTTP is the HTTP Entry point for a Bigo instance.
// Useful if you want to control your own HTTP server.
// Be aware that none of middleware will run without registering any router.
func (m *Bigo) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	req.URL.Path = strings.TrimPrefix(req.URL.Path, m.urlPrefix)
	for _, h := range m.befores {
		if h(rw, req) {
			return
		}
	}
	m.Router.ServeHTTP(rw, req)
}

func (m *Bigo) Run() {
	conf := GetConfig()
	if conf.EnableHttps {
		if conf.ForceHttps {
			go m.RunHttp(conf.HttpAddr + ":" + utl.ToStr(conf.HttpPort))
		}

		m.RunHttps(conf.HttpsAddr+":"+utl.ToStr(conf.HttpsPort), conf.HttpsCertFile, conf.HttpsKeyFile)
	} else {
		m.RunHttp(conf.HttpAddr + ":" + utl.ToStr(conf.HttpPort))
	}
}

func (m *Bigo) RunHttp(addr string) {

	logger := m.Injector.GetVal(reflect.TypeOf(m.logger)).Interface().(*Logger)
	logger.LogInfo("Http listening on %s (%s)\n", addr, Env)
	if err := http.ListenAndServe(addr, m); err != nil {
		panic(err)
	}
}

func (m *Bigo) RunHttps(addr string, cerFile string, keyFile string) {
	logger := m.Injector.GetVal(reflect.TypeOf(m.logger)).Interface().(*Logger)
	logger.LogInfo("Https listening on %s (%s)\n", addr, Env)

	m.Use(Secure(SecureOptions{
		SSLRedirect: true,
		SSLHost:     addr, // This is optional in production. The default behavior is to just redirect the request to the https protocol. Example: http://github.com/some_page would be redirected to https://github.com/some_page.
	}))
	if err := http.ListenAndServeTLS(addr, cerFile, keyFile, m); err != nil {
		panic(err)
	}

}

// SetURLPrefix sets URL prefix of router layer, so that it support suburl.
func (m *Bigo) SetURLPrefix(prefix string) {
	m.urlPrefix = prefix
}

const (
	Dev  = "development"
	Prod = "production"
	Test = "test"
)

var (
	// Env is the environment that Bigo is executing in.
	Env = Dev

	// Path of work directory.
	Root string

	// Flash applies to current request.
	FlashNow bool

	// Configuration convention object.

)

func SetEnv(e string) {
	if len(e) > 0 {
		Env = e
	}
}

func init() {

	var err error
	Root, err = os.Getwd()
	if err != nil {
		panic("error getting work directory: " + err.Error())
	}
}
