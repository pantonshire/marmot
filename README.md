# Marmot
Marmot provides inheritance and caching for text/template and html/template.

## Examples
### Loading template files
```go
package main

import (
  "os"
  "github.com/pantonshire/marmot"
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

### Templates from strings
[Try in the Go playground](https://play.golang.org/p/c_bWx5iZGTU)

```go
package main

import (
  "fmt"
  "github.com/pantonshire/marmot"
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
  
  str, err := builder.ExecStr()
  if err != nil {
    panic(err)
  }
  
  //The quick brown fox jumps over the lazy dog
  fmt.Println(str)
}
```

## License
[The MIT License](./LICENSE)
