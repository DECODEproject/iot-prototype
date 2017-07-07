package utils

import (
	"net/url"
	"strconv"
	"strings"
)

// HostAndIpToBits parses an address parsing out the host and port
// Example input - http://abc.com:1234 returns { true, abc.com, 1234 }
func HostAndIpToBits(address string) (bool, string, int) {

	url, err := url.Parse(address)

	if err != nil {
		return false, "", 0
	}

	bits := strings.Split(url.Host, ":")

	if len(bits) != 2 {
		return false, "", 0
	}

	port, err := strconv.Atoi(bits[1])

	if err != nil {
		return false, "", 0
	}

	return true, bits[0], port
}
