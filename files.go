package marmot

import (
  "fmt"
  "io/ioutil"
  "os"
  "path/filepath"
  "regexp"
  "strings"
  "unicode"
  "unicode/utf8"
)

type TemplateType int

const (
  ContentType  TemplateType = iota
  SkeletonType TemplateType = iota
)

type FileCollection interface {
  Resolve() (ResolvedFileCollection, error)
  Read(path string) ([]byte, error)
  TemplateTypeOf(path string) TemplateType
}

type ResolvedFileCollection struct {
  FileCollection
  Names []string
  Paths map[string]string
}

func (fc ResolvedFileCollection) Read(name string) ([]byte, error) {
  path, ok := fc.Paths[name]
  if !ok {
    return nil, fmt.Errorf("could not load unknown template %s", name)
  }
  return fc.FileCollection.Read(path)
}

type Directory struct {
  path    string
  pattern *regexp.Regexp
}

func DirAll(path string) Directory {
  return Directory{path: path}
}

func DirMatch(path string, pattern string) Directory {
  var r *regexp.Regexp
  if pattern != "" {
    r = regexp.MustCompile(fmt.Sprintf(`^%s$`, pattern))
  }
  return Directory{path: path, pattern: r}
}

func DirExtensions(path string, extensions ...string) Directory {
  if len(extensions) == 0 {
    return DirAll(path)
  }
  escapedExtensions := make([]string, len(extensions))
  for i, extension := range extensions {
    escapedExtensions[i] = regexp.QuoteMeta(extension)
  }
  var pattern string
  if len(escapedExtensions) > 1 {
    pattern = fmt.Sprintf(`.*\.(%s)`, strings.Join(escapedExtensions, `|`))
  } else {
    pattern = fmt.Sprintf(`.*\.%s`, escapedExtensions[0])
  }
  return DirMatch(path, pattern)
}

func (d Directory) Read(path string) ([]byte, error) {
  return ioutil.ReadFile(path)
}

func (d Directory) Resolve() (ResolvedFileCollection, error) {
  var names []string
  paths := make(map[string]string)
  err := filepath.Walk(d.path, func(path string, info os.FileInfo, err error) error {
    if err != nil {
      return err
    } else if info == nil || info.IsDir() {
      return nil
    }
    path = filepath.Clean(path)
    if d.pattern != nil && !d.pattern.MatchString(filepath.ToSlash(path)) {
      return nil
    }
    rel, err := filepath.Rel(d.path, path)
    if err != nil {
      return err
    }
    name := filepath.ToSlash(strings.TrimSuffix(rel, filepath.Ext(rel)))
    if dup, ok := paths[name]; ok {
      return fmt.Errorf("duplicate template name %s used by both %s and %s", name, dup, path)
    }
    paths[name] = path
    names = append(names, name)
    return nil
  })
  return ResolvedFileCollection{
    FileCollection: d,
    Names:          names,
    Paths:          paths,
  }, err
}

//func (d Directory) Walk(step func(name string, tplType TemplateType) error) error {
//  err := filepath.Walk(d.Path, func(path string, info os.FileInfo, err error) error {
//    if err != nil {
//      return err
//    } else if info == nil || info.IsDir() {
//      return nil
//    } else if extension := filepath.Ext(path); extension != d.Extension {
//      return nil
//    }
//    rel, err := filepath.Rel(d.Path, path)
//    if err != nil {
//      return err
//    }
//    name := strings.TrimSuffix(rel, d.Extension)
//    tplType := d.TemplateTypeOf(name)
//    return step(name, tplType)
//  })
//  return err
//}

func (d Directory) TemplateTypeOf(path string) TemplateType {
  if r, _ := utf8.DecodeRuneInString(filepath.Base(path)); unicode.IsUpper(r) {
    return ContentType
  }
  return SkeletonType
}
