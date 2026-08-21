# Cross-repository compatibility matrix

skcr is one of four sibling repos in an agentic skill supply chain:

```
skcr (author/compile)  →  skil (scan/attest)  →  skpm (package/publish)  →  SkillForge (registry)
```

Each pairing that matters for skcr is enforced by a CI job, not just
documented — a version bump on either side that breaks the pairing fails
CI, not silently drifts. This file records *what's currently pinned* so
the pairing is auditable without reading workflow YAML.

| Pairing                          | Enforced by                         | Currently pinned to | Status |
|-----------------------------------|--------------------------------------|----------------------|--------|
| skcr `main` (current) × skil stable | `.github/workflows/ci.yml` → `skil-interop` | [skil v0.2.0](https://github.com/domehahn/skil/releases/tag/v0.2.0) | ✅ enforced |

## What the `skil-interop` job actually checks

For every fixture under `tests/interop/*/`, it compiles the fixture with
this branch's `skcr` and then runs the pinned `skil`'s `validate`,
`scan --static-only`, and `verify` against the compiled output — i.e. it
proves skcr's *current* compiled output is still something skil's
*stable* release accepts and scans cleanly, not just that skcr's own
internal tests pass.

## Bumping the pin

When skil cuts a new stable release that skcr's interop job should
track:

1. Build that skil tag locally: `go install github.com/domehahn/skil/cmd/skil@vX.Y.Z`.
2. Run the exact loop `skil-interop` runs (see the job for the precise
   commands) against every fixture under `tests/interop/*/` — confirm
   `validate`/`scan`/`verify` all still pass before touching CI.
3. Update the `go install ... @vX.Y.Z` line and this file's pinned version
   together, in the same commit.

Don't bump the pin reflexively on every skil release — only when you've
actually verified the pairing still holds, or deliberately want to prove
it doesn't (and fix skcr's output or file an issue against skil).
