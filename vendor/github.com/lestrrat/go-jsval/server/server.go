package server

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"sync"

	"github.com/lestrrat/go-jsschema"
	"github.com/lestrrat/go-jsval"
	"github.com/lestrrat/go-jsval/builder"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

const indexTmpl = `<!doctype html>
<html>
	<head>
		<title>JSVal Playground</title>
<style type="text/css">
<!--
	body {
		width: 800px;
	}

	#json {
		width: 800px;
		height: 10em;
	}

	#output {
		background-color: #eee;
		border: 1px solid #ccc;
		display: none;
		font-family: monospace;
		padding: 0em 2em 1em 2em;
		white-space: pre;
		width: 800px;
	}
-->
</style>
	</head>
	<body>
		<div id="banner">
			<div id="head" itemprop="name">The JSval Playground</div>
			<div id="controls">
				<input type="button" value="Run" id="run">
		</div>
		<div id="wrap">
			<textarea itemprop="description" id="json" name="json" autocorrect="off" autocomplete="off" autocapitalize="off" spellcheck="false">{
  "type": "string",
  "enum": ["foo", "bar", "baz"]
}
</textarea>
		</div>
		<div id="output"></div>
<script src="//ajax.googleapis.com/ajax/libs/jquery/2.2.4/jquery.min.js"></script>
<script type="text/javascript">
(function() {
	var updateCode = function(data, status, xhr) {
		$("#output").show();
		if (data.success) {
			$("#output").text(data["code"]);
		} else {
			$("#output").text("Failed to generate code: " + data.message);
		}
	};

	$("#run").click(function() {
		$.ajax({
			"type": "POST",
			"url": "/generate.json",
			"data": $("#json").val(),
			"success": updateCode,
			"dataType": "json"
		})
	})
})();
</script>
	</body>
</html>`

type Server struct {
	mutex     sync.Mutex
	templates *template.Template
}

func New() *Server {
	return &Server{}
}

func (s *Server) getTemplates() (*template.Template, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if t := s.templates; t != nil {
		return t, nil
	}

	var t *template.Template
	var root *template.Template

	tmpls := map[string]string{
		"index.tmpl": indexTmpl,
	}
	for name, body := range tmpls {
		if t == nil {
			t = template.New(name)
			root = t
		} else {
			t = t.New(name)
		}

		if _, err := t.Parse(body); err != nil {
			return nil, errors.Wrapf(err, "failed to parse template '%s'", name)
		}
	}

	s.templates = root
	return root, nil
}

func (s *Server) Run(listen string) error {
	return http.ListenAndServe(listen, s)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	switch r.URL.Path {
	case "/":
		s.httpIndex(ctx, w, r)
	case "/generate.json":
		s.httpGenerateValidator(ctx, w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) httpIndex(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	t, err := s.getTemplates()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	buf := bytes.Buffer{}
	if err := t.ExecuteTemplate(&buf, "index.tmpl", nil); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

func (s *Server) httpGenerateValidator(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]interface{})

	buf := bytes.Buffer{}
	io.Copy(&buf, r.Body)

	sc, err := schema.Read(bytes.NewReader(buf.Bytes()))
	if err != nil {
		resp["success"] = false
		resp["message"] = err.Error()
		json.NewEncoder(w).Encode(resp)
		return
	}

	var m map[string]interface{}
	if err := json.NewDecoder(bytes.NewReader(buf.Bytes())).Decode(&m); err != nil {
		resp["success"] = false
		resp["message"] = err.Error()
		json.NewEncoder(w).Encode(resp)
		return
	}

	b := builder.New()
	v, err := b.BuildWithCtx(sc, m)
	if err != nil {
		resp["success"] = false
		resp["message"] = err.Error()
		json.NewEncoder(w).Encode(resp)
		return
	}

	buf.Reset()
	g := jsval.NewGenerator()
	if err := g.Process(&buf, v); err != nil {
		resp["success"] = false
		resp["message"] = err.Error()
		json.NewEncoder(w).Encode(resp)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	resp["success"] = true
	resp["code"] = buf.String()
	json.NewEncoder(w).Encode(resp)
}
