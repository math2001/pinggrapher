package main

func Average(nums []float64) float64 {
	sum := 0.0
	for _, n := range nums {
		sum += n
	}
	return sum / float64(len(nums))
}

func Min(nums []float64) float64 {
	min := nums[0]
	for _, n := range nums {
		if n < min {
			min = n
		}
	}
	return min
}
func Max(nums []float64) float64 {
	max := nums[0]
	for _, n := range nums {
		if n > max {
			max = n
		}
	}
	return max
}
