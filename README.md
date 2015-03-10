Bigo is a web framework base on martini and macaron for quick start web develop

Copy the `config.json` to your workdir and create main.go:
```go
package main

import (
	"github.com/fym201/bigo"
)

func main() {
	m := bigo.Classic()
	m.Get("/", func() string {
		return "Hello world!"
	})
	m.Run()
}
```