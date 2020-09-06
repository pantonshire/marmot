package marmot

import (
  "io"
  "path"
  "regexp"
  "strings"
  "unicode"
  "unicode/utf8"
)

// Responsible for loading and storing a collection of templates, and creating Builder structs which are used to
// execute the templates as necessary.
type Cache interface {
  // Loads all of the templates in the given FileCollection, resolving any inheritance hierarchies between templates.
  //
  // A template can extend another using {{extend "parent"}} and can include other templates using
  // {{include "tmpl1 tmpl2 ..."}}. An argument to extend or include should be the template's path in forward slash
  // format minus its extension. If the FileCollection is a Dir, then the paths should be relative to the path of
  // the Dir.
  //
  // Once this function returns, any exported templates in the FileCollection can be executed via Cache.Builder.
  // By default, exported templates are ones whose file name begins with a capital letter, but this behaviour can be
  // overridden using Cache.WithExportRule.
  Load(FileCollection) error

  // Specifies a collection of functions which can be used in the templates.
  WithFuncs(FuncMap) Cache

  // Specifies a custom export rule that is used to determine whether templates are exported or not; exported
  // templates can be executed via Cache.Builder, while unexported templates' only purpose is to be inherited from.
  WithExportRule(ExportRule) Cache

  // Creates a new Builder for the template indexed by the given key.
  //
  // The key is the template's path in forward slash format minus its extension, case insensitive. If the
  // FileCollection used to load the templates was a Dir, then the paths should be relative to the path of the Dir.
  //
  // For example, if you have a template templates/customer/Checkout.gohtml:
  //  _ = cache.Load(marmot.Directory("templates"))
  //  builder := cache.Builder("customer/checkout")
  Builder(key string) *Builder

  exec(w io.Writer, key string, data DataMap) error
  exportRule() ExportRule
  functions() FuncMap
}

type FuncMap map[string]interface{}

// A DataMap stores values to be passed to a template on execution, indexed by string keys. The template can use
// the key to retrieve the value: {{$.key}}.
type DataMap map[string]interface{}

// A TemplateType is either Exported or Unexported.
// This represents whether or not a template can be used by Cache.Builder.
type TemplateType bool

// An ExportRule is a function which takes a template's name (its path in forward slash format minus its extension)
// and returns whether the template should be exported or not.
type ExportRule func(name string) TemplateType

const (
  // Exported templates can be used by Cache.Builder.
  Exported TemplateType = true

  // Unexported templates cannot be used by Cache.Builder.
  // Their only purpose is for other templates to inherit from them.
  Unexported TemplateType = false
)

type templateCreator interface {
  Create(name, content string, funcs FuncMap) (templateCreator, error)
}

type tpldata struct {
  content  []byte
  extends  []string
  includes []string
}

var (
  reExtend  = regexp.MustCompile(`(?s){{\s*extend(\s[^{}]*}}|}})(\s*(\r\n|\r|\n))*`)
  reInclude = regexp.MustCompile(`(?s){{\s*include(\s[^{}]*}}|}})(\s*(\r\n|\r|\n))*`)
  reString  = regexp.MustCompile(`"([^"\\]|\\.)*"`)
)

func createTemplates(cache Cache, fc FileCollection, root templateCreator) (map[string]templateCreator, error) {
  tcs := make(map[string]templateCreator)
  data := make(map[string]*tpldata)
  files, err := fc.Resolve()
  if err != nil {
    return nil, err
  }
  var exportRule ExportRule
  if customRule := cache.exportRule(); customRule != nil {
    exportRule = customRule
  } else {
    exportRule = defaultExportRule
  }
  funcs := cache.functions()
  for _, name := range files.Names {
    if tplType := exportRule(name); tplType == Exported {
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
      tcs[templateKey(name)] = tpl
    }
  }
  return tcs, nil
}

func loadTemplate(fc ResolvedFileCollection, name string) (data tpldata, err error) {
  content, err := fc.Read(name)
  if err != nil {
    return data, err
  }

  if match := reExtend.FindIndex(content); match != nil {
    data.extends = parseDependencies(strings.TrimSpace(string(content[match[0]:match[1]])))
    content = append(content[:match[0]], content[match[1]:]...)
  }

  if match := reInclude.FindIndex(content); match != nil {
    data.includes = parseDependencies(strings.TrimSpace(string(content[match[0]:match[1]])))
    content = append(content[:match[0]], content[match[1]:]...)
  }

  data.content = content

  return data, nil
}

func parseDependencies(dependencyStatement string) []string {
  var dependencies []string
  for _, str := range reString.FindAllString(dependencyStatement, -1) {
    if len(str) >= 2 {
      dependencies = append(dependencies, strings.Fields(str[1:len(str)-1])...)
    }
  }
  return dependencies
}

func recurseTemplates(fc ResolvedFileCollection, data map[string]*tpldata, name string) (map[string]*tpldata, error) {
  if _, ok := data[name]; ok {
    return data, nil
  }

  tplData, err := loadTemplate(fc, name)
  if err != nil {
    return data, err
  }

  data[name] = &tplData

  dependencies := make(map[string]bool)
  var extends, includes []string

  for _, parent := range tplData.extends {
    data, err = recurseTemplates(fc, data, parent)
    if err != nil {
      return data, err
    }

    for _, dep := range data[parent].extends {
      if !dependencies[dep] {
        extends, dependencies[dep] = append(extends, dep), true
      }
    }

    if !dependencies[parent] {
      extends, dependencies[parent] = append(extends, parent), true
    }

    for _, dep := range data[parent].includes {
      if !dependencies[dep] {
        extends, dependencies[dep] = append(extends, dep), true
      }
    }
  }

  for _, included := range tplData.includes {
    data, err = recurseTemplates(fc, data, included)
    if err != nil {
      return data, err
    }

    for _, dep := range data[included].extends {
      if !dependencies[dep] {
        includes, dependencies[dep] = append(includes, dep), true
      }
    }

    if !dependencies[included] {
      includes, dependencies[included] = append(includes, included), true
    }

    for _, dep := range data[included].includes {
      if !dependencies[dep] {
        includes, dependencies[dep] = append(includes, dep), true
      }
    }
  }

  tplData.extends, tplData.includes = extends, includes

  return data, err
}

func defaultExportRule(name string) TemplateType {
  if r, _ := utf8.DecodeRuneInString(path.Base(name)); unicode.IsUpper(r) {
    return Exported
  }
  return Unexported
}

func templateKey(name string) string {
  return strings.ToLower(name)
}
