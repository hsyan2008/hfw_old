package hfw

import (
	"crypto/md5"
	"fmt"
)

type Result struct {
	ErrNo   int64       `json:"err_no"`
	ErrMsg  string      `json:"err_msg"`
	Results interface{} `json:"results"`
}

func CheckErr(err error) {
	if nil != err {
		panic(err)
	}
}

func Max(i int, j ...int) int {
	for _, v := range j {
		if v > i {
			i = v
		}
	}
	return i
}

func Min(i int, j ...int) int {
	for _, v := range j {
		if v < i {
			i = v
		}
	}
	return i
}

func Md5(str string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(str)))
}
