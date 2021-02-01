package util

import (
	"math/rand"
	"net/url"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func ExtractURL(s string) (string, string, string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", "", "", err
	}

	host := strings.TrimRight(u.Hostname(), ".")
	sub := strings.IndexByte(host, '.')
	if sub != -1 && host[:sub] == "www" {
		host = host[sub+1:]
	}
	urlBase := host
	urlScheme := u.Scheme
	// TODO: remove port 80+443 if scheme is http/https?
	uri := strings.TrimRight(u.Host, ".") + strings.TrimRight(u.RequestURI(), "/")

	return uri, urlScheme, urlBase, nil
}

func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

func RandomUsername() string {
	return RandomString(6)
}

func RandomEmail() string {
	return RandomString(4) + "@" + RandomString(6) + ".com"
}

func RandomDomain() string {
	return "www." + RandomString(6) + ".com"
}

func RandomURI() string {
	return RandomDomain() + "/" + RandomString(10)
}
