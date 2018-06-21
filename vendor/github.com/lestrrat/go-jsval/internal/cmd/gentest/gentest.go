package main

import (
	"bytes"
	"io"
	"log"
	"os"

	schema "github.com/lestrrat/go-jsschema"
	jsval "github.com/lestrrat/go-jsval"
	"github.com/lestrrat/go-jsval/builder"
)

func main() {
	os.Exit(_main())
}

func _main() int {
	s, err := schema.ReadFile(os.Args[1])
	if err != nil {
		log.Printf("%s", err)
		return 1
	}

	b := builder.New()
	v, err := b.Build(s)
	if err != nil {
		log.Printf("%s", err)
		return 1
	}

	var out bytes.Buffer
	out.WriteString("package jsval_test")
	out.WriteString("\n\nimport \"github.com/lestrrat/go-jsval\"")
	out.WriteString("\n")

	g := jsval.NewGenerator()
	if err := g.Process(&out, v); err != nil {
		log.Printf("%s", err)
		return 1
	}

	f, err := os.OpenFile(os.Args[2], os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Printf("%s", err)
		return 1
	}
	defer f.Close()

	io.Copy(f, &out)
	return 0
}
