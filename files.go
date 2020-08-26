package marmot

import (
    "os"
    "path/filepath"
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
    Walk(step func(path string, tplType TemplateType) error) error
}

type Directory struct {
    Path      string
    Extension string
}

func (d Directory) Find(path string) string {
    return filepath.Join(d.Path, filepath.Clean("/"+path+d.Extension))
}

func (d Directory) Walk(step func(path string, tplType TemplateType) error) error {
    err := filepath.Walk(d.Path, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        } else if info == nil || info.IsDir() {
            return nil
        } else if extension := filepath.Ext(path); extension != d.Extension {
            return nil
        }
        return step(path, d.templateTypeOf(path))
    })
    return err
}

func (d Directory) templateTypeOf(path string) TemplateType {
    if r, _ := utf8.DecodeRuneInString(filepath.Base(path)); unicode.IsUpper(r) {
        return ContentType
    }
    return SkeletonType
}
