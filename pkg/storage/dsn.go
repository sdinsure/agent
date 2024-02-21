package storage

import (
	"net/url"

	"github.com/google/go-querystring/query"
)

type DsnConfig struct {
	ConnectionTimeout int    `url:"connect_timeout"` // in second
	TimeZone          string `url:"timezone"`
}

var defaultOption = &DsnConfig{ConnectionTimeout: 10, TimeZone: "Asia/Taipei"}

type protocol interface {
	ProtocolPrefix() string
}

type stringProtocol string

var (
	Postgres stringProtocol = "postgres"
)

func (s stringProtocol) ProtocolPrefix() string {
	return string(s)
}

func NewDSN(p protocol, endpoint string, username, password string, dbName string, options *DsnConfig) (string, error) {
	var queryparams string
	if options == nil {
		queryparams = dSNParameter(defaultOption)
	} else {
		queryparams = dSNParameter(options)
	}

	u := &url.URL{
		Scheme:   p.ProtocolPrefix(),
		User:     url.UserPassword(username, password),
		Host:     endpoint,
		Path:     dbName,
		RawQuery: queryparams,
	}

	return u.String(), nil
}

func dSNParameter(options *DsnConfig) string {
	v, _ := query.Values(options)
	return v.Encode()
}
