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

type templateType int

const (
  contentType  templateType = iota
  skeletonType templateType = iota
)

type FileCollection interface {
  Resolve() (ResolvedFileCollection, error)
  Read(path string) ([]byte, error)
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

func newDirectory(path string, pattern *regexp.Regexp) Directory {
  return Directory{
    path:    filepath.Clean(path),
    pattern: pattern,
  }
}

func DirAll(path string) Directory {
  return newDirectory(path, nil)
}

func DirMatch(path string, pattern string) Directory {
  var r *regexp.Regexp
  if pattern != "" {
    r = regexp.MustCompile(fmt.Sprintf(`^%s$`, pattern))
  }
  return newDirectory(path, r)
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
    name := relPathTemplateName(rel)
    if dup, ok := paths[name]; ok {
      return fmt.Errorf("duplicate template name %s used by both %s and %s", name, dup, path)
    }
    paths[name] = path
    names = append(names, name)
    return nil
  })
  if err != nil {
    return ResolvedFileCollection{}, err
  }
  return ResolvedFileCollection{
    FileCollection: d,
    Names:          names,
    Paths:          paths,
  }, nil
}

type PathList struct {
  root  string
  paths []string
}

func Paths(root string, paths ...string) PathList {
  pl := PathList{
    root:  filepath.Clean(root),
    paths: make([]string, len(paths)),
  }
  for i, path := range paths {
    pl.paths[i] = filepath.Clean(path)
  }
  return pl
}

func (pl PathList) Read(path string) ([]byte, error) {
  return ioutil.ReadFile(path)
}

func (pl PathList) Resolve() (ResolvedFileCollection, error) {
  var names []string
  paths := make(map[string]string)
  for _, path := range pl.paths {
    name := relPathTemplateName(path)
    if dup, ok := paths[name]; ok {
      return ResolvedFileCollection{},
        fmt.Errorf("duplicate template name %s used by both %s and %s", name, dup, path)
    }
    paths[name] = filepath.Join(pl.root, path)
    names = append(names, name)
  }
  return ResolvedFileCollection{
    FileCollection: pl,
    Names:          names,
    Paths:          paths,
  }, nil
}

func relPathTemplateName(path string) string {
  return filepath.ToSlash(strings.TrimSuffix(path, filepath.Ext(path)))
}

func templateTypeOf(path string) templateType {
  if r, _ := utf8.DecodeRuneInString(filepath.Base(path)); unicode.IsUpper(r) {
    return contentType
  }
  return skeletonType
}
