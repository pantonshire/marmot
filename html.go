package marmot

import (
  "fmt"
  "html/template"
  "io"
  "strings"
  "sync"
)

type htmlCache struct {
  lock      sync.RWMutex
  templates map[string]*template.Template
  funcs     template.FuncMap
}

func HTMLCache() Cache {
  return &htmlCache{
    templates: make(map[string]*template.Template),
    funcs:     make(template.FuncMap),
  }
}

func (c *htmlCache) Load(fc FileCollection) error {
  c.lock.Lock()
  defer c.lock.Unlock()
  return c.load(fc)
}

func (c *htmlCache) WithFuncs(funcs map[string]interface{}) Cache {
  for key, fn := range funcs {
    c.funcs[key] = fn
  }
  return c
}

func (c *htmlCache) Builder(key string) *Builder {
  return &Builder{cache: c, key: key, data: make(map[string]interface{})}
}

func (c *htmlCache) exec(w io.Writer, key string, data map[string]interface{}) error {
  if tpl, ok := c.lookup(key); ok {
    return tpl.Execute(w, data)
  }
  return fmt.Errorf("template %s not found", key)
}

func (c *htmlCache) lookup(key string) (*template.Template, bool) {
  c.lock.RLock()
  defer c.lock.RUnlock()
  tpl, ok := c.templates[strings.ToLower(key)]
  return tpl, ok
}

func (c *htmlCache) load(fc FileCollection) error {
  tcs, err := createTemplates(fc, htmlTemplateCreator{}, c.funcs)
  if err != nil {
    return err
  }
  c.templates = make(map[string]*template.Template)
  for name, tc := range tcs {
    c.templates[name] = tc.(htmlTemplateCreator).template
  }
  return nil
}

type htmlTemplateCreator struct {
  template *template.Template
}

func (tc htmlTemplateCreator) Create(name, content string, funcs map[string]interface{}) (templateCreator, error) {
  var tmpl *template.Template
  var err error
  if tc.template != nil {
    tmpl, err = tc.template.New(name).Funcs(funcs).Parse(content)
  } else {
    tmpl, err = template.New(name).Funcs(funcs).Parse(content)
  }
  if err != nil {
    return htmlTemplateCreator{}, err
  }
  return htmlTemplateCreator{template: tmpl}, nil
}
