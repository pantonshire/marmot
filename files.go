package marmot

import (
    "os"
    "path/filepath"
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
    Find(path string) string
    Walk(step func(name string, tplType TemplateType) error) error
    TemplateTypeOf(path string) TemplateType
}

type Directory struct {
    Path      string
    Extension string
}

func (d Directory) Find(path string) string {
    return filepath.Join(d.Path, filepath.Clean("/"+path+d.Extension))
}

func (d Directory) Walk(step func(name string, tplType TemplateType) error) error {
    err := filepath.Walk(d.Path, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        } else if info == nil || info.IsDir() {
            return nil
        } else if extension := filepath.Ext(path); extension != d.Extension {
            return nil
        }
        rel, err := filepath.Rel(d.Path, path)
        if err != nil {
            return err
        }
        name := strings.TrimSuffix(rel, d.Extension)
        tplType := d.TemplateTypeOf(name)
        return step(name, tplType)
    })
    return err
}

func (d Directory) TemplateTypeOf(path string) TemplateType {
    if r, _ := utf8.DecodeRuneInString(filepath.Base(path)); unicode.IsUpper(r) {
        return ContentType
    }
    return SkeletonType
}
