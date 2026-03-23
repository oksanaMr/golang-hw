package hw10programoptimization

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/goccy/go-json"
)

type User struct {
	ID       int
	Name     string
	Username string
	Email    string
	Phone    string
	Password string
	Address  string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	scanner := bufio.NewScanner(r)
	result := make(DomainStat)

	domainPattern := "\\." + domain
	re := regexp.MustCompile(domainPattern)

	for scanner.Scan() {
		var user User
		if err := json.Unmarshal(scanner.Bytes(), &user); err != nil {
			return nil, fmt.Errorf("unmarshal error: %w", err)
		}

		if re.MatchString(user.Email) {
			emailParts := strings.SplitN(user.Email, "@", 2)
			if len(emailParts) != 2 {
				continue
			}
			domain := strings.ToLower(emailParts[1])
			result[domain]++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan error: %w", err)
	}

	return result, nil
}
