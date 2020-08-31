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

      var tpl templateCreator

      for _, parent := range data[name].extends {
        if tpl == nil {
          tpl, err = root.Create(parent, string(data[parent].content), funcs)
          if err != nil {
            return nil, err
          }
        } else {
          _, err = tpl.Create(parent, string(data[parent].content), funcs)
          if err != nil {
            return nil, err
          }
        }
      }

      if tpl == nil {
        tpl, err = root.Create(name, string(data[name].content), funcs)
        if err != nil {
          return nil, err
        }
      } else {
        _, err = tpl.Create(name, string(data[name].content), funcs)
        if err != nil {
          return nil, err
        }
      }

      for _, included := range data[name].includes {
        _, err = tpl.Create(included, string(data[included].content), funcs)
        if err != nil {
          return nil, err
        }
      }

      tcs[strings.ToLower(name)] = tpl
    }
  }

  return tcs, nil
}
