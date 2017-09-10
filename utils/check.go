package utils

import (
	"os"
	"errors"
)
func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func CheckNotExists(e error) {
	if e == nil {
		return
	}

	if os.IsNotExist(e) {
		panic(errors.New("File does not exist"))
	}
	
	panic(e)
}
