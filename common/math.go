package common

/**
 * Returns the given value clamped to the given bounds.
 */
func Clamp(value, min, max float32) float32 {

  if (value < min) {
    return min
  } else if (value > max) {
    return max
  } else {
    return value
  }

}
