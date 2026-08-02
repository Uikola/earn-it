package tasks

import (
	"strconv"
	"time"
)

func validateTaskTitle(title string) bool {
	return len(title) > 0
}

func validateDate(date string) bool {
	_, err := time.Parse("2006-01-02", date)
	return err == nil
}

func validateReward(reward string) bool {
	num, err := strconv.Atoi(reward)
	if err != nil {
		return false
	}
	return num > 0
}
