package web

import "text/template"

const bootstrapTemplateStr = `
<!doctype html>
<html lang="en">

<head>
  <title>{{ .Bootstrap.Title }}</title>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="route-pattern" content="{{ .RoutePattern }}">
  {{- range .Bootstrap.Meta }}
  <meta {{ range $k, $v :=. }} {{ $k }}="{{ $v }}"{{end}}/>
  {{ end -}}
  {{- range .Bootstrap.Link}}
  <link {{ range $k, $v :=. }} {{ $k }}="{{ $v }}"{{end}}/>
  {{ end -}}
  {{- range .Bootstrap.Script}}
  <script {{ range $k, $v :=. }} {{ $k }}="{{ $v }}"{{end}}></script>
  {{ end -}}
  <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0-alpha3/dist/css/bootstrap.min.css" rel="stylesheet" integrity="sha384-KK94CHFLLe+nY2dmCWGMq91rCGa5gtU4mk92HdvYe+M/SXH301p5ILy+dN9+nJOZ" crossorigin="anonymous">
  <script src="/static/wasm_exec.js"></script>
  <script>
    'use strict';

    const WASM_URL = '{{ .WasmPath }}';
    const go = new Go(); // Defined in wasm_exec.js

    var wasm;

    if ('instantiateStreaming' in WebAssembly) {
      WebAssembly.instantiateStreaming(fetch(WASM_URL), go.importObject).then(function (obj) {
        wasm = obj.instance;
        go.run(wasm);
      })
    } else {
      fetch(WASM_URL).then(resp =>
        resp.arrayBuffer()
      ).then(bytes =>
        WebAssembly.instantiate(bytes, go.importObject).then(function (obj) {
          wasm = obj.instance;
          go.run(wasm);
        })
      )
    }
  </script>
  <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0-alpha3/dist/js/bootstrap.bundle.min.js" integrity="sha384-ENjdO4Dr2bkBIFxQpeoTz1HIcje39Wm4jDKdf19U8gI4ddQ3GYNS7NTKfAdVQSZe" crossorigin="anonymous"></script>
</head>

<body>
{{ printf "%#v" .Bootstrap  }}
{{ .WasmPath }}

</body>

</html>
`

var bootstrapTemplate = template.Must(template.New("bootstrap").Parse(bootstrapTemplateStr))
