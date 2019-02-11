package main

import "math"

func Average(nums []float64) float64 {
	sum := float64(0)
	for _, n := range nums {
		sum += n
	}
	result := sum / float64(len(nums))
	if result == math.NaN() {
		return 0
	}
	return result
}

func Min(nums []float64) float64 {
	if len(nums) == 0 {
		return 0
	}
	min := nums[0]
	for _, n := range nums {
		if n < min {
			min = n
		}
	}
	return min
}
func Max(nums []float64) float64 {
	if len(nums) == 0 {
		return 0
	}
	max := nums[0]
	for _, n := range nums {
		if n > max {
			max = n
		}
	}
	return max
}

func AboveBy(ms float64, nums []float64) float64 {
	abovecount := 0
	for _, n := range nums {
		if n >= ms {
			abovecount++
		}
	}
	result := float64(abovecount) / float64(len(nums)) * 100
	if result == math.NaN() {
		return 0
	}
	return result
}
