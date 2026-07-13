package ipgetter

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type errorReadCloser struct {
	err error
}

func (e errorReadCloser) Read([]byte) (int, error) {
	return 0, e.err
}

func (e errorReadCloser) Close() error {
	return nil
}

func TestNewIPify(t *testing.T) {
	got := NewIPify()

	require.NotNil(t, got)
	assert.Equal(t, ipifyAPIURL, got.ipv4URL)
	assert.Equal(t, ipifyAPIURL6, got.ipv6URL)
	require.NotNil(t, got.ipifyClient)
}

func TestIPify_publicIP(t *testing.T) {
	testCases := []struct {
		name                string
		useNilContext       bool
		iType               ipType
		client              *http.Client
		ipv4URL             string
		ipv6URL             string
		expectedPublicIP    string
		expectedErrContains string
	}{
		{
			name:  "success IPv4",
			iType: ipTypeIPv4,
			client: &http.Client{
				Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBufferString("127.0.0.1")),
						Header:     make(http.Header),
					}, nil
				}),
			},
			expectedPublicIP: "127.0.0.1",
		},
		{
			name:  "success IPv6",
			iType: ipTypeIPv6,
			client: &http.Client{
				Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBufferString("::1")),
						Header:     make(http.Header),
					}, nil
				}),
			},
			expectedPublicIP: "::1",
		},
		{
			name:                "nil context",
			useNilContext:       true,
			iType:               ipTypeIPv4,
			ipv4URL:             "http://example.com",
			client:              &http.Client{},
			expectedErrContains: "failed to get public IPv4 IP",
		},
		{
			name:                "nil context IPv6",
			useNilContext:       true,
			iType:               ipTypeIPv6,
			ipv6URL:             "http://example.com",
			client:              &http.Client{},
			expectedErrContains: "failed to get public IPv6 IP",
		},
		{
			name:  "http client error IPv4",
			iType: ipTypeIPv4,
			client: &http.Client{
				Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
					return nil, errors.New("connection refused")
				}),
			},
			ipv4URL:             "http://example.com",
			expectedErrContains: "failed to get public IPv4 IP",
		},
		{
			name:  "http client error IPv6",
			iType: ipTypeIPv6,
			client: &http.Client{
				Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
					return nil, errors.New("connection refused")
				}),
			},
			ipv6URL:             "http://example.com",
			expectedErrContains: "failed to get public IPv6 IP",
		},
		{
			name:  "read body error IPv4",
			iType: ipTypeIPv4,
			client: &http.Client{
				Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       errorReadCloser{err: errors.New("read failed")},
						Header:     make(http.Header),
					}, nil
				}),
			},
			ipv4URL:             "http://example.com",
			expectedErrContains: "failed to get public IPv4 IP",
		},
		{
			name:                "invalid request URL",
			iType:               ipTypeIPv4,
			ipv4URL:             "://invalid-url",
			client:              &http.Client{},
			expectedErrContains: "failed to get public IPv4 IP",
		},
		{
			name:  "empty response body",
			iType: ipTypeIPv4,
			client: &http.Client{
				Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(http.NoBody),
						Header:     make(http.Header),
					}, nil
				}),
			},
			ipv4URL:          "http://example.com",
			expectedPublicIP: "",
		},
		{
			name:  "read body error IPv6",
			iType: ipTypeIPv6,
			client: &http.Client{
				Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       errorReadCloser{err: errors.New("read failed")},
						Header:     make(http.Header),
					}, nil
				}),
			},
			ipv6URL:             "http://example.com",
			expectedErrContains: "failed to get public IPv6 IP",
		},
	}

	for _, testCase := range testCases {
		ipifyClient := testCase.client
		ipv4URL := testCase.ipv4URL
		ipv6URL := testCase.ipv6URL
		ipType := testCase.iType
		expectedPublicIP := testCase.expectedPublicIP

		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			if testCase.useNilContext {
				ctx = nil
			}

			ipify := &ipify{
				ipv4URL:     ipv4URL,
				ipv6URL:     ipv6URL,
				ipifyClient: ipifyClient,
			}

			ip, err := ipify.publicIP(ctx, ipType)

			if testCase.expectedErrContains != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, testCase.expectedErrContains)
				assert.Empty(t, ip)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, expectedPublicIP, ip)
		})
	}
}
