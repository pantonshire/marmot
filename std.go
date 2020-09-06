package marmot

import (
  "fmt"
  "reflect"
)

// Returns the Marmot template function standard library. Pass into Cache.Functions to use.
//
// The standard library contains the following functions:
//  - (add a b): adds two ints
//  - (mul a b): multiplies two ints
//  - (mod a b): returns a mod b, where a and b are ints. The result will be in the range [0,b), even for negative values of a
//  - (signed n): converts a uint to an int
//  - (count n): returns an array of ints [0,n)
//  - (interval low high step): returns and array of ints [low, low+step, low+(2*step), ...], ending at the largest value less than high
//  - (strfy x): converts any value x to a string. If x is a pointer then it is dereferenced first
//  - (strfyf pattern x): like strfy, but formats the value according to the given pattern
func Std() FuncMap {
  return map[string]interface{}{
    "add":      stdAdd,
    "mul":      stdMul,
    "mod":      stdMod,
    "signed":   stdSigned,
    "count":    stdCount,
    "interval": stdInterval,
    "strfy":    stdStringify,
    "strfyf":   stdStringifyf,
  }
}

func stdAdd(a, b int) int {
  return a + b
}

func stdMul(a, b int) int {
  return a * b
}

func stdMod(a, b int) int {
  return ((a % b) + b) % b
}

func stdSigned(n uint) int {
  return int(n)
}

func stdCount(n int) []int {
  if n <= 0 {
    return nil
  }
  ns := make([]int, n)
  for i := 0; i < n; i++ {
    ns[i] = i
  }
  return ns
}

func stdInterval(low, high, step int) []int {
  n := ((high - low) + step - 1) / step
  if n <= 0 {
    return []int{}
  }
  ns := make([]int, n)
  for i := 0; i < n; i++ {
    ns[i] = low + (i * step)
  }
  return ns
}

func stdStringify(x interface{}) string {
  return stdStringifyf("%v", x)
}

func stdStringifyf(format string, x interface{}) string {
  if x != nil {
    r := reflect.ValueOf(x)
    for r.Kind() == reflect.Ptr {
      r = r.Elem()
    }
    x = r
  }
  return fmt.Sprintf(format, x)
}
