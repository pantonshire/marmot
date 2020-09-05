package marmot

import (
  "io"
  "strings"
)

type Cache interface {
  Load(fc FileCollection) error
  Functions(funcs map[string]interface{})
  Builder(key string) *Builder
  exec(w io.Writer, key string, data map[string]interface{}) error
}

type templateCreator interface {
  Create(name, content string, funcs map[string]interface{}) (templateCreator, error)
}

func createTemplates(fc FileCollection, root templateCreator, funcs map[string]interface{}) (map[string]templateCreator, error) {
  tcs := make(map[string]templateCreator)
  data := make(map[string]*tpldata)
  files, err := fc.Resolve()
  if err != nil {
    return nil, err
  }
  for _, name := range files.Names {
    path := files.Paths[name]
    if tplType := templateTypeOf(path); tplType == contentType {
      data, err := recurseTemplates(files, data, name)
      if err != nil {
        return nil, err
      }
      var templateStack []string
      for _, parent := range data[name].extends {
        templateStack = append(templateStack, parent)
      }
      templateStack = append(templateStack, name)
      for _, included := range data[name].includes {
        templateStack = append(templateStack, included)
      }
      tpl, err := root.Create(templateStack[0], string(data[templateStack[0]].content), funcs)
      if err != nil {
        return nil, err
      }
      for i := 1; i < len(templateStack); i++ {
        _, err := tpl.Create(templateStack[i], string(data[templateStack[i]].content), nil)
        if err != nil {
          return nil, err
        }
      }
      tcs[strings.ToLower(name)] = tpl
    }
  }
  return tcs, nil
}
