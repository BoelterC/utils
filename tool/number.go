package tool

import (
	"fmt"

	"github.com/BoelterC/utils/math"
)

func FibonacciQue(n int) []int {
	seque := make([]int, math.Max(2, n))
	seque[0] = 1
	seque[1] = 1

	for i := 2; i < len(seque); i++ {
		seque[i] = seque[i-1] + seque[i-2]
	}
	return seque[:n]
}

func Roman2Num(numeral string) int {
	romanMap := map[string]int{
		"M": 1000,
		"D": 500,
		"C": 100,
		"L": 50,
		"X": 10,
		"V": 5,
		"I": 1,
	}

	arabicVals := make([]int, len(numeral)+1)

	for index, digit := range numeral {
		if val, present := romanMap[string(digit)]; present {
			arabicVals[index] = val
		} else {
			fmt.Printf("Error: The roman numeral %s has a bad digit: %c\n", numeral, digit)
			return 0
		}
	}

	total := 0

	for index := 0; index < len(numeral); index++ {
		if arabicVals[index] < arabicVals[index+1] {
			arabicVals[index] = -arabicVals[index]
		}
		total += arabicVals[index]
	}

	return total
}
