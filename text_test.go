package marmot

import (
  "strings"
  "testing"
)

func TestText(t *testing.T) {
  tests := []struct {
    fc       FileCollection
    template string
    expect   string
  }{
    {
      DirExtensions("testdata/text", "tmpl"),
      "smolbotbot",
      `Hello! The small robot of the day is Teabot. 1 + 1 = 2`,
    },
    {
      PathList("testdata/text", "base.tmpl", "greeting.tmpl", "Smolbotbot.tmpl"),
      "smolbotbot",
      `Hello! The small robot of the day is Teabot. 1 + 1 = 2`,
    },
    {
      PreloadedFiles(map[string][]byte{
        "base.tmpl": []byte(`{{template "greeting" .}} {{template "message" .}}`),
        "greeting.tmpl": []byte(strings.Join([]string{
          `{{define "greeting" -}}`,
          `    {{.Greeting}}!`,
          `{{- end}}`,
        }, "\n")),
        "Smolbotbot.tmpl": []byte(strings.Join([]string{
          `{{extend "base"}}`,
          `{{include "greeting"}}`,
          `{{define "message" -}}`,
          `    The small robot of the day is {{.Robot}}. 1 + 1 = {{add 1 1}}`,
          `{{- end}}`,
        }, "\n")),
      }),
      "smolbotbot",
      `Hello! The small robot of the day is Teabot. 1 + 1 = 2`,
    },
  }

  for _, testData := range tests {
    cache := TextCache()
    cache.Functions(Std())

    if err := cache.Load(testData.fc); err != nil {
      t.Error(err)
    }

    builder := cache.Builder(testData.template).
      With("Greeting", "Hello").
      With("Robot", "Teabot")

    str, err := builder.ExecStr()

    if err != nil {
      t.Error(err)
    }

    if strings.TrimSpace(str) != testData.expect {
      t.Error("incorrect text output")
    }
  }
}
