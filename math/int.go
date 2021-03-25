package math

const MaxUint = ^uint(0)
const MinUint = 0
const MaxInt = int(MaxUint >> 1)
const MinInt = -MaxInt - 1

func Min(values ...int) int {
	minVal := MaxInt
	for _, val := range values {
		if val < minVal {
			minVal = val
		}
	}
	return minVal
}

func Max(values ...int) int {
	maxVal := MinInt
	for _, val := range values {
		if val > maxVal {
			maxVal = val
		}
	}
	return maxVal
}
