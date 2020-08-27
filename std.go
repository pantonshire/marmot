package marmot

func Std() map[string]interface{} {
  return map[string]interface{}{
    "add":      stdAdd,
    "mul":      stdMul,
    "mod":      stdMod,
    "signed":   stdSigned,
    "count":    stdCount,
    "interval": stdInterval,
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
