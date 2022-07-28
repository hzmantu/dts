package utils

import "log"

func PanicError(err error) {
	if err != nil {
		panic(err)
	}
}

func CheckError(err error) bool {
	if err != nil {
		log.Println(err)
		return true
	}
	return false
}
