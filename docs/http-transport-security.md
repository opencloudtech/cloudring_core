# HTTP transport security audit

`cloudring-httpcheck` is a provider-neutral, read-only probe for a public HTTP
surface. It checks the redirect boundary, the direct HTTPS response, TLS, and a
small response-header policy without following redirects or exposing response
content in its report. Provider repositories supply their own target URLs and
run the probe during deployment promotion.

This contract follows the method-preserving `307` and `308` semantics in
[RFC 9110](https://www.rfc-editor.org/rfc/rfc9110.html), the HSTS processing
model in [RFC 6797](https://www.rfc-editor.org/rfc/rfc6797.html), and the
browser isolation directives defined by the
[Content Security Policy specification](https://www.w3.org/TR/CSP/).

## Probe model

One invocation checks one canonical HTTPS URL and one declared surface:

```bash
go run ./cmd/cloudring-httpcheck check \
  --target-id public-browser \
  --url 'https://public.example.test/ready?probe=transport' \
  --mode canary \
  --surface browser
```

The probe performs two `GET` requests:

1. It derives the HTTP URL by changing only the scheme. Redirect following is
   disabled. Canary mode requires exactly `307`; steady mode requires exactly
   `308`. `Location` must be an absolute HTTPS URL with the same authority,
   host, port, escaped path, and raw query, and with no fragment.
2. It requests the canonical HTTPS URL directly. The response must be `2xx`,
   complete a verified TLS 1.2-or-newer handshake whose leaf certificate is
   valid for the canonical hostname, and satisfy the selected response-header
   policy.

The client does not carry a cookie jar into either request. The probe closes
response bodies without reading them. Target URLs must not contain user
information or fragments. Operators must also keep credentials, signed query
parameters, tenant identifiers, and other private values out of probe URLs;
command arguments can be visible to other local processes even though the
report is redacted.

## Promotion modes

The HSTS policy is deliberately reversible until a separate domain-wide audit
approves broader scope:

| Mode | Redirect | HSTS `max-age` | `includeSubDomains` | `preload` |
| --- | --- | --- | --- | --- |
| `canary` | exactly `307` | exactly `300` | forbidden | forbidden |
| `steady` | exactly `308` | at least `31536000` | forbidden by default | forbidden by default |

Do not add `includeSubDomains` or `preload` as an incidental steady-state
change. Either one needs an explicit inventory and recovery review for every
affected hostname before the public contract is extended.
The fixed contract accepts only `max-age`, `includeSubDomains`, and `preload`
directives. Unknown directives, values on either flag, quoted or non-decimal
`max-age` values, and duplicates fail closed.

## Surface policies

Both `browser` and `api` require:

- `Strict-Transport-Security` according to the promotion table.
- `Content-Security-Policy` with `object-src 'none'`, `base-uri 'none'`, and
  `frame-ancestors 'none'`.
- `X-Frame-Options: DENY` and `X-Content-Type-Options: nosniff`.
- `Referrer-Policy: no-referrer`.
- `Permissions-Policy` containing exactly the lowercase `camera`, `microphone`,
  `geolocation`, `payment`, and `usb` members, each with the exact empty
  allowlist `()` and no extra members.
- `Cross-Origin-Opener-Policy: same-origin` and
  `Cross-Origin-Resource-Policy: same-origin`.

For a browser response, `default-src` must be exactly `'self'` or `'none'`.
More specific CSP directives may permit only the resources the application
actually needs. Browser script policies must not contain `'unsafe-inline'` or
`'unsafe-eval'`. An API response instead requires `default-src 'none'` and a
valueless `sandbox` directive.

Audited field values must contain only printable ASCII plus HTTP horizontal
tabs; other control bytes, non-ASCII bytes, and Unicode lookalikes fail closed.
Only HTTP optional whitespace (space and horizontal tab) is trimmed. Repeated
security headers are accepted only when every trimmed field value is exactly
the same. Conflicting duplicates fail closed. Duplicate directives inside
HSTS, CSP, or Permissions Policy also fail closed.

## Report and exit contract

The JSON report contains only a validated target identifier, mode, surface,
overall result, and stable rule identifiers with booleans. It never contains
the target URL, response body, header values, status text, certificate data, or
transport error text.

Exit codes are stable:

- `0`: every rule passed.
- `1`: command usage or target configuration was invalid, or an internal/report
  error prevented the audit from completing.
- `2`: the probe ran but a request or policy rule was blocked.

A green report proves only the observed transport/header boundary at that
instant. It does not prove application authorization, tenant isolation,
availability, backup recovery, or deployment readiness.
