package publicip

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"testing"

	"github.com/h2non/gock"
	"github.com/jorgesanchez-e/localenvironment/ddns/internal/adapters/publicip/ipify"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	ipv4test int = iota
	ipv6test
)

func Test_DDNS(t *testing.T) {
	tests := []struct {
		name        string
		setHTTPMock func()
		ipTestType  int
		expectedIP  net.IP
		expectedLog string
	}{
		{
			name: "ipv4-ok",
			setHTTPMock: func() {
				gock.New(ipify.IPv4URL).Get("/").Reply(http.StatusOK).JSON(map[string]string{"ip": "192.168.100.1"})
			},
			ipTestType: ipv4test,
			expectedIP: net.ParseIP("192.168.100.1"),
		},
		{
			name:        "ipv4-err",
			setHTTPMock: func() { gock.New(ipify.IPv4URL).Get("/").Reply(http.StatusInternalServerError) },
			ipTestType:  ipv4test,
			expectedIP:  nil,
			expectedLog: `time="[0-9:T-]+"\s+level=error msg="unable to get public ip, http code 500 from https:\/\/api.ipify.org\/\?format=json"`,
		},
		{
			name: "ipv6-ok",
			setHTTPMock: func() {
				gock.New(ipify.IPv6URL).Get("/").Reply(http.StatusOK).JSON(map[string]string{"ip": "2001:0db8:85a3:0000:0000:8a2e:0370:7334"})
			},
			ipTestType: ipv6test,
			expectedIP: net.ParseIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334"),
		},
		{
			name:        "ipv6-err",
			setHTTPMock: func() { gock.New(ipify.IPv6URL).Get("/").Reply(http.StatusInternalServerError) },
			ipTestType:  ipv6test,
			expectedIP:  nil,
			expectedLog: `time="[0-9:T-]+"\s+level=error msg="unable to get public ip, http code 500 from https:\/\/api6.ipify.org\/\?format=json"`,
		},
	}

	for _, tc := range tests {
		setHTTPMock := tc.setHTTPMock
		testType := tc.ipTestType
		expectedIP := tc.expectedIP
		expectedLog := tc.expectedLog
		client := New()

		t.Run(tc.name, func(t *testing.T) {
			defer gock.Off()
			setHTTPMock()

			originalOutput := logrus.StandardLogger().Out
			defer logrus.SetOutput(originalOutput)

			var log bytes.Buffer
			logrus.SetOutput(&log)

			ctx := context.Background()

			var ip net.IP
			switch testType {
			case ipv4test:
				ip = client.IPv4(ctx)
			case ipv6test:
				ip = client.IPv6(ctx)
			default:
				t.Fatal("invalid ip test type")
			}

			assert.Equal(t, expectedIP, ip)
			assert.Regexp(t, expectedLog, log.String())
		})
	}
}
