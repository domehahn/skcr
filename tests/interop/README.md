# skcr → skil interop fixtures

Each subdirectory here is a full skcr descriptor-format skill (schema_version
"2": `descriptor.yaml`, `contract.yaml`, `assurance.yaml`, `dependencies.yaml`,
`evals/`, `integrations/`) that the `skil-interop` CI job (see
`.github/workflows/ci.yml`) compiles with `skcr compile --target skil
--require-lossless` and then runs through skil's `validate` / `scan
--static-only` / `verify` — proving the compiler's skil target output is
actually accepted by skil, not just that skcr's own unit tests believe it
should be.

Fixtures intentionally cover different declared-capability shapes, not just
one "everything is empty and clean" case:

| Fixture | Declares | Expected skil verdict |
| --- | --- | --- |
| `complete-skill` | No runtime capabilities (read-only review skill) | `CLEAR` / `PASS` |
| `network-tool-skill` | Outbound network, command execution, a tool effect | `REVIEW` / `WARN` — skil correctly flags these as overdeclared, since a descriptor-only fixture has no actual code exercising them. This is the intended, useful outcome: it proves skil's declared-vs-observed capability verification actually runs against skcr's compiled output. |

Both are expected to exit 0 (skil's CLI treats `REVIEW`/`WARN` as
non-fatal), so the job stays green while still asserting each fixture's
verdict is not silently downgraded to `CLEAR`/`PASS` when it shouldn't be.

## Adding a fixture

1. Copy an existing fixture directory and adjust `descriptor.yaml` /
   `contract.yaml` / `assurance.yaml` / `dependencies.yaml` / `evals/` /
   `integrations/` for the shape you want to exercise (a new capability
   combination, a real MCP server declaration, delegation, etc.).
2. Validate locally before opening a PR — the CI job as written will run
   every directory under `tests/interop/` automatically:

   ```bash
   go build -o bin/skcr ./cmd/skcr
   go install github.com/domehahn/skil/cmd/skil@<pinned-commit>  # see ci.yml
   ./bin/skcr compile tests/interop/<your-fixture> --target skil \
     --output /tmp/compiled --require-lossless
   skil validate /tmp/compiled/<your-fixture>/<your-fixture>
   skil scan /tmp/compiled/<your-fixture>/<your-fixture> --static-only
   skil verify /tmp/compiled/<your-fixture>/<your-fixture>
   ```
3. No code changes are needed to wire a new fixture into CI — the job loops
   over `tests/interop/*/`.
