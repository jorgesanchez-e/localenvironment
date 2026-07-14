package ipgetter

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

const (
	ipifyAPIURL  = "https://api.ipify.org"
	ipifyAPIURL6 = "https://api6.ipify.org"
)

type ipify struct {
	ipv4URL     string
	ipv6URL     string
	ipifyClient *http.Client
}

func NewIPify() *ipify {
	return &ipify{
		ipv4URL:     ipifyAPIURL,
		ipv6URL:     ipifyAPIURL6,
		ipifyClient: &http.Client{},
	}
}

func (i *ipify) publicIP(ctx context.Context, iType ipType) (_ string, err error) {
	url := i.ipv6URL
	if iType == ipTypeIPv4 {
		url = i.ipv4URL
	}

	defer func() {
		if err != nil {
			if iType == ipTypeIPv4 {
				err = fmt.Errorf("failed to get public IPv4 IP: %w", err)
			} else {
				err = fmt.Errorf("failed to get public IPv6 IP: %w", err)
			}
		}
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := i.ipifyClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
