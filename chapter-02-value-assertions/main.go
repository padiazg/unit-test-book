package value_assertions

func ParseInt(input string) int {
	if len(input) == 0 {
		return 0
	}

	var (
		result int
		sign   = 1
		start  = 0
	)

	if input[0] == '-' {
		sign = -1
		start = 1
	} else if input[0] == '+' {
		start = 1
	}

	for i := start; i < len(input); i++ {
		ch := input[i]
		if ch < '0' || ch > '9' {
			return 0
		}
		result = result*10 + int(ch-'0')
	}

	return result * sign
}
