package marmot

import (
  "fmt"
  "html/template"
  "io"
  "sync"
)

type htmlCache struct {
  lock      sync.RWMutex
  templates map[string]*template.Template
  funcs     template.FuncMap
  export    ExportRule
}

// Returns a new cache which uses html/template.
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

func (c *htmlCache) WithFuncs(funcs FuncMap) Cache {
  for key, fn := range funcs {
    c.funcs[key] = fn
  }
  return c
}

func (c *htmlCache) WithExportRule(rule ExportRule) Cache {
  c.export = rule
  return c
}

func (c *htmlCache) Builder(key string) *Builder {
  return &Builder{cache: c, key: key, data: make(map[string]interface{})}
}

func (c *htmlCache) exec(w io.Writer, key string, data DataMap) error {
  if tpl, ok := c.lookup(key); ok {
    return tpl.Execute(w, data)
  }
  return fmt.Errorf("template %s not found", key)
}

func (c *htmlCache) exportRule() ExportRule {
  return c.export
}

func (c *htmlCache) functions() FuncMap {
  return FuncMap(c.funcs)
}

func (c *htmlCache) lookup(key string) (*template.Template, bool) {
  c.lock.RLock()
  defer c.lock.RUnlock()
  tpl, ok := c.templates[templateKey(key)]
  return tpl, ok
}

func (c *htmlCache) load(fc FileCollection) error {
  tcs, err := createTemplates(c, fc, htmlTemplateCreator{})
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

func (tc htmlTemplateCreator) Create(name, content string, funcs FuncMap) (templateCreator, error) {
  var tmpl *template.Template
  var err error
  if tc.template != nil {
    tmpl, err = tc.template.New(name).Parse(content)
  } else {
    tmpl, err = template.New(name).Funcs(template.FuncMap(funcs)).Parse(content)
  }
  if err != nil {
    return htmlTemplateCreator{}, err
  }
  return htmlTemplateCreator{template: tmpl}, nil
}
