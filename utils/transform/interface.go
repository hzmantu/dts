package transform

import (
	"code.hzmantu.com/dts/utils"
	"errors"
	"strconv"
)

func InterfaceToString(in interface{}) string {
	switch in.(type) {
	case string:
		return in.(string)
	case int32:
		return strconv.Itoa(int(in.(int32)))
	case int64:
		return strconv.Itoa(int(in.(int64)))
	case int:
		return strconv.Itoa(in.(int))
	case uint32:
		return strconv.Itoa(int(in.(uint32)))
	case uint64:
		return strconv.Itoa(int(in.(uint64)))
	case float64:
		return strconv.Itoa(int(in.(float64)))
	case float32:
		return strconv.Itoa(int(in.(float32)))
	case int8:
		return strconv.Itoa(int(in.(int8)))
	case int16:
		return strconv.Itoa(int(in.(int16)))
	case uint:
		return strconv.Itoa(int(in.(uint)))
	case []uint8:
		return string(in.([]uint8))
	default:
		utils.PanicError(errors.New("InterfaceToString Error"))
		return ""
	}
}