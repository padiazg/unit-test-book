package error_message_verification

import (
	"fmt"
	"regexp"
	"strings"
)

type EmailAddress struct {
	Local  string
	Domain string
	Raw    string
}

func ParseEmail(raw string) (*EmailAddress, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("email address is empty")
	}

	parts := strings.Split(raw, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("email %q must contain exactly one @ symbol", raw)
	}

	local := parts[0]
	domain := parts[1]

	if local == "" {
		return nil, fmt.Errorf("email %q has empty local part", raw)
	}
	if domain == "" {
		return nil, fmt.Errorf("email %q has empty domain", raw)
	}

	if len(local) > 64 {
		return nil, fmt.Errorf("email %q local part exceeds 64 characters", raw)
	}

	if !strings.Contains(domain, ".") {
		return nil, fmt.Errorf("email %q domain %q must contain a dot", raw, domain)
	}

	re := regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+$`)
	if !re.MatchString(local) {
		return nil, fmt.Errorf("email %q contains invalid characters in local part", raw)
	}

	return &EmailAddress{Local: local, Domain: domain, Raw: raw}, nil
}
