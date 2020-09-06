# Marmot
Marmot provides inheritance and caching for text/template and html/template.

## Examples
### Loading template files
Template `templates/MyTemplate.gohtml`:
```html
{{extend "base"}}
{{include "products"}}

{{define "title" -}}
    The answer to life, the universe and everything is {{.Answer}}
{{- end}}

{{define "content"}}
    <h1>Products</h1>
    {{template "product" .ProductName}}
{{end}}
```

Template `templates/base.gohtml`:
```html
<html lang="en">
<head>
    <title>{{template "title" .}}</title>
</head>
<body>{{template "content" .}}</body>
</html>
```

Template `templates/products.gohtml`:
```html
{{define "product" -}}
    <p>A fine {{.}} for sale at a bargain price!</p>
{{- end}}
```

Go code:
```go
package main

import (
  "github.com/pantonshire/marmot"
  "os"
)

func main() {
  cache := marmot.HTMLCache()
  
  //Load all of the templates in the directory templates/ with the extension .gohtml
  directory := marmot.Directory("templates").MatchExtensions("gohtml")
  if err := cache.Load(directory); err != nil {
    panic(err)
  }
  
  //Access a template "templates/MyTemplate.gohtml"
  //The template must be exported (the first character of the file name is a capital letter)
  builder := cache.Builder("mytemplate")
  
  //The values can be accessed inside the template as {{$.Answer}} and {{$.ProductName}}
  builder.With("Answer", 42).With("ProductName", "Sandwich")

  if err := builder.Exec(os.Stdout); err != nil {
    panic(err)
  }
}
```

HTML output:
```html
<html lang="en">
<head>
    <title>The answer to life, the universe and everything is 42</title>
</head>
<body>
    <h1>Products</h1>
    <p>A fine Sandwich for sale at a bargain price!</p>
</body>
</html>
```

### Templates from strings
[Try in the Go playground](https://play.golang.org/p/c_bWx5iZGTU)

```go
package main

import (
  "github.com/pantonshire/marmot"
  "os"
)

func main() {
  cache := marmot.TextCache()
  
  templates := map[string][]byte {
    "base.tmpl":    []byte(`The {{template "jumper" .}} jumps over the lazy dog`),
    "Example.tmpl": []byte(`{{extend "base"}} {{define "jumper"}}{{.Adj1}} {{.Adj2}} {{.Noun}}{{end}}`),
  }
  
  if err := cache.Load(marmot.PreloadedFiles(templates)); err != nil {
    panic(err)
  }
  
  builder := cache.Builder("Example").
    With("Adj1", "quick").
    With("Adj2", "brown").
    With("Noun", "fox")
  
  //The quick brown fox jumps over the lazy dog
  if err := builder.Exec(os.Stdout); err != nil {
    panic(err)
  }
}
```

## License
[The MIT License](./LICENSE)
