package ipify

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetIP(t *testing.T) {
	tests := []struct {
		name                 string
		httpMockServer       *httptest.Server
		expectedIP           net.IP
		expectedErrorMessage string
	}{
		{
			name: "get-ipv4-ok",
			httpMockServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"ip":"192.168.100.1"}`))
			})),
			expectedIP:           net.ParseIP("192.168.100.1"),
			expectedErrorMessage: "",
		},
		{
			name: "get-ipv6-ok",
			httpMockServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"ip":"2001:0db8:85a3:0000:0000:8a2e:0370:7334"}`))
			})),
			expectedIP:           net.ParseIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334"),
			expectedErrorMessage: "",
		},
		{
			name: "invalid-ip",
			httpMockServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"ip":"this-is-not-an-ip"}`))
			})),
			expectedIP:           nil,
			expectedErrorMessage: "invalid ip value this-is-not-an-ip",
		},
		{
			name: "get-internal-server-error",
			httpMockServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			})),
			expectedIP:           nil,
			expectedErrorMessage: `unable to get public ip, http code 500 from http:\/\/[0-9\.]+:\d+`,
		},
		{
			name: "unmarshal-error",
			httpMockServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"ip":"2001:0db8:85a3:0000:0000:8a2e:0370:7334"`))
			})),
			expectedIP:           nil,
			expectedErrorMessage: `unable to get public ip, body unmarshal error: unexpected end of JSON input from http:\/\/[0-9\.]+:\d+`,
		},
	}

	for _, tc := range tests {
		httpMockServer := tc.httpMockServer
		expectedIP := tc.expectedIP
		expectedErrorMessage := tc.expectedErrorMessage

		t.Run(tc.name, func(t *testing.T) {
			defer httpMockServer.Close()

			ctx := context.Background()
			client := New()

			ip, err := client.GetIP(ctx, httpMockServer.URL)

			assert.Equal(t, expectedIP, ip)

			if err != nil {
				assert.Regexp(t, expectedErrorMessage, err.Error())
			}
		})
	}
}

// GetIP(ctx context.Context, url string) (_ net.IP, err error) {
