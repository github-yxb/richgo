package utils

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"runtime"
)

func Protect(f func()) {
	defer func() {
		if r := recover(); r != nil {

			buf := make([]byte, 2048)
			n := runtime.Stack(buf, false)
			stackInfo := fmt.Sprintf("%s", buf[:n])

			logrus.Errorf("function panicing:p, %s", r, stackInfo)
		}
	}()

	f()
}
