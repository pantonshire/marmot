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

type directory struct {
  path    string
  pattern *regexp.Regexp
}

func newDirectory(path string, pattern *regexp.Regexp) directory {
  return directory{
    path:    filepath.Clean(path),
    pattern: pattern,
  }
}

func DirAll(path string) FileCollection {
  return newDirectory(path, nil)
}

func DirMatch(path string, pattern string) FileCollection {
  var r *regexp.Regexp
  if pattern != "" {
    r = regexp.MustCompile(fmt.Sprintf(`^%s$`, pattern))
  }
  return newDirectory(path, r)
}

func DirExtensions(path string, extensions ...string) FileCollection {
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

func (d directory) Read(path string) ([]byte, error) {
  return ioutil.ReadFile(path)
}

func (d directory) Resolve() (ResolvedFileCollection, error) {
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
      return errDuplicateTemplate(name, dup, path)
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

type pathList struct {
  root  string
  paths []string
}

func PathList(root string, paths ...string) FileCollection {
  pl := pathList{
    root:  filepath.Clean(root),
    paths: make([]string, len(paths)),
  }
  for i, path := range paths {
    pl.paths[i] = filepath.Clean(path)
  }
  return pl
}

func (pl pathList) Read(path string) ([]byte, error) {
  return ioutil.ReadFile(path)
}

func (pl pathList) Resolve() (ResolvedFileCollection, error) {
  var names []string
  paths := make(map[string]string)
  for _, path := range pl.paths {
    name := relPathTemplateName(path)
    if dup, ok := paths[name]; ok {
      return ResolvedFileCollection{}, errDuplicateTemplate(name, dup, path)
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

type preloadedFiles struct {
  data map[string][]byte
}

func PreloadedFiles(data map[string][]byte) FileCollection {
  return preloadedFiles{data: data}
}

func (d preloadedFiles) Read(path string) ([]byte, error) {
  data, ok := d.data[path]
  if !ok {
    return nil, fmt.Errorf("no data stored for %s", path)
  }
  return data, nil
}

func (d preloadedFiles) Resolve() (ResolvedFileCollection, error) {
  var names []string
  paths := make(map[string]string)
  for path, _ := range d.data {
    path = filepath.Clean(path)
    name := relPathTemplateName(path)
    if dup, ok := paths[name]; ok {
      return ResolvedFileCollection{}, errDuplicateTemplate(name, dup, path)
    }
    paths[name] = path
    names = append(names, name)
  }
  return ResolvedFileCollection{
    FileCollection: d,
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

func errDuplicateTemplate(name, dup1, dup2 string) error {
  return fmt.Errorf("duplicate template name %s used by both %s and %s", name, dup1, dup2)
}
