package marmot

import (
    "fmt"
    "html/template"
    "io"
    "strings"
    "sync"
)

type HTMLCache struct {
    lock      sync.RWMutex
    templates map[string]*template.Template
    funcs     template.FuncMap
}

func (c *HTMLCache) Load(fc FileCollection) error {
    c.lock.Lock()
    defer c.lock.Unlock()
    return c.load(fc)
}

func (c *HTMLCache) Functions(funcs map[string]interface{}) {
    for key, fn := range funcs {
        c.funcs[key] = fn
    }
}

func (c *HTMLCache) Builder(key string) *Builder {
    return &Builder{cache: c, key: key, data: make(map[string]interface{})}
}

func (c *HTMLCache) exec(w io.Writer, key string, data map[string]interface{}) error {
    if tpl, ok := c.lookup(key); ok {
        return tpl.Execute(w, data)
    }
    return fmt.Errorf("template %s not found", key)
}

func (c *HTMLCache) lookup(key string) (*template.Template, bool) {
    c.lock.RLock()
    defer c.lock.RUnlock()
    tpl, ok := c.templates[strings.ToLower(key)]
    return tpl, ok
}

func (c *HTMLCache) load(fc FileCollection) error {
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

func recurseTemplates(fc FileCollection, data map[string]*tpldata, name string) (map[string]*tpldata, error) {
    if _, ok := data[name]; ok {
        return data, nil
    }

    tplData, err := loadTemplate(fc, name)
    if err != nil {
        return data, err
    }

    data[name] = &tplData

    dependencies := make(map[string]bool)
    var extends, includes []string

    for _, parent := range tplData.extends {
        data, err = recurseTemplates(fc, data, parent)
        if err != nil {
            return data, err
        }

        for _, dep := range data[parent].extends {
            if !dependencies[dep] {
                extends, dependencies[dep] = append(extends, dep), true
            }
        }

        if !dependencies[parent] {
            extends, dependencies[parent] = append(extends, parent), true
        }

        for _, dep := range data[parent].includes {
            if !dependencies[dep] {
                extends, dependencies[dep] = append(extends, dep), true
            }
        }
    }

    for _, included := range tplData.includes {
        data, err = recurseTemplates(fc, data, included)
        if err != nil {
            return data, err
        }

        for _, dep := range data[included].extends {
            if !dependencies[dep] {
                includes, dependencies[dep] = append(includes, dep), true
            }
        }

        if !dependencies[included] {
            includes, dependencies[included] = append(includes, included), true
        }

        for _, dep := range data[included].includes {
            if !dependencies[dep] {
                includes, dependencies[dep] = append(includes, dep), true
            }
        }
    }

    tplData.extends, tplData.includes = extends, includes

    return data, err
}
