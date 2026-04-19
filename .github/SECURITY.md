# Security Policy

## Supported Versions
Only the latest version of mimetype is supported.

We will only release security updates for the latest version.
If you've discovered a security vulnerability in an older version of mimetype,
please report it and we will [retract](https://go.dev/ref/mod#go-mod-file-retract)
that version from module proxies.

## Reporting a Vulnerability
We want to keep our software safe for everyone.

If you've discovered a security vulnerability in mimetype,
we appreciate your help in disclosing it to us in a responsible manner,
by creating a [security advisory](https://github.com/gabriel-vasile/mimetype/security/advisories).

## Threat model
These are the most common risk areas we think about when working on `mimetype`.

### Stealing maintainer's credentials
> The maintainers will have credentials stolen.

Measures:

1. `2FA` authentication enabled for maintainers.
2. `OIDC` auth for GitHub Actions instead of long lived access tokens.

### Malware in dependencies
> The library’s dependencies will be hijacked and malware will be installed to
importers of the library.

Measures:

1. Eliminate dependencies: `mimetype` has no compile time dependencies, except for the Golang standard library.
2. Keep a short list of Github Actions dependencies and pin them to a commit.
No workflows have write permission and a new release can only be done by a human maintainer.

### Logic errors in the library
> Logic errors in the library can lead to crashes or infinite loops. Specially
crafted inputs can exploit them and lead to denial-of-service attacks.

Measures:

1. Fuzzing: complex code is fuzzed.
2. Each release is tested against a [corpora of files](https://github.com/gabriel-vasile/mimetype_tests/)
that includes specially crafted, polyglot and edge-case files.

## Incident Response Plan
Next paragraphs outlines how we respond to security incidents, critical bugs,
or operational disruptions that could affect users or the trustworthiness of the project.

### Principles:

- Transparency: All incidents and fixes are documented and publicly available.
- Stewardship: Take responsibility for protecting users and the project.
- Protection: Act to minimize harm and provide guidance.

### Identification
- Incidents may be identified through vulnerability reports, dependency alerts, or community reports.
- All reports received via GitHub Security Advisories are treated as potential incidents.

### Assessment
- The maintainer evaluates severity of the incident.
- Affected versions and attack surface are determined.

### Containment
- If a released version of the library is compromised, then it will be flagged in
the advisory and [retracted](https://go.dev/ref/mod#go-mod-file-retract).
- Users are directed to pin a known-safe version.

### Remediation
- A fix is developed and validated in a private branch.
- The fix is released as a new version of the library.
- The GitHub Security Advisory is updated with the fixed version and mitigation steps.

### Notification
- Users are notified through the GitHub Security Advisory.
- Critical issues may also be announced via release notes and the project README.

### Post-Incident Review
- The root cause and timeline are documented in the advisory.
- Process improvements are applied to prevent recurrence.
