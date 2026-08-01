package habits

import "strconv"

func validateHabitName(habitName string) bool {
	return len(habitName) > 0
}

func validateNumber(number string) bool {
	num, err := strconv.Atoi(number)
	if err != nil {
		return false
	}

	return num > 0
}
