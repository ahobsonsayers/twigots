package twigots

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type Proxy struct {
	Host     string
	Port     int
	User     string
	Password string
}

// NewProxy creates a new Proxy instance with the given host, port, user, and password.
func NewProxy(host string, port int, user, password string) (*Proxy, error) {
	if host == "" {
		return nil, errors.New("host is required")
	}
	if port <= 0 {
		return nil, errors.New("port must be positive")
	}
	return &Proxy{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
	}, nil
}

// GenerateProxyList creates a list of Proxy instances from a list of proxy URLs.
func GenerateProxyList(proxyHosts []string, user, password string) ([]Proxy, error) {
	var proxyList []Proxy
	for _, p := range proxyHosts {
		parsedUrl, err := url.Parse(p)
		if err != nil {
			return nil, fmt.Errorf("failed to parse proxy URL '%s': %w", p, err)
		}
		if parsedUrl.Scheme != "socks5" {
			return nil, fmt.Errorf("unsupported proxy scheme for '%s': %s", p, parsedUrl.Scheme)
		}
		host, port, err := net.SplitHostPort(parsedUrl.Host)
		if err != nil {
			return nil, fmt.Errorf("failed to split host and port for '%s': %w", p, err)
		}
		portNum, err := strconv.Atoi(port)
		if err != nil {
			return nil, fmt.Errorf("invalid port number for '%s': %w", p, err)
		}
		proxyList = append(proxyList, Proxy{
			Host:     host,
			Port:     portNum,
			User:     user,
			Password: password,
		})
	}
	return proxyList, nil
}

// String returns the proxy URL as a string.
func (p *Proxy) String() (string, error) {
	parsedUrl, err := p.URL()
	if err != nil {
		return "", fmt.Errorf("failed to get proxy URL: %w", err)
	}
	return parsedUrl.String(), nil
}

// URL returns the proxy URL as a *url.URL instance.
func (p *Proxy) URL() (*url.URL, error) {
	if p.Host == "" || p.Port <= 0 {
		return nil, errors.New("host and port must be set for proxy URL")
	}

	var proxyUrl string

	if strings.TrimSpace(p.User) == "" || strings.TrimSpace(p.Password) == "" {
		proxyUrl = fmt.Sprintf("socks5://%s:%d", p.Host, p.Port)
	} else {
		proxyUrl = fmt.Sprintf(
			"socks5://%s:%s@%s:%d",
			p.User,
			p.Password,
			p.Host,
			p.Port,
		)
	}
	parsedUrl, err := url.Parse(proxyUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proxy URL: %w", err)
	}
	return parsedUrl, nil
}
