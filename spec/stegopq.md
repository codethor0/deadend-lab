# StegoPQ Handshake Transport Encoding

## 1. Overview

StegoPQ provides optional encoding of handshake payload bytes into plausible-looking carriers for anti-fingerprinting research. It embeds binary handshake data into normal-looking HTTP/JSON fields.

**Non-goals**: Does not defeat global traffic analysis. Only aims to reduce obvious static markers for basic classifiers.

## 2. Carriers

Three carrier types (A, B, C) are provided:

| ID | Carrier | Description | Example |
|----|---------|-------------|---------|
| A | JSON telemetry fields | Embed in JSON object values (e.g., `{"trace_id":"...", "span_id":"..."}`) | `{"trace_id":"aGVsbG8gd29ybGQ="}` |
| B | HTTP headers | Embed in tracing-like header values | `X-Trace-ID: aGVsbG8gd29ybGQ=` |
| C | URL query params | Embed in feature-flag style params | `?ff=abc&v=aGVsbG8gd29ybGQ=` |

## 3. Encoding Format

- Payload is base64-URL encoded (no padding) to produce ASCII-safe output.
- Carrier-specific wrappers add context (field names, delimiters).

### Carrier A (JSON)

```
{"trace_id":"<base64url(payload)>","span_id":"<optional_metadata>"}
```

### Carrier B (HTTP)

```
X-Trace-ID: <base64url(payload)>
```

### Carrier C (URL)

```
?ff=<base64url(payload)[:8]>&v=<base64url(payload)>
```

For C, the full payload goes in `v`; `ff` is a short prefix for plausible feature-flag look.

## 4. Decoding Rules

1. Parse carrier by type.
2. Extract embedded base64url string.
3. Base64url-decode to recover payload bytes.
4. Validate length and structure before passing to DEE.

## 5. Constraints

- Max embedded payload: 4096 bytes (carrier-dependent).
- Invalid base64 or truncated data returns decode error.
- Carrier type must be selected at encode; decode requires type hint or heuristic.

## 6. Fingerprint Test Harness

- Generate mixed traffic (stego vs non-stego).
- Simple classifiers (regex, length, entropy) attempt detection.
- Metric: detection accuracy over labeled corpus.
