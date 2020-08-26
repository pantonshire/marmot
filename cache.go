package marmot

import "io"

type Cache interface {
    Load(fc FileCollection) error
    Functions(funcs map[string]interface{})
    Builder(key string) *Builder
    exec(w io.Writer, key string, data map[string]interface{}) error
}
