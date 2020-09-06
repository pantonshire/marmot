package marmot

import (
  "bytes"
  "io"
)

type Builder struct {
  cache Cache
  key   string
  data  DataMap
}

func (b *Builder) Exec(w io.Writer) error {
  return b.cache.exec(w, b.key, b.data)
}

func (b *Builder) ExecStr() (string, error) {
  buf := new(bytes.Buffer)
  if err := b.Exec(buf); err != nil {
    return "", err
  }
  return buf.String(), nil
}

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
