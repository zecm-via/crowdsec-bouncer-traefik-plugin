package crowdsec_bouncer_traefik_plugin //nolint:revive,stylecheck

import (
	"context"
	htmltemplate "html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"text/template"

	cache "github.com/zecm-via/crowdsec-bouncer-traefik-plugin/pkg/cache"
	captcha "github.com/zecm-via/crowdsec-bouncer-traefik-plugin/pkg/captcha"
	configuration "github.com/zecm-via/crowdsec-bouncer-traefik-plugin/pkg/configuration"
	ip "github.com/zecm-via/crowdsec-bouncer-traefik-plugin/pkg/ip"
	logger "github.com/zecm-via/crowdsec-bouncer-traefik-plugin/pkg/logger"
)

func TestServeHTTP(t *testing.T) {
	cfg := CreateConfig()
	cfg.CrowdsecLapiKey = "test"

	ctx := context.Background()
	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	handler, err := New(ctx, next, cfg, "demo-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)
}

func TestNew(t *testing.T) {
	type args struct {
		ctx    context.Context //nolint:containedctx
		next   http.Handler
		config *configuration.Config
		name   string
	}
	tests := []struct {
		name    string
		args    args
		want    http.Handler
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.ctx, tt.args.next, tt.args.config, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBouncer_ServeHTTP(t *testing.T) {
	type fields struct {
		next                   http.Handler
		name                   string
		template               *template.Template
		enabled                bool
		crowdsecScheme         string
		crowdsecHost           string
		crowdsecKey            string
		crowdsecMode           string
		updateInterval         int64
		defaultDecisionTimeout int64
		forwardedCustomHeader  string
		clientPoolStrategy     *ip.PoolStrategy
		serverPoolStrategy     *ip.PoolStrategy
		httpClient             *http.Client
		cacheClient            *cache.Client
	}
	type args struct {
		rw  http.ResponseWriter
		req *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			bouncer := &Bouncer{
				next:                   tt.fields.next,
				name:                   tt.fields.name,
				template:               tt.fields.template,
				enabled:                tt.fields.enabled,
				crowdsecScheme:         tt.fields.crowdsecScheme,
				crowdsecHost:           tt.fields.crowdsecHost,
				crowdsecKey:            tt.fields.crowdsecKey,
				crowdsecMode:           tt.fields.crowdsecMode,
				updateInterval:         tt.fields.updateInterval,
				defaultDecisionTimeout: tt.fields.defaultDecisionTimeout,
				forwardedCustomHeader:  tt.fields.forwardedCustomHeader,
				clientPoolStrategy:     tt.fields.clientPoolStrategy,
				serverPoolStrategy:     tt.fields.serverPoolStrategy,
				httpClient:             tt.fields.httpClient,
				cacheClient:            tt.fields.cacheClient,
			}
			bouncer.ServeHTTP(tt.args.rw, tt.args.req)
		})
	}
}

func Test_handleNoStreamCache(t *testing.T) {
	type args struct {
		bouncer  *Bouncer
		remoteIP string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := handleNoStreamCache(tt.args.bouncer, tt.args.remoteIP); (err != nil) != tt.wantErr {
				t.Errorf("handleNoStreamCache() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_handleStreamCache(t *testing.T) {
	type args struct {
		bouncer *Bouncer
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handleStreamCache(tt.args.bouncer)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleStreamCache() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_crowdsecQuery(t *testing.T) {
	type args struct {
		bouncer   *Bouncer
		stringURL string
		data      []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := crowdsecQuery(tt.args.bouncer, tt.args.stringURL, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("crowdsecQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("crowdsecQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleBanServeHTTPWithDifferentMethods(t *testing.T) {
	html := "<html>You are banned</html>"
	banTemplate, _ := htmltemplate.New("html").Parse(html)
	tests := []struct {
		name              string
		method            string
		banTemplate       *htmltemplate.Template
		expectBodyContent bool
	}{
		{
			name:              "GET request should have body with template",
			method:            http.MethodGet,
			banTemplate:       banTemplate,
			expectBodyContent: true,
		},
		{
			name:              "HEAD request should NOT have body even with template",
			method:            http.MethodHead,
			banTemplate:       banTemplate,
			expectBodyContent: false,
		},
		{
			name:              "POST request should have body with template",
			method:            http.MethodPost,
			banTemplate:       banTemplate,
			expectBodyContent: true,
		},
		{
			name:              "PUT request should have body with template",
			method:            http.MethodPut,
			banTemplate:       banTemplate,
			expectBodyContent: true,
		},
		{
			name:              "DELETE request should have body with template",
			method:            http.MethodDelete,
			banTemplate:       banTemplate,
			expectBodyContent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bouncer := &Bouncer{
				remediationStatusCode:   http.StatusForbidden,
				remediationCustomHeader: "X-Test-Remediation",
				banTemplate:             tt.banTemplate,
			}

			rw := httptest.NewRecorder()
			req := &http.Request{Method: tt.method}
			bouncer.handleBanServeHTTP(rw, req, "0.0.0.0", "TEST")

			// Check status code
			if rw.Code != http.StatusForbidden {
				t.Errorf("Expected status code 403, got %d", rw.Code)
			}

			// Check custom header
			headerValue := rw.Header().Get("X-Test-Remediation")
			if headerValue != "ban" {
				t.Errorf("Expected header X-Test-Remediation to be 'ban', got %s", headerValue)
			}

			// Check body content
			body := rw.Body.String()
			hasBodyContent := len(body) > 0

			if hasBodyContent != tt.expectBodyContent {
				t.Errorf("Method %s: expected body content: %v, got body content: %v (body: %q)",
					tt.method, tt.expectBodyContent, hasBodyContent, body)
			}

			// If we expect body content, verify it matches template
			if tt.expectBodyContent && body != html {
				t.Errorf("Expected body %q, got %q", html, body)
			}
		})
	}
}

func TestCaptchaMethodBasedLogic(t *testing.T) {
	tests := []struct {
		name              string
		method            string
		remediation       string
		expectBanFallback bool
	}{
		{
			name:              "GET with captcha remediation should allow captcha",
			method:            http.MethodGet,
			remediation:       cache.CaptchaValue,
			expectBanFallback: false,
		},
		{
			name:              "HEAD with captcha remediation should fallback to ban",
			method:            http.MethodHead,
			remediation:       cache.CaptchaValue,
			expectBanFallback: true,
		},
		{
			name:              "POST with captcha remediation should allow captcha",
			method:            http.MethodPost,
			remediation:       cache.CaptchaValue,
			expectBanFallback: false,
		},
		{
			name:              "PUT with captcha remediation should allow captcha",
			method:            http.MethodPut,
			remediation:       cache.CaptchaValue,
			expectBanFallback: false,
		},
		{
			name:              "DELETE with captcha remediation should allow captcha",
			method:            http.MethodDelete,
			remediation:       cache.CaptchaValue,
			expectBanFallback: false,
		},
		{
			name:              "PATCH with captcha remediation should allow captcha",
			method:            http.MethodPatch,
			remediation:       cache.CaptchaValue,
			expectBanFallback: false,
		},
		{
			name:              "OPTIONS with captcha remediation should allow captcha",
			method:            http.MethodOptions,
			remediation:       cache.CaptchaValue,
			expectBanFallback: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the core logic: captcha is served for all methods except HEAD
			shouldUseCaptcha := tt.remediation == cache.CaptchaValue && tt.method != http.MethodHead

			if shouldUseCaptcha == tt.expectBanFallback {
				t.Errorf("Method %s with %s remediation: expected ban fallback %v, but logic would use captcha %v",
					tt.method, tt.remediation, tt.expectBanFallback, shouldUseCaptcha)
			}
		})
	}
}

func Test_appsecQuery_captchaRemediation(t *testing.T) {
	appsecServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusForbidden)
		_, _ = rw.Write([]byte(`{"action":"captcha","http_status":403}`))
	}))
	defer appsecServer.Close()

	appsecURL, _ := url.Parse(appsecServer.URL)
	bouncer := &Bouncer{
		appsecScheme:     appsecURL.Scheme,
		appsecHost:       appsecURL.Host,
		appsecPath:       "/",
		httpAppsecClient: appsecServer.Client(),
		log:              logger.New("INFO", ""),
	}

	req, _ := http.NewRequest(http.MethodGet, "http://localhost/", nil)
	remediation, err := appsecQuery(bouncer, "1.2.3.4", req)
	if err == nil {
		t.Fatal("appsecQuery() expected an error for AppSec 403, got nil")
	}
	if remediation != cache.CaptchaValue {
		t.Fatalf("appsecQuery() remediation = %q, want %q", remediation, cache.CaptchaValue)
	}
}

func Test_appsecQuery_banRemediationFallback(t *testing.T) {
	appsecServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusForbidden)
		_, _ = rw.Write([]byte(`not-json`))
	}))
	defer appsecServer.Close()

	appsecURL, _ := url.Parse(appsecServer.URL)
	bouncer := &Bouncer{
		appsecScheme:     appsecURL.Scheme,
		appsecHost:       appsecURL.Host,
		appsecPath:       "/",
		httpAppsecClient: appsecServer.Client(),
		log:              logger.New("INFO", ""),
	}

	req, _ := http.NewRequest(http.MethodGet, "http://localhost/", nil)
	remediation, err := appsecQuery(bouncer, "1.2.3.4", req)
	if err == nil {
		t.Fatal("appsecQuery() expected an error for AppSec 403, got nil")
	}
	if remediation != cache.BannedValue {
		t.Fatalf("appsecQuery() remediation = %q, want %q", remediation, cache.BannedValue)
	}
}

func TestBouncer_setCaptchaSolvedAppsecHeader(t *testing.T) {
	cacheClient := &cache.Client{}
	cacheClient.New(logger.New("INFO", ""), false, "", "", "")
	cacheClient.Set("1.2.3.4_captcha", cache.CaptchaDoneValue, 60)

	bouncer := &Bouncer{
		captchaClient: &captcha.Client{},
		cacheClient:   cacheClient,
	}
	err := bouncer.captchaClient.New(
		logger.New("INFO", ""),
		cacheClient,
		http.DefaultClient,
		configuration.CustomProvider,
		"https://captcha.example.test/fast.js",
		"wicketkeeper",
		"wicketkeeper_solution",
		"https://captcha.example.test/v0/siteverify",
		"wicketkeeper",
		"wicketkeeper",
		"",
		"/captcha.html",
		60,
	)
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest(http.MethodPost, "http://localhost/", strings.NewReader("payload=keep"))
	bodyBefore := req.Body
	req.Header.Set(crowdsecCaptchaSolved, "spoofed")
	req.Header.Del(crowdsecCaptchaSolved)

	cleanup := bouncer.setCaptchaSolvedAppsecHeader(req, "1.2.3.4")
	if got := req.Header.Get(crowdsecCaptchaSolved); got != "1" {
		t.Fatalf("setCaptchaSolvedAppsecHeader() header = %q, want %q", got, "1")
	}
	if req.Body != bodyBefore {
		t.Fatal("setCaptchaSolvedAppsecHeader() changed request body")
	}

	cleanup()
	if got := req.Header.Get(crowdsecCaptchaSolved); got != "" {
		t.Fatalf("setCaptchaSolvedAppsecHeader() cleanup header = %q, want empty", got)
	}
	if req.Body != bodyBefore {
		t.Fatal("setCaptchaSolvedAppsecHeader() cleanup changed request body")
	}
}

func TestBouncer_handleNextServeHTTP_solvedCaptchaPreservesPostBody(t *testing.T) {
	cacheClient := &cache.Client{}
	cacheClient.New(logger.New("INFO", ""), false, "", "", "")
	cacheClient.Set("1.2.3.4_captcha", cache.CaptchaDoneValue, 60)

	var appsecBody string
	appsecServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get(crowdsecCaptchaSolved); got != "1" {
			t.Errorf("AppSec solved header = %q, want %q", got, "1")
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read AppSec body: %v", err)
		}
		appsecBody = string(body)
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte(`{"action":"allow","http_status":200}`))
	}))
	defer appsecServer.Close()

	appsecURL, _ := url.Parse(appsecServer.URL)

	var backendBody string
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get(crowdsecCaptchaSolved); got != "" {
			t.Errorf("backend received solved header = %q, want empty", got)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read backend body: %v", err)
		}
		backendBody = string(body)
	})

	bouncer := &Bouncer{
		next:             next,
		appsecEnabled:    true,
		appsecScheme:     appsecURL.Scheme,
		appsecHost:       appsecURL.Host,
		appsecPath:       "/",
		appsecBodyLimit:  1024,
		httpAppsecClient: appsecServer.Client(),
		cacheClient:      cacheClient,
		captchaClient:    &captcha.Client{},
		log:              logger.New("INFO", ""),
	}
	err := bouncer.captchaClient.New(
		logger.New("INFO", ""),
		cacheClient,
		http.DefaultClient,
		configuration.CustomProvider,
		"https://captcha.example.test/fast.js",
		"wicketkeeper",
		"wicketkeeper_solution",
		"https://captcha.example.test/v0/siteverify",
		"wicketkeeper",
		"wicketkeeper",
		"",
		"/captcha.html",
		60,
	)
	if err != nil {
		t.Fatal(err)
	}

	const postBody = "authenticity_token=abc&SAMLRequest=xyz"
	req := httptest.NewRequest(http.MethodPost, "http://localhost/users/auth/saml", strings.NewReader(postBody))
	rw := httptest.NewRecorder()

	bouncer.handleNextServeHTTP(rw, req, "1.2.3.4")

	if appsecBody != postBody {
		t.Fatalf("AppSec body = %q, want %q", appsecBody, postBody)
	}
	if backendBody != postBody {
		t.Fatalf("backend body = %q, want %q", backendBody, postBody)
	}
}

func TestBouncer_shouldHandleAppsecCaptchaValidation(t *testing.T) {
	cacheClient := &cache.Client{}
	cacheClient.New(logger.New("INFO", ""), false, "", "", "")

	bouncer := &Bouncer{
		captchaClient:             &captcha.Client{},
		cacheClient:               cacheClient,
		captchaGracePeriodSeconds: 60,
	}
	err := bouncer.captchaClient.New(
		logger.New("INFO", ""),
		cacheClient,
		http.DefaultClient,
		configuration.CustomProvider,
		"https://captcha.example.test/fast.js",
		"wicketkeeper",
		"wicketkeeper_solution",
		"https://captcha.example.test/v0/siteverify",
		"wicketkeeper",
		"wicketkeeper",
		"",
		"/captcha.html",
		60,
	)
	if err != nil {
		t.Fatal(err)
	}

	newCaptchaValidationRequest := func() *http.Request {
		body := "wicketkeeper_solution=test-token"
		req, _ := http.NewRequest(http.MethodPost, "http://localhost/", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return req
	}

	const remoteIP = "5.6.7.8"

	if bouncer.shouldHandleAppsecCaptchaValidation(newCaptchaValidationRequest(), remoteIP) {
		t.Fatal("shouldHandleAppsecCaptchaValidation() = true without pending marker, want false")
	}

	bouncer.setAppsecCaptchaPending(remoteIP)
	if !bouncer.shouldHandleAppsecCaptchaValidation(newCaptchaValidationRequest(), remoteIP) {
		t.Fatal("shouldHandleAppsecCaptchaValidation() = false with pending marker and captcha response, want true")
	}
}
