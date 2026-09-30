package main

type Stats struct {
	Count         int
	Sum, Min, Max int64
}

func Calc(nums []int64) Stats {
	if len(nums) < 2 {
		return Stats{}
	}

	firstDiff := nums[1] - nums[0]
	sum := firstDiff
	minimum := firstDiff
	maximum := firstDiff

	for i := 2; i < len(nums); i++ {
		diff := nums[i] - nums[i-1]
		sum += diff
		if diff < minimum {
			minimum = diff
		}
		if diff > maximum {
			maximum = diff
		}
	}
	return Stats{
		Count: len(nums) - 1,
		Sum:   sum,
		Min:   minimum,
		Max:   maximum,
	}
}
