package panic_recovery

import (
	"errors"
	"fmt"
)

var ErrPanic = errors.New("panic recovered")

func SafeDivide(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w: %v", ErrPanic, r)
		}
	}()
	return a / b, nil
}

func MustParse(input string) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w: parsing %q: %v", ErrPanic, input, r)
		}
	}()
	return parse(input), nil
}

func parse(input string) int {
	if input == "" {
		panic("empty input")
	}
	n := 0
	for _, c := range input {
		if c < '0' || c > '9' {
			panic(fmt.Sprintf("invalid character: %c", c))
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func PanicIfNegative(n int) int {
	if n < 0 {
		panic("negative value not allowed")
	}
	return n * 2
}

func SafeExecute(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w: %v", ErrPanic, r)
		}
	}()
	fn()
	return nil
}
