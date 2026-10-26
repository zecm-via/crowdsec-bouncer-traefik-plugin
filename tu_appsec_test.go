package crowdsec_bouncer_traefik_plugin //nolint:revive,stylecheck

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/zecm-via/crowdsec-bouncer-traefik-plugin/pkg/configuration"
)

//nolint:gocyclo // Keep the successive requests of the CAPTCHA lifecycle in one regression test.
func TestTUAppsecCaptchaFlow(t *testing.T) {
	const remoteIP = "192.0.2.80"
	var appsecBody, backendBody, solvedHeader string
	appsecCalls, validations, backendCalls := 0, 0, 0
	appsec := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/siteverify" {
			validations++
			if r.FormValue("response") != "valid-token" {
				t.Error("CAPTCHA provider did not receive the validation token")
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"success":true}`)
			return
		}
		appsecCalls++
		solvedHeader = r.Header.Get(crowdsecCaptchaSolved)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		appsecBody = string(body)
		uri := r.Header.Get(crowdsecAppsecURIHeader)
		switch {
		case strings.Contains(uri, "/ban"):
			w.WriteHeader(http.StatusForbidden)
			_, _ = io.WriteString(w, `{"action":"ban","http_status":403}`)
		case strings.Contains(uri, "/allowed"), solvedHeader == "1":
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusForbidden)
			_, _ = io.WriteString(w, `{"action":"captcha","http_status":403}`)
		}
	}))
	defer appsec.Close()

	endpoint, err := url.Parse(appsec.URL)
	if err != nil {
		t.Fatal(err)
	}
	cfg := CreateConfig()
	cfg.Enabled = true
	cfg.CrowdsecMode = configuration.AppsecMode
	cfg.CrowdsecLapiKey = "test-key"
	cfg.CrowdsecAppsecEnabled = true
	cfg.CrowdsecAppsecScheme = endpoint.Scheme
	cfg.CrowdsecAppsecHost = endpoint.Host
	cfg.CaptchaProvider = configuration.CustomProvider
	cfg.CaptchaCustomJsURL = appsec.URL + "/captcha.js"
	cfg.CaptchaCustomKey = "wicketkeeper"
	cfg.CaptchaCustomResponse = "wicketkeeper_solution"
	cfg.CaptchaCustomValidateURL = appsec.URL + "/siteverify"
	cfg.CaptchaSiteKey = "test-site"
	cfg.CaptchaSecretKey = "test-secret"
	cfg.CaptchaFilePath = "captcha.html"
	cfg.RemediationHeadersCustomName = "X-Remediation"
	handler, err := New(context.Background(), http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		backendCalls++
		if r.Header.Get(crowdsecCaptchaSolved) != "" {
			t.Error("internal solved-CAPTCHA header leaked to the backend")
		}
		body, readErr := io.ReadAll(r.Body)
		if readErr != nil {
			t.Error(readErr)
		}
		backendBody = string(body)
	}), cfg, "tu-appsec-test")
	if err != nil {
		t.Fatal(err)
	}
	bouncer, ok := handler.(*Bouncer)
	if !ok {
		t.Fatal("New did not return a Bouncer")
	}
	bouncer.cacheClient.Delete(remoteIP + "_captcha")
	bouncer.cacheClient.Delete(appsecCaptchaPendingKey(remoteIP))
	t.Cleanup(func() {
		bouncer.cacheClient.Delete(remoteIP + "_captcha")
		bouncer.cacheClient.Delete(appsecCaptchaPendingKey(remoteIP))
	})
	request := func(method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, "http://gitlab.example"+path, strings.NewReader(body))
		req.RemoteAddr = remoteIP + ":1234"
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set(crowdsecCaptchaSolved, "spoofed")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		return recorder
	}

	response := request(http.MethodGet, "/protected", "")
	if response.Header().Get("X-Remediation") != "captcha" || solvedHeader != "" || backendCalls != 0 {
		t.Fatalf("spoofed header bypassed CAPTCHA: status=%d, header=%q, backend calls=%d", response.Code, solvedHeader, backendCalls)
	}
	if _, pendingErr := bouncer.cacheClient.Get(appsecCaptchaPendingKey(remoteIP)); pendingErr != nil {
		t.Fatal("AppSec CAPTCHA did not record a pending validation")
	}

	const postBody = "authenticity_token=abc&SAMLRequest=xyz"
	request(http.MethodPost, "/allowed", postBody)
	if appsecBody != postBody || backendBody != postBody {
		t.Errorf("pending CAPTCHA consumed an unrelated POST: AppSec=%q, backend=%q", appsecBody, backendBody)
	}
	callsBeforeValidation := appsecCalls
	response = request(http.MethodPost, "/protected", "wicketkeeper_solution=valid-token")
	if response.Code != http.StatusFound || validations != 1 || appsecCalls != callsBeforeValidation {
		t.Fatalf("CAPTCHA validation failed: status=%d, validations=%d, AppSec calls=%d", response.Code, validations, appsecCalls)
	}
	if _, pendingErr := bouncer.cacheClient.Get(appsecCaptchaPendingKey(remoteIP)); pendingErr == nil {
		t.Error("successful validation left the pending marker behind")
	}
	request(http.MethodPost, "/protected", postBody)
	if solvedHeader != "1" || appsecBody != postBody || backendBody != postBody {
		t.Fatalf("solved CAPTCHA lost the trusted signal or POST body: header=%q, AppSec=%q, backend=%q", solvedHeader, appsecBody, backendBody)
	}
	callsBeforeBan := backendCalls
	response = request(http.MethodGet, "/ban", "")
	if response.Code != http.StatusForbidden || backendCalls != callsBeforeBan {
		t.Fatal("solving CAPTCHA bypassed an independent AppSec ban")
	}
}
