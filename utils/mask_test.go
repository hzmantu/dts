package utils

import (
	"fmt"
	"reflect"
	"testing"
)

func TestNumberRounding(t *testing.T) {
	got := NumberRounding("1.2345")
	want := "1.23"
	if !reflect.DeepEqual(got, want) {
		t.Errorf("excepted:%v, got:%v", want, got)
	}
}

func TestDateRounding(t *testing.T) {
	got := DateRounding("2001-02-03 04:05:06")
	want := "2001-02-03 04:00:00"
	if !reflect.DeepEqual(got, want) {
		t.Errorf("excepted:%v, got:%v", want, got)
	}
}

func TestMD5(t *testing.T) {
	got := MD5("123")
	want := "202cb962ac59075b964b07152d234b70"
	if !reflect.DeepEqual(got, want) {
		t.Errorf("excepted:%v, got:%v", want, got)
	}
}

func TestHmacSha256(t *testing.T) {
	got := HmacSha256("123")
	want := "a60d9233be50b494bb2a0912eb4c3581fcc75b27272a9600a693e79dae265dbc"
	if !reflect.DeepEqual(got, want) {
		t.Errorf("excepted:%v, got:%v", want, got)
	}
}

func TestSha1(t *testing.T) {
	got := Sha1("123")
	want := "40bd001563085fc35165329ea1ff5c5ecbdbbeef"
	if !reflect.DeepEqual(got, want) {
		t.Errorf("excepted:%v, got:%v", want, got)
	}
}

func TestSha256(t *testing.T) {
	got := Sha256("123")
	want := "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3"
	if !reflect.DeepEqual(got, want) {
		t.Errorf("excepted:%v, got:%v", want, got)
	}
}

func TestMobile(t *testing.T) {
	got := Mobile("13812345678")
	want := "138****5678"
	if !reflect.DeepEqual(got, want) {
		t.Errorf("excepted:%v, got:%v", want, got)
	}
}

func TestEmail(t *testing.T) {
	got := Email("623278977@qq.com")
	want := "7a89427280de0129ebb6f6d1ada1471c@qq.com"
	if !reflect.DeepEqual(got, want) {
		t.Errorf("excepted:%v, got:%v", want, got)
	}
}

func TestBankCard(t *testing.T) {
	got := BankCard("1234567890123456")
	want := "12345678*******6"
	if !reflect.DeepEqual(got, want) {
		t.Errorf("excepted:%v, got:%v", want, got)
	}
}

func TestIdCard(t *testing.T) {
	got := IdCard("32122120010203123X")
	want := "3212**********123X"
	if !reflect.DeepEqual(got, want) {
		t.Errorf("excepted:%v, got:%v", want, got)
	}
}

func TestCall(t *testing.T) {
	result, _ := Call("numberRounding", "1.2334")
	fmt.Println(result[0].String())
}
