package utils

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type StringFilter func(string) string

var funcMap = make(map[string]StringFilter)

func init() {
	RegFunc("NumberRounding", NumberRounding)
	RegFunc("DateRounding", DateRounding)
	RegFunc("MD5", MD5)
	RegFunc("SHA1", Sha1)
	RegFunc("SHA256", Sha256)
	RegFunc("HmacSha256", HmacSha256)
	RegFunc("Mobile", Mobile)
	RegFunc("Email", Email)
	RegFunc("BankCard", BankCard)
	RegFunc("IdCard", IdCard)
}

func Call(name string, params ...interface{}) (result []reflect.Value, err error) {
	f := reflect.ValueOf(funcMap[strings.ToUpper(name)])
	if len(params) != f.Type().NumIn() {
		err = errors.New("the number of params is not adapted")
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = f.Call(in)
	return
}

func CallString(funcName string, str string) string {
	f, exist := funcMap[strings.ToUpper(funcName)]
	if !exist {
		return str
	}
	return f(str)
}

func RegFunc(name string, fc StringFilter) {
	funcMap[strings.ToUpper(name)] = fc
}

func NumberRounding(str string) string {
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		PanicError(err)
	}
	return fmt.Sprintf("%.2f", f)
}

func DateRounding(str string) string {
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", str, time.Local)
	return t.Format("2006-01-02 15:00:00")
}

func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func HmacSha256(data string) string {
	key := "dts"
	hash := hmac.New(sha256.New, []byte(key))
	hash.Write([]byte(data))
	return hex.EncodeToString(hash.Sum(nil))
}

func Sha1(data string) string {
	sha1 := sha1.New()
	sha1.Write([]byte(data))
	return hex.EncodeToString(sha1.Sum(nil))
}

func Sha256(data string) string {
	sha256 := sha256.New()
	sha256.Write([]byte(data))
	return hex.EncodeToString(sha256.Sum(nil))
}

func Mobile(str string) string {
	if len(str) == 11 {
		return str[0:3] + "****" + str[7:]
	} else {
		return str
	}
}

func Email(str string) string {
	i := strings.Index(str, "@")
	if i == -1 {
		i = len(str)
	}
	h := md5.New()
	h.Write([]byte(str[0:i]))
	b := hex.EncodeToString(h.Sum(nil)) + str[i:]
	return b
}

func BankCard(str string) string {
	if len(str) == 16 || len(str) == 19 {
		return str[0:8] + "*******" + str[15:]
	} else {
		return str
	}
}

func IdCard(str string) string {
	if len(str) == 15 || len(str) == 18 {
		return str[0:4] + "**********" + str[14:]
	} else {
		return str
	}
}
