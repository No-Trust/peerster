package common

/**
 * Returns the given float32 clamped to the given bounds.
 */
func ClampFloat32(value, min, max float32) float32 {

	if value < min {

		return min

	} else if value > max {

		return max

	} else {

		return value

	}

}

/**
 * Returns the absolute value of the given float32.
 */
func AbsFloat32(x float32) float32 {

	if x < 0 {

		return -x

	} else if x == 0 {

		return 0

	} else {

		return x

	}

}
