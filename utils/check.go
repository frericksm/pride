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


func TestEq(a, b []string) bool {

    if a == nil && b == nil { 
        return true; 
    }

    if a == nil || b == nil { 
        return false; 
    }

    if len(a) != len(b) {
        return false
    }

    for i := range a {
        if a[i] != b[i] {
            return false
        }
    }

    return true
}
