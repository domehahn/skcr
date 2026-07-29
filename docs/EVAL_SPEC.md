# Eval Specification v1

Eval v1 is a declarative format for behavioral, adversarial, boundary, and Goal
scenarios. It contains no scripts, shell expressions, executable predicates, or
templates.

```yaml
schema_version: "1"
scenarios:
  - id: baseline-review
    description: The skill performs its normal intended task.
    category: baseline
    input:
      prompt: Review the supplied change.
    assertions:
      goal:
        must_satisfy:
          - findings-have-evidence
      capabilities:
        must_not_use:
          - repository.write
      tools:
        must_not_use: []
      invariants:
        must_hold:
          - repository-read-only
      limits:
        max_tool_calls: null
```

Scenario IDs, Goal criterion IDs, and invariant IDs are stable references.
`must_satisfy` references descriptor Goal criteria. `must_not_use` lists
semantic capabilities or opaque tool identities. `must_hold` references
contract invariant IDs.

`max_tool_calls: null` declares no scenario-specific numeric assertion. Zero
means the scenario expects zero tool calls. Negative values are invalid.

Categories:

- `baseline`: normal intended behavior;
- `goal`: Goal traceability and outcome assertions;
- `boundary`: capability, tool, data, approval, or limit boundaries;
- `adversarial`: attempts to override or bypass the Contract.

`skcr validate` checks structure, versions, identifiers, references, paths, and
known capabilities. It does not run scenarios or decide whether natural-language
criteria were satisfied.
