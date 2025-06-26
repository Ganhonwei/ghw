package libs

import (
	"fmt"
	"testing"
)

func TestPasswor(t *testing.T) {
	pwd := Md5([]byte("maya778899" + "J7X2MJGS4N"))
	fmt.Println(pwd)

	// 91fe61e72c8bb5d1d9e044f4eda50a93
	// e01b3753997dd7d6646d754137bc4c92
}
