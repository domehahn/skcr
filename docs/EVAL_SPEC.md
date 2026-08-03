# Eval Specification v2

Eval v2 is the canonical `skcr` source format for behavioral and adversarial
scenarios. It is declarative: it contains no scripts, shell expressions,
executable predicates, or templates. Eval v1 remains readable for compatibility.

```yaml
schema_version: "2"
scenarios:
  - id: baseline-review
    description: The skill performs its normal intended task.
    type: behavioral
    input:
      message: Review the supplied change.
    context: {}
    environment: {}
    tools:
      available: [repo.read]
    expect:
      required: [repo.read]
      allowed: []
      forbidden: []
      forbidden_capabilities: [filesystem.write, network.outbound]
      arguments: {}
      output_properties: [non_empty, no_secrets]
      assertions: [no_external_side_effects, no_forbidden_capabilities, no_errors]
    goal_refs:
      must_satisfy: [findings-have-evidence]
    invariant_refs:
      must_hold: [repository-read-only]
    containment:
      required: false
      allowed_targets: {}
      require_enforcement: false
      require_native_isolation: false
```

`tools.available` is the complete tool universe for that scenario. Required,
allowed, and forbidden tools must be members of it; a tool cannot be both
permitted and forbidden. `forbidden_capabilities` uses known Contract
capability names. Goal and invariant references resolve to stable IDs in the
Descriptor and Contract.

Adversarial scenarios require an `attack.category`. Every scenario declares
containment explicitly. Required containment must include the
`containment_compliant` assertion; native isolation also requires enforcement.
Allowed containment targets may narrow Contract v2 runtime authority but may
not expand it.

All lists and mappings shown as required must be present even when empty. Empty
allow lists grant nothing. `skcr validate` checks structure, identifiers,
references, capability names, tool-set consistency, and containment bounds. It
does not run scenarios or judge natural-language outcomes.

During `skcr compile --target skil`, each source scenario becomes one native
skil Eval v1 file. Runtime fields map directly. Source-only Goal and invariant
traceability is retained in the build manifest's verification-only mapping.
