---
name: safe-implementer
description: Create or modify code, tests, configuration, and project files safely with real file changes.

metadata:
  slash-command: enabled

---

# Safe Implementer

## Purpose

Create or modify code, tests, configuration, and project files safely with real file changes.

## When to use

Use this skill when the task matches the description above or when the central agent instructions route work to `/safe-implementer`.

## Operating model

1. Clarify the goal and constraints.
2. Inspect the minimum relevant repository context.
3. Produce a concise execution plan for non-trivial work.
4. Execute with tools when implementation is requested.
5. Validate the result with repository-native checks.
6. Summarize changed files, validation results, and residual risks.

## DevSecOps guardrails

- Do not read secrets, `.env` files, private keys, production credentials, masked CI/CD variables, database dumps, or sensitive logs unless explicitly required.
- Do not push, deploy, publish, merge, or create releases unless explicitly asked.
- Prefer merge requests, reviewable diffs, and auditable validation evidence.
- Prefer least privilege, minimal changes, and explicit rollback notes.
- Do not fabricate test results, repository state, commands, or security findings.

## Output

Provide:

- Actions taken
- Files changed or reviewed
- Validation performed
- Findings or risks
- Recommended next step
