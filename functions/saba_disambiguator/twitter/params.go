package twitter2

import (
	"net/url"
	"strings"
)

type Params struct {
	url.Values
}

func newParams() Params {
	return Params{make(url.Values)}
}

func (p Params) Set(key string, value ...string) {
	p.Values.Set(key, strings.Join(value, ","))
}

func (p Params) Encode() string {
	// A space should be escaped into '%20' instead of '+' on twitter's query parameter.
	return strings.Replace(p.Values.Encode(), "+", "%20", -1)
}
