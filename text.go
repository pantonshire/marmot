package marmot

import (
  "fmt"
  "io"
  "strings"
  "sync"
  "text/template"
)

type textCache struct {
  lock      sync.RWMutex
  templates map[string]*template.Template
  funcs     template.FuncMap
}

func TextCache() Cache {
  return &textCache{
    templates: make(map[string]*template.Template),
    funcs:     make(template.FuncMap),
  }
}

func (c *textCache) Load(fc FileCollection) error {
  c.lock.Lock()
  defer c.lock.Unlock()
  return c.load(fc)
}

func (c *textCache) WithFuncs(funcs map[string]interface{}) Cache {
  for key, fn := range funcs {
    c.funcs[key] = fn
  }
  return c
}

func (c *textCache) Builder(key string) *Builder {
  return &Builder{cache: c, key: key, data: make(map[string]interface{})}
}

func (c *textCache) exec(w io.Writer, key string, data map[string]interface{}) error {
  if tpl, ok := c.lookup(key); ok {
    return tpl.Execute(w, data)
  }
  return fmt.Errorf("template %s not found", key)
}

func (c *textCache) lookup(key string) (*template.Template, bool) {
  c.lock.RLock()
  defer c.lock.RUnlock()
  tpl, ok := c.templates[strings.ToLower(key)]
  return tpl, ok
}

func (c *textCache) load(fc FileCollection) error {
  tcs, err := createTemplates(fc, textTemplateCreator{}, c.funcs)
  if err != nil {
    return err
  }
  c.templates = make(map[string]*template.Template)
  for name, tc := range tcs {
    c.templates[name] = tc.(textTemplateCreator).template
  }
  return nil
}

type textTemplateCreator struct {
  template *template.Template
}

func (tc textTemplateCreator) Create(name, content string, funcs map[string]interface{}) (templateCreator, error) {
  var tmpl *template.Template
  var err error
  if tc.template != nil {
    tmpl, err = tc.template.New(name).Parse(content)
  } else {
    tmpl, err = template.New(name).Funcs(funcs).Parse(content)
  }
  if err != nil {
    return textTemplateCreator{}, err
  }
  return textTemplateCreator{template: tmpl}, nil
}
