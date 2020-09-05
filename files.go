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

type templateType bool

const (
  exported   templateType = true
  unexported templateType = false
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

type Dir interface {
  FileCollection
  PartialMatch(pattern *regexp.Regexp) Dir
  FullMatch(pattern *regexp.Regexp) Dir
  MatchExtensions(extensions ...string) Dir
}

type directory struct {
  path     string
  patterns []*regexp.Regexp
}

func Directory(path string) Dir {
  return directory{path: filepath.Clean(path)}
}

// Files in the directory will be matched if the specified pattern appears anywhere in the file's path.
// The paths checked against this pattern will be cleaned paths in slash format relative to the directory.
func (d directory) PartialMatch(pattern *regexp.Regexp) Dir {
  d.patterns = append(d.patterns, pattern)
  return d
}

// Files in the directory will be matched if the entire path conforms to the specified pattern.
// The paths checked against this pattern will be cleaned paths in slash format relative to the directory.
func (d directory) FullMatch(pattern *regexp.Regexp) Dir {
  d.patterns = append(d.patterns, regexp.MustCompile(fmt.Sprintf("^%s$", pattern.String())))
  return d
}

// Files in the directory will be matched if they have any of the given extensions.
// The extensions should exclude the leading dot.
func (d directory) MatchExtensions(extensions ...string) Dir {
  if len(extensions) == 0 {
    return d
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
  d.patterns = append(d.patterns, regexp.MustCompile(pattern))
  return d
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
    rel, err := filepath.Rel(d.path, path)
    if err != nil {
      return err
    }
    if !d.match(rel) {
      return nil
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

func (d directory) match(path string) bool {
  if len(d.patterns) == 0 {
    return true
  }
  slashPath := filepath.ToSlash(path)
  for _, pattern := range d.patterns {
    if pattern.MatchString(slashPath) {
      return true
    }
  }
  return false
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
    return exported
  }
  return unexported
}

func errDuplicateTemplate(name, dup1, dup2 string) error {
  return fmt.Errorf("duplicate template name %s used by both %s and %s", name, dup1, dup2)
}
