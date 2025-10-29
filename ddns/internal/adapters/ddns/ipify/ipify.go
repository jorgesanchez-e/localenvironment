package ipify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/go-playground/validator/v10"
)

const (
	IPv4URL string = "https://api.ipify.org/?format=json"
	IPv6URL string = "https://api6.ipify.org/?format=json"
)

type ipifyAddress struct {
	IP string `json:"ip"`
}

type Client struct{}

func New() Client {
	return Client{}
}

func (c Client) GetIP(ctx context.Context, url string) (_ net.IP, err error) {
	var req *http.Request

	if req, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil); err != nil {
		return nil, err
	}

	client := &http.Client{}

	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to get public ip, http code %d from %s", resp.StatusCode, url)
	}

	var body []byte
	if body, err = io.ReadAll(resp.Body); err != nil {
		return nil, fmt.Errorf("unable to get public ip, body read error: %w from %s", err, url)
	}

	address := ipifyAddress{}
	if err = json.Unmarshal(body, &address); err != nil {
		return nil, fmt.Errorf("unable to get public ip, body unmarshal error: %w from %s", err, url)
	}

	validate := validator.New()
	if err = validate.Var(address.IP, "required,ipv4"); err == nil {
		return net.ParseIP(address.IP), nil
	}

	if err = validate.Var(address.IP, "required,ipv6"); err == nil {
		return net.ParseIP(address.IP), nil
	}

	return nil, fmt.Errorf("invalid ip value %s", address.IP)
}
