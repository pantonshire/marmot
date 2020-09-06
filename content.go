package marmot

import (
  "bytes"
  "io"
)

// A Builder is used to execute a template.
//
// Data to be used in the template can be provided to the builder via the Builder.With and Builder.WithAll methods.
// Once all of the data required has been provided, Builder.Exec is used to execute the template and write the
// output to a given destination.
//
// To create a Builder, use Cache.Builder.
type Builder struct {
  cache Cache
  key   string
  data  DataMap
}

// Looks up the template in the Cache, executes the template and writes the output to w.
func (b *Builder) Exec(w io.Writer) error {
  return b.cache.exec(w, b.key, b.data)
}

// Looks up the template in the Cache, executes the template and writes the output to a string.
func (b *Builder) ExecStr() (string, error) {
  buf := new(bytes.Buffer)
  if err := b.Exec(buf); err != nil {
    return "", err
  }
  return buf.String(), nil
}

// Specifies a single piece of data to be used in the template when executing it.
// The template can reference the value val as {{$.key}}.
//
// If the given key is already in use by this builder, the old value will be overwritten.
//
// To provide multiple pieces of data in a single function call, use Builder.WithAll.
func (b *Builder) With(key string, val interface{}) *Builder {
  b.data[key] = val
  return b
}

func (b *Builder) WithAll(data DataMap) *Builder {
  for key, val := range data {
    b.data[key] = val
  }
  return b
}
