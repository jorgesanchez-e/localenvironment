package config

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
)

const validSimpleDDNSYAML = `
ddns:
  log-level: "debug"
  check-every-seconds: 300
  process-timeout-seconds: 20
  aws:
    - account-name: "example"
      region: "us-east-1"
      access-key: "1234567890"
      secret-key: "1234567890"
      zones:
        - record-name: vpn.example.net.
          record-type: A
          record-ttl: 3600
          record-value: 127.0.0.1
        - record-name: localhost.example.io.
          record-type: A
          record-ttl: 3600
          record-value: 127.0.0.1
        - record-name: test-a.example.dev.
          record-type: AAAA
          record-ttl: 3600
          record-value: ::1
`

func testConfigFromYAML(t *testing.T, raw string) *config {
	t.Helper()

	vp := viper.New()
	vp.SetConfigType("yaml")
	if err := vp.ReadConfig(strings.NewReader(raw)); err != nil {
		t.Fatalf("ReadConfig() error = %v", err)
	}

	return &config{vp: vp}
}

func TestGetSimpleDDNSConfig(t *testing.T) {
	t.Parallel()

	c := testConfigFromYAML(t, validSimpleDDNSYAML)

	got, err := c.GetSimpleDDNSConfig()
	if err != nil {
		t.Fatalf("GetSimpleDDNSConfig() error = %v", err)
	}

	if got.DDNS.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q", got.DDNS.LogLevel, "debug")
	}
	if got.DDNS.CheckEverySeconds != 300 {
		t.Errorf("CheckEverySeconds = %d, want %d", got.DDNS.CheckEverySeconds, 300)
	}
	if got.DDNS.UpdateTimeoutSeconds != 20 {
		t.Errorf("UpdateTimeoutSeconds = %d, want %d", got.DDNS.UpdateTimeoutSeconds, 20)
	}

	if len(got.DDNS.AWS) != 1 {
		t.Fatalf("AWS len = %d, want %d", len(got.DDNS.AWS), 1)
	}

	aws := got.DDNS.AWS[0]
	if aws.AccountName != "example" {
		t.Errorf("AccountName = %q, want %q", aws.AccountName, "example")
	}
	if aws.Region != "us-east-1" {
		t.Errorf("Region = %q, want %q", aws.Region, "us-east-1")
	}
	if aws.AccessKey != "1234567890" {
		t.Errorf("AccessKey = %q, want %q", aws.AccessKey, "1234567890")
	}
	if aws.SecretKey != "1234567890" {
		t.Errorf("SecretKey = %q, want %q", aws.SecretKey, "1234567890")
	}

	if len(aws.Zones) != 3 {
		t.Fatalf("Zones len = %d, want %d", len(aws.Zones), 3)
	}

	wantZones := []zoneConfig{
		{
			RecordName:  "vpn.example.net.",
			RecordType:  "A",
			RecordTTL:   3600,
			RecordValue: "127.0.0.1",
		},
		{
			RecordName:  "localhost.example.io.",
			RecordType:  "A",
			RecordTTL:   3600,
			RecordValue: "127.0.0.1",
		},
		{
			RecordName:  "test-a.example.dev.",
			RecordType:  "AAAA",
			RecordTTL:   3600,
			RecordValue: "::1",
		},
	}

	for i, want := range wantZones {
		if aws.Zones[i] != want {
			t.Errorf("Zones[%d] = %+v, want %+v", i, aws.Zones[i], want)
		}
	}
}

func TestGetSimpleDDNSConfig_MissingAccountName(t *testing.T) {
	t.Parallel()

	raw := strings.Replace(
		validSimpleDDNSYAML,
		`account-name: "example"`,
		`account-name: ""`,
		1,
	)

	c := testConfigFromYAML(t, raw)

	_, err := c.GetSimpleDDNSConfig()
	if err == nil {
		t.Fatal("GetSimpleDDNSConfig() error = nil, want validation error")
	}
}

func TestGetSimpleDDNSConfig_InvalidIPv4ForARecord(t *testing.T) {
	t.Parallel()

	raw := strings.Replace(
		validSimpleDDNSYAML,
		"record-value: 127.0.0.1",
		"record-value: ::1",
		1,
	)

	c := testConfigFromYAML(t, raw)

	_, err := c.GetSimpleDDNSConfig()
	if err == nil {
		t.Fatal("GetSimpleDDNSConfig() error = nil, want validation error")
	}
}

func TestGetSimpleDDNSConfig_InvalidIPv6ForAAAARecord(t *testing.T) {
	t.Parallel()

	raw := strings.Replace(
		validSimpleDDNSYAML,
		"record-value: ::1",
		"record-value: 127.0.0.1",
		1,
	)

	c := testConfigFromYAML(t, raw)

	_, err := c.GetSimpleDDNSConfig()
	if err == nil {
		t.Fatal("GetSimpleDDNSConfig() error = nil, want validation error")
	}
}

func TestGetSimpleDDNSConfig_InvalidIPAddress(t *testing.T) {
	t.Parallel()

	raw := strings.Replace(
		validSimpleDDNSYAML,
		"record-value: 127.0.0.1",
		"record-value: not-an-ip",
		1,
	)

	c := testConfigFromYAML(t, raw)

	_, err := c.GetSimpleDDNSConfig()
	if err == nil {
		t.Fatal("GetSimpleDDNSConfig() error = nil, want validation error")
	}
}
