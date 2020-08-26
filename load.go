package marmot

import (
    "io/ioutil"
    "regexp"
    "strings"
)

type tpldata struct {
    content  []byte
    extends  []string
    includes []string
}

var (
    reExtend  = regexp.MustCompile(`(?s){{\s*extend\s[^{}]*}}`)
    reInclude = regexp.MustCompile(`(?s){{\s*include\s[^{}]*}}`)
)

func loadTemplate(fc FileCollection, name string) (data tpldata, err error) {
    content, err := ioutil.ReadFile(fc.Find(name))
    if err != nil {
        return data, err
    }

    if match := reExtend.FindIndex(content); match != nil {
        data.extends = parseDependencies(string(content[match[0]:match[1]]))
        content = append(content[:match[0]], content[match[1]:]...)
    }

    if match := reInclude.FindIndex(content); match != nil {
        data.includes = parseDependencies(string(content[match[0]:match[1]]))
        content = append(content[:match[0]], content[match[1]:]...)
    }

    data.content = content

    return data, nil
}

func parseDependencies(dependencyStatement string) []string {
    return strings.Fields(dependencyStatement[2 : len(dependencyStatement)-2])[1:]
}
