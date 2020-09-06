package marmot

import (
  "io"
  "strings"
)

type Cache interface {
  Load(fc FileCollection) error
  WithFuncs(funcs map[string]interface{}) Cache
  Builder(key string) *Builder
  exec(w io.Writer, key string, data map[string]interface{}) error
}

type templateCreator interface {
  Create(name, content string, funcs map[string]interface{}) (templateCreator, error)
}

type TemplateType bool

const (
  Exported   TemplateType = true
  Unexported TemplateType = false
)

func createTemplates(fc FileCollection, root templateCreator, funcs map[string]interface{}) (map[string]templateCreator, error) {
  tcs := make(map[string]templateCreator)
  data := make(map[string]*tpldata)
  files, err := fc.Resolve()
  if err != nil {
    return nil, err
  }
  for _, name := range files.Names {
    path := files.Paths[name]
    if tplType := templateTypeOf(path); tplType == Exported {
      data, err := recurseTemplates(files, data, name)
      if err != nil {
        return nil, err
      }
      templateStack, i := make([]string, 1 + len(data[name].extends) + len(data[name].includes)), 0
      for _, parent := range data[name].extends {
        templateStack[i] = parent
        i++
      }
      templateStack[i] = name
      i++
      for _, included := range data[name].includes {
        templateStack[i] = included
        i++
      }
      tpl, err := root.Create(templateStack[0], string(data[templateStack[0]].content), funcs)
      if err != nil {
        return nil, err
      }
      for j := 1; j < len(templateStack); j++ {
        _, err := tpl.Create(templateStack[j], string(data[templateStack[j]].content), nil)
        if err != nil {
          return nil, err
        }
      }
      tcs[strings.ToLower(name)] = tpl
    }
  }
  return tcs, nil
}
