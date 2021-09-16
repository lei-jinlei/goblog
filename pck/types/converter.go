package types

import (
	"goblog/pck/logger"
	"strconv"
)

func Int64ToString(num int64) string {
	return strconv.FormatInt(num, 10)
}

func StringToInt(str string) int {
	i, err := strconv.Atoi(str)
	if err != nil {
		logger.LogError(err)
	}
	return i
}

func Unit64ToString(num uint64) string {
	return strconv.FormatUint(num, 10)
}