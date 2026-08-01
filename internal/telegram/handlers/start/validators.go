package start

import (
	"strconv"
	"time"
)

func validateTimezone(tz string) bool {
	_, err := time.LoadLocation(tz)
	return err == nil
}

func validateNumber(number string) bool {
	num, err := strconv.Atoi(number)
	if err != nil {
		return false
	}

	return num > 0
}
