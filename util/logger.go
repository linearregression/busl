package util

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const prefix = "busl"
const defaultDomain = "development"

func init() {
	log.SetPrefix(fmt.Sprintf("busl source=%s pid=%v ", buslSource(envOrDefaultString("DOMAIN", defaultDomain)), os.Getpid()))
	log.SetFlags(0)
}

func envOrDefaultString(envvar, envdefault string) string {
	var v string
	if e := os.Getenv(envvar); e != "" {
		v = e
	} else {
		v = envdefault
	}

	return v
}

func environment(domain string) string {
	if domain == defaultDomain {
		return domain
	}

	// reverse the domain
	s := strings.Split(domain, ".")
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	s = append(s, prefix, strconv.Itoa(os.Getpid()))
	return strings.Join(s, ".")
}

func buslSource(domain string) string {
	parts := []string{environment(domain)}
	return strings.Join(parts, ".")
}
