package marmot

import (
  "fmt"
  "io"
  "strings"
  "sync"
  "text/template"
)

type TextCache struct {
  lock      sync.RWMutex
  templates map[string]*template.Template
  funcs     template.FuncMap
}

func Text() *TextCache {
  return &TextCache{
    templates: make(map[string]*template.Template),
    funcs:     make(template.FuncMap),
  }
}

func (c *TextCache) Load(fc FileCollection) error {
  c.lock.Lock()
  defer c.lock.Unlock()
  return c.load(fc)
}

func (c *TextCache) Functions(funcs map[string]interface{}) {
  for key, fn := range funcs {
    c.funcs[key] = fn
  }
}

func (c *TextCache) Builder(key string) *Builder {
  return &Builder{cache: c, key: key, data: make(map[string]interface{})}
}

func (c *TextCache) exec(w io.Writer, key string, data map[string]interface{}) error {
  if tpl, ok := c.lookup(key); ok {
    return tpl.Execute(w, data)
  }
  return fmt.Errorf("template %s not found", key)
}

func (c *TextCache) lookup(key string) (*template.Template, bool) {
  c.lock.RLock()
  defer c.lock.RUnlock()
  tpl, ok := c.templates[strings.ToLower(key)]
  return tpl, ok
}

func (c *TextCache) load(fc FileCollection) error {
  c.templates = make(map[string]*template.Template)

  data := make(map[string]*tpldata)

  return fc.Walk(func(name string, tplType TemplateType) error {
    if tplType != ContentType {
      return nil
    }

    data, err := recurseTemplates(fc, data, name)
    if err != nil {
      return err
    }

    var tpl *template.Template

    for _, parent := range data[name].extends {
      if tpl == nil {
        tpl, err = template.New(parent).Funcs(c.funcs).Parse(string(data[parent].content))
        if err != nil {
          return err
        }
      } else {
        _, err = tpl.New(parent).Parse(string(data[parent].content))
        if err != nil {
          return err
        }
      }
    }

    if tpl == nil {
      tpl, err = template.New(name).Funcs(c.funcs).Parse(string(data[name].content))
      if err != nil {
        return err
      }
    } else {
      _, err = tpl.New(name).Parse(string(data[name].content))
      if err != nil {
        return err
      }
    }

    for _, included := range data[name].includes {
      _, err = tpl.New(included).Parse(string(data[included].content))
      if err != nil {
        return err
      }
    }

    c.templates[strings.ToLower(name)] = tpl

    return nil
  })
}
