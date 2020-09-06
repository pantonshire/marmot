package marmot

import (
  "io"
  "path/filepath"
  "strings"
  "unicode"
  "unicode/utf8"
)

type Cache interface {
  Load(FileCollection) error
  WithFuncs(FuncMap) Cache
  Builder(key string) *Builder
  exec(w io.Writer, key string, data DataMap) error
  exportRule() ExportRule
  functions() FuncMap
}

type templateCreator interface {
  Create(name, content string, funcs FuncMap) (templateCreator, error)
}

type FuncMap map[string]interface{}
type DataMap map[string]interface{}
type TemplateType bool
type ExportRule func(path string) TemplateType

const (
  Exported   TemplateType = true
  Unexported TemplateType = false
)

func createTemplates(cache Cache, fc FileCollection, root templateCreator) (map[string]templateCreator, error) {
  tcs := make(map[string]templateCreator)
  data := make(map[string]*tpldata)
  files, err := fc.Resolve()
  if err != nil {
    return nil, err
  }
  funcs := cache.functions()
  //var exportRule ExportRule
  //if customRule :=
  for _, name := range files.Names {
    path := files.Paths[name]
    if tplType := defaultExportRule(path); tplType == Exported {
      data, err := recurseTemplates(files, data, name)
      if err != nil {
        return nil, err
      }
      templateStack, i := make([]string, 1+len(data[name].extends)+len(data[name].includes)), 0
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

func defaultExportRule(path string) TemplateType {
  if r, _ := utf8.DecodeRuneInString(filepath.Base(path)); unicode.IsUpper(r) {
    return Exported
  }
  return Unexported
}
