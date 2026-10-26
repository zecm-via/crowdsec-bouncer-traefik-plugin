# TU Berlin local changes

This fork carries TU-specific changes on top of the upstream Traefik CrowdSec
bouncer plugin.

Upstream project:

- `github.com/maxlerebourg/crowdsec-bouncer-traefik-plugin`

TU fork module path:

- `github.com/zecm-via/crowdsec-bouncer-traefik-plugin`

Target TU release base:

- upstream tag: `v1.8.1`
- upstream tag commit: `f9c85b1472d618e9188a025c9b2308d0b27e5321`
- TU release tag: `v1.8.1-tu.1`
- date: `2026-10-26`

Important:

- Do not create `v1.8.1-tu.1` from the current fork `main` branch.
- The fork `main` can contain commits newer than upstream `v1.8.1`.
- The TU release tag must be exactly:
  `upstream v1.8.1 + TU-001 patch`.

## TU-001: AppSec-triggered CAPTCHA support

Goal:

- Let CrowdSec AppSec return `captcha` remediation and have Traefik render the
  configured CAPTCHA page instead of a generic 403.
- After a visitor solves CAPTCHA, allow the same IP through for the configured
  CAPTCHA grace period.
- Let AppSec rules bypass only the intended anonymous GitLab CAPTCHA rules after
  CAPTCHA is solved, while keeping the rest of AppSec active.
- Prevent clients from spoofing the internal solved-CAPTCHA signal.

Changed files:

- `bouncer.go`
  - Adds internal header name `X-Crowdsec-Captcha-Solved`.
  - Removes any client-supplied `X-Crowdsec-Captcha-Solved` header before
    processing.
  - Uses upstream v1.8.1 structured AppSec responses and CAPTCHA rendering.
    Invalid responses still fail closed; upstream bot-detection challenges,
    response-size limits, HTTP version forwarding, and body-read error handling
    are preserved.
  - Records a short pending AppSec CAPTCHA marker so the CAPTCHA validation POST
    is handled by the plugin.
  - After successful CAPTCHA, temporarily adds `X-Crowdsec-Captcha-Solved: 1`
    only while calling AppSec, then removes it before forwarding to the backend.
  - Avoids cloning the request for solved-CAPTCHA AppSec checks because the
    cloned request shares the original request body; reading the clone consumed
    form POST bodies before they reached GitLab.
  - Skips CAPTCHA form inspection for unreadable streaming bodies and keeps
    upstream v1.8.1 AppSec unreadable-body handling intact:
    HTTP/2+/HTTP/3 streaming bodies without `Content-Length` are not read for
    AppSec, preventing false 403s/timeouts on long-lived streams.
- `pkg/captcha/captcha.go`
  - Adds `IsValidationRequest()` so the bouncer can detect the configured
    CAPTCHA response field without hardcoding provider-specific form names.
  - Preserves bytes consumed during form inspection, so an unrelated POST while
    CAPTCHA is pending reaches AppSec and the backend with its original body.
- `bouncer_test.go`
  - Adds coverage for AppSec CAPTCHA response parsing.
  - Adds coverage for non-JSON AppSec 403 falling back to ban.
  - Adds coverage for solved-CAPTCHA AppSec header injection and cleanup without
    changing the request body.
  - Adds coverage for the pending AppSec CAPTCHA validation path.
- `tu_appsec_test.go`
  - Exercises the complete CAPTCHA flow through `ServeHTTP`: spoofed header
    removal, pending validation, preservation of unrelated POST bodies,
    successful provider validation, trusted solved-header injection and cleanup,
    and continued enforcement of independent AppSec bans.
- `.golangci.yml` and `tests/e2e/mock/lib/`
  - Use the TU module path for linting and Traefik integration checks.
- `version.go`
  - Reports `v1.8.1-tu.1` to AppSec and LAPI.
- `go.mod`, `.traefik.yml`, and Go imports in test/package files
  - Change module/import path from upstream to the TU fork path:
    `github.com/zecm-via/crowdsec-bouncer-traefik-plugin`.

Required AppSec rule-side integration:

```apache
SecRule REQUEST_HEADERS:X-Crowdsec-Captcha-Solved "@streq 1" "id:500019,phase:1,pass,nolog,ctl:ruleRemoveById=500020,ctl:ruleRemoveById=500021"
```

This rule must be evaluated before the GitLab anonymous CAPTCHA rules. It should
only remove those CAPTCHA rules, not disable AppSec globally.

## Build the TU branch from exact upstream v1.8.1

Create the real TU release branch from the exact upstream tag:

```bash
git remote get-url upstream || git remote add upstream https://github.com/maxlerebourg/crowdsec-bouncer-traefik-plugin.git
git fetch upstream tag v1.8.1
git checkout -b tu/v1.8.1-appsec-captcha v1.8.1
```

## Validation and release workflow

Validate using Go 1.22.12, golangci-lint v1.63.4, Yaegi v0.16.1, and the pinned
Traefik version in `tests/e2e/mock/lib/common.sh`:

```bash
go test -cover ./...
golangci-lint run
make yaegi_test
make -k e2e_mock
git diff --check
```

Create the release base branch `tu/v1.8.1` at upstream `v1.8.1` and open a PR
from `tu/v1.8.1-appsec-captcha` into it. This keeps the release based on the
exact upstream tag independently of the fork's `main` branch. After checks
pass and the PR is merged, tag the merged release commit:

```bash
git fetch origin
git switch tu/v1.8.1
git pull --ff-only
git tag -a v1.8.1-tu.1 -m "Upstream v1.8.1 with TU AppSec CAPTCHA support"
git push origin v1.8.1-tu.1
```

Traefik static plugin config should then use:

```yaml
experimental:
  plugins:
    crowdsec-bouncer:
      moduleName: github.com/zecm-via/crowdsec-bouncer-traefik-plugin
      version: v1.8.1-tu.1
```

Use a new tag for every TU patch, for example:

- `v1.6.0-tu.1`
- `v1.6.0-tu.2`
- `v1.8.1-tu.1`

## Upstream sync workflow

Add the upstream remote once if it does not already exist:

```bash
git remote get-url upstream || git remote add upstream https://github.com/maxlerebourg/crowdsec-bouncer-traefik-plugin.git
git fetch upstream --tags
```

For each upstream release update, create the next TU branch from the upstream
release tag, not from fork `main`:

```bash
git fetch upstream tag vX.Y.Z
git checkout -b tu/vX.Y.Z-appsec-captcha vX.Y.Z
```

Then re-apply/cherry-pick the TU patch. During conflict resolution, keep:

- TU fork module path in `go.mod`.
- TU fork import path in Go files.
- `.traefik.yml` import path pointing to `github.com/zecm-via/...`.
- The AppSec CAPTCHA behavior documented in `TU-001`.

Update this document and `version.go`, run the validation commands above,
open a PR against the corresponding TU release base, and tag the merged commit.
Never move an existing release tag.
