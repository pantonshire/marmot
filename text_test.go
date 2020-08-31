package marmot

import (
  "strings"
  "testing"
)

func TestText(t *testing.T) {
  expect := `Hello! The small robot of the day is Teabot. 1 + 1 = 2`

  cache := TextCache()
  cache.Functions(Std())

  if err := cache.Load(DirExtensions("testdata/text", "tmpl")); err != nil {
    t.Error(err)
  }

  builder := cache.Builder("smolbotbot").
    With("Greeting", "Hello").
    With("Robot", "Teabot")

  str, err := builder.ExecStr()

  if err != nil {
    t.Error(err)
  }

  if strings.TrimSpace(str) != expect {
    t.Error("incorrect text output")
  }
}
