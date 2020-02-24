package proxy

import (
	"os"
	"reflect"
	"testing"
	"time"

	"golang.org/x/xerrors"
)

func assertEq(want, have interface{}, t *testing.T) {
	if !reflect.DeepEqual(want, have) {
		t.Errorf("want: %#v", want)
		t.Errorf("have: %#v", have)
		t.Errorf("expected values to be equal")
	}
}

func TestDefaultConfiguration(t *testing.T) {
	want := DefaultProxyConfig()
	have, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected err loading config: %v", err)
	}
	assertEq(want, have, t)
}

func TestEnvironmentOverridesConfiguration(t *testing.T) {
	testCases := []struct {
		Name         string
		EnvOverrides map[string]string
		CheckFunc    func(c Configuration, t *testing.T)
	}{
		{
			Name: "Test Session Cookie Config Name Overrides",
			EnvOverrides: map[string]string{
				"SESSION_COOKIE_NAME": "foo_cookie_name",
			},
			CheckFunc: func(c Configuration, t *testing.T) {
				assertEq("foo_cookie_name", c.SessionConfig.CookieConfig.Name, t)
			},
		},
		{
			Name: "Test Request Timeout Overrides",
			EnvOverrides: map[string]string{
				"SERVER_TIMEOUT_WRITE": "60s",
				"SERVER_TIMEOUT_READ":  "60s",
			},
			CheckFunc: func(c Configuration, t *testing.T) {
				assertEq(60*time.Second, c.ServerConfig.TimeoutConfig.Write, t)
				assertEq(60*time.Second, c.ServerConfig.TimeoutConfig.Read, t)
			},
		},
		{
			Name: "Test Provider Config Overrides",
			EnvOverrides: map[string]string{
				"PROVIDER_SLUG": "foo-slug",
			},
			CheckFunc: func(c Configuration, t *testing.T) {
				assertEq("foo-slug", c.ProviderConfig.ProviderSlug, t)
			},
		},
		//{
		//	Name: "some upstream config tests?",
		//},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tc.EnvOverrides {
				err := os.Setenv(k, v)
				if err != nil {
					t.Fatalf("unexpected err setting env: %v", err)
				}
			}
			have, err := LoadConfig()
			if err != nil {
				t.Fatalf("unexpected err loading config: %v", err)
			}
			tc.CheckFunc(have, t)
		})
	}
}

func TestConfigValidate(t *testing.T) {
	testCases := []struct {
		Name        string
		Validator   Validator
		ExpectedErr error
	}{
		{
			Name: "config validation should pass",
			Validator: Configuration{
				ServerConfig: ServerConfig{
					Port: 4180,
					TimeoutConfig: TimeoutConfig{
						Write:    30 * time.Second,
						Read:     30 * time.Second,
						Shutdown: 30 * time.Second,
					},
				},
				ProviderConfig: ProviderConfig{
					ProviderType:        "sso",
					ProviderSlug:        "google",
					Scope:               "test",
					ProviderURLExternal: "https://sso-external.com",
					ProviderURLInternal: "https://sso-internal.com",
				},
				ClientConfig: ClientConfig{
					ID:     "foo-id",
					Secret: "bar-secret",
				},
				SessionConfig: SessionConfig{
					CookieConfig: CookieConfig{
						Name:     "_sso_proxy",
						Expire:   168 * time.Hour,
						Secure:   true,
						Secret:   "SMoPfinxNz0fuaGFPUr5vwQpvoG+CGcLd2nkxRVI+H4=",
						HTTPOnly: true,
					},
					TTLConfig: TTLConfig{
						Lifetime:    720 * time.Hour,
						Valid:       60 * time.Second,
						GracePeriod: 3 * time.Hour,
					},
				},
				UpstreamConfigs: UpstreamConfigs{
					DefaultConfig: DefaultConfig{
						Timeout:       10 * time.Second,
						ResetDeadline: 60 * time.Second,
					},
					Scheme:  "https",
					Cluster: "foo-cluster",
				},
				LoggingConfig: LoggingConfig{
					Enable: true,
					Level:  "INFO",
				},
				MetricsConfig: MetricsConfig{
					StatsdConfig: StatsdConfig{
						Port: 8125,
						Host: "localhost",
					},
				},
			},
		},
		{
			Name: "missing server.port configuration",
			Validator: Configuration{
				ServerConfig: ServerConfig{
					TimeoutConfig: TimeoutConfig{
						Write:    30 * time.Second,
						Read:     30 * time.Second,
						Shutdown: 30 * time.Second,
					},
				},
			},
			ExpectedErr: xerrors.New("invalid server config: no server.port configured"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Validator.Validate()
			if err != nil && tc.ExpectedErr != nil {
				assertEq(tc.ExpectedErr.Error(), err.Error(), t)
			} else {
				assertEq(tc.ExpectedErr, err, t)
			}
		})
	}
}
