package transform

import (
	"code.hzmantu.com/dts/utils"
	"strconv"
)

func StringToInt(in string) int {
	out, err := strconv.Atoi(in)
	utils.PanicError(err)
	return out
}

func StringToInt64(in string) int64 {
	if in == "" {
		return 0
	}
	out, err := strconv.ParseInt(in, 10, 64)
	utils.PanicError(err)
	return out
}
