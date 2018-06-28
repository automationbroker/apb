package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/lestrrat/go-jspointer"
	"github.com/lestrrat/go-jsschema"
	"github.com/lestrrat/go-jsval"
	"github.com/lestrrat/go-jsval/builder"
	"github.com/lestrrat/go-jsval/server"
)

func main() {
	os.Exit(_main())
}

// jsval schema.json [ref]
// jsval hyper-schema.json -ptr /path/to/schema1 -ptr /path/to/schema2 -ptr /path/to/schema3
// jsval server -listen :8080

func _main() int {
	if len(os.Args) > 1 && os.Args[1] == "server" {
		return _server()
	}

	return _cli()
}

type serverOptions struct {
	Listen string  `short:"l" long:"listen" description:"the address to listen to" default:":8080"`
}

func _server() int {
	var opts serverOptions
	if _, err := flags.Parse(&opts); err != nil {
		log.Printf("%s", err)
		return 1
	}

	s := server.New()
	if err := s.Run(opts.Listen); err != nil {
		return 1
	}

	return 0
}

type cliOptions struct {
	Schema  string   `short:"s" long:"schema" description:"the source JSON schema file"`
	OutFile string   `short:"o" long:"outfile" description:"output file to generate"`
	Pointer []string `short:"p" long:"ptr" description:"JSON pointer(s) within the document to create validators with"`
	Prefix  string   `short:"P" long:"prefix" description:"prefix for validator name(s)"`
}

func _cli() int {
	var opts cliOptions
	if _, err := flags.Parse(&opts); err != nil {
		log.Printf("%s", err)
		return 1
	}

	f, err := os.Open(opts.Schema)
	if err != nil {
		log.Printf("%s", err)
		return 1
	}
	defer f.Close()

	var m map[string]interface{}
	if err := json.NewDecoder(f).Decode(&m); err != nil {
		log.Printf("%s", err)
		return 1
	}

	// Extract possibly multiple schemas out of the main JSON document.
	var schemas []*schema.Schema
	ptrs := opts.Pointer
	if len(ptrs) == 0 {
		s, err := schema.ReadFile(opts.Schema)
		if err != nil {
			log.Printf("%s", err)
			return 1
		}
		schemas = []*schema.Schema{s}
	} else {
		for _, ptr := range ptrs {
			log.Printf("Resolving pointer '%s'", ptr)
			resolver, err := jspointer.New(ptr)
			if err != nil {
				log.Printf("%s", err)
				return 1
			}

			resolved, err := resolver.Get(m)
			if err != nil {
				log.Printf("%s", err)
				return 1
			}

			m2, ok := resolved.(map[string]interface{})
			if !ok {
				log.Printf("Expected map")
				return 1
			}

			s := schema.New()
			if err := s.Extract(m2); err != nil {
				log.Printf("%s", err)
				return 1
			}
			schemas = append(schemas, s)
		}
	}

	b := builder.New()

	validators := make([]*jsval.JSVal, len(schemas))
	for i, s := range schemas {
		v, err := b.BuildWithCtx(s, m)
		if err != nil {
			log.Printf("%s", err)
			return 1
		}
		if p := opts.Prefix; p != "" {
			v.Name = fmt.Sprintf("%s%d", p, i)
		}
		validators[i] = v
	}

	var out io.Writer

	out = os.Stdout
	if fn := opts.OutFile; fn != "" {
		f, err := os.Create(fn)
		if err != nil {
			log.Printf("%s", err)
			return 1
		}
		defer f.Close()

		out = f
	}

	g := jsval.NewGenerator()
	if err := g.Process(out, validators...); err != nil {
		log.Printf("%s", err)
		return 1
	}

	return 0
}