package marmot

import (
  "regexp"
  "strings"
)

type tpldata struct {
  content  []byte
  extends  []string
  includes []string
}

var (
  reExtend  = regexp.MustCompile(`(?s){{\s*extend(\s[^{}]*}}|}})(\s*(\r\n|\r|\n))*`)
  reInclude = regexp.MustCompile(`(?s){{\s*include(\s[^{}]*}}|}})(\s*(\r\n|\r|\n))*`)
  reString  = regexp.MustCompile(`"([^"\\]|\\.)*"`)
)

func loadTemplate(fc ResolvedFileCollection, name string) (data tpldata, err error) {
  content, err := fc.Read(name)
  if err != nil {
    return data, err
  }

  if match := reExtend.FindIndex(content); match != nil {
    data.extends = parseDependencies(strings.TrimSpace(string(content[match[0]:match[1]])))
    content = append(content[:match[0]], content[match[1]:]...)
  }

  if match := reInclude.FindIndex(content); match != nil {
    data.includes = parseDependencies(strings.TrimSpace(string(content[match[0]:match[1]])))
    content = append(content[:match[0]], content[match[1]:]...)
  }

  data.content = content

  return data, nil
}

func parseDependencies(dependencyStatement string) []string {
  var dependencies []string
  for _, str := range reString.FindAllString(dependencyStatement, -1) {
    if len(str) >= 2 {
      dependencies = append(dependencies, strings.Fields(str[1 : len(str)-1])...)
    }
  }
  return dependencies
}

// TODO: respect cache's case sensitivity
func recurseTemplates(fc ResolvedFileCollection, data map[string]*tpldata, name string) (map[string]*tpldata, error) {
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
