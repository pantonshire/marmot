package marmot

import (
    "regexp"
    "strings"
)

var (
    reExtend  = regexp.MustCompile(`{{\s*extend\s[^{}]*}}`)
    reInclude = regexp.MustCompile(`{{\s*include\s[^{}]*}}`)
)

func parseDependencies(dependencyStatement string) []string {
    return strings.Fields(dependencyStatement[2 : len(dependencyStatement)-2])[1:]
}
