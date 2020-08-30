# Marmot
Wrapper providing caching and inheritance for text/template and html/template

## Usage
```go
package main

import (
  "fmt"
  "github.com/pantonshire/marmot"
)

func main() {
  cache := marmot.HTML()
  
  //Load all of the templates in the directory templates/ with the extension .gohtml
  if err := cache.Load(marmot.DirExtensions("templates", "gohtml")); err != nil {
    panic(err)
  }
  
  //Access a template "templates/MyTemplate.gohtml"
  //The template must be exported (the first character of the file name is a capital letter)
  builder := cache.Builder("mytemplate")
  
  //The values can be accessed inside the template as {{$.Answer}} and {{$.ProductName}}
  builder.With("Answer", 42).With("ProductName", "Sandwich")
  
  str, err := builder.ExecStr()
  if err != nil {
    panic(err)
  }
  
  fmt.Println(str)
}
```

## License
[The MIT License](./LICENSE)
