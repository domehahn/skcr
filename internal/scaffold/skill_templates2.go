package scaffold

const dependencyRiskReviewerBody = `
# {{.Title}}

## Purpose

Review direct and transitive dependencies, container base images, and third-party libraries for known vulnerabilities, license risks, maintainer health, pinning gaps, and supply-chain attack surface. Produce a prioritized risk register with concrete update or removal recommendations. Every finding must be actionable: a specific package, version, CVE, or license conflict with a recommended resolution.

## When to Use This Skill

- A pull request adds, removes, or upgrades a dependency.
- A periodic dependency audit is scheduled.
- A CVE scanner or SCA tool has produced findings that need triage.
- A container image is being added or its base image is being changed.
- You are routed here by the central agent via $dependency-risk-reviewer.

## Inputs

- Package manifest files: go.mod, package.json, requirements.txt, Gemfile, pom.xml, Cargo.toml, build.gradle.
- Lock files: go.sum, package-lock.json, yarn.lock, Pipfile.lock, Gemfile.lock.
- Dockerfile or container image specification.
- SCA or vulnerability scanner output, if available.
- Project license policy and approved license list.
- SBOM, if available.

## Skill-Specific Operating Model

1. **Inventory direct dependencies.** List every direct dependency with its current version and the project's stated version constraint.
2. **Inventory transitive dependencies.** Identify transitive dependencies brought in by direct dependencies. Flag any transitive dependency that introduces a significantly different risk profile from the direct dependency.
3. **Check version pinning.** Verify that every dependency is pinned to an exact version or a reproducible hash in the lock file. Flag dependencies pinned only to a major or minor version range.
4. **Scan for known CVEs.** For each dependency, check for known CVEs at the current pinned version. Classify each finding by CVSS score: Critical (≥9.0), High (7.0–8.9), Medium (4.0–6.9), Low (<4.0).
5. **Assess license compliance.** For each dependency, identify its license. Flag licenses that conflict with the project's license, that impose copyleft obligations (GPL, AGPL), or that are not in the approved license list.
6. **Assess maintainer health.** Check whether the package is actively maintained: last release date, issue response rate, number of active maintainers. Flag unmaintained packages (no release in 24+ months) as Medium or High risk depending on security criticality.
7. **Check for supply-chain attack vectors.** Look for: dependency confusion risk (internal package names that could be squatted on a public registry), typosquatting (package names one character away from popular packages), packages with very recent first releases and high download counts (potential compromise), packages installed from non-canonical sources.
8. **Check SBOM completeness.** If an SBOM exists, verify it is current. If none exists, flag its absence for projects with compliance or supply-chain requirements.
9. **Assess update strategy.** For each flagged dependency, state whether an upgrade path exists, whether the upgrade introduces breaking changes, and what the recommended action is.
10. **Produce a risk register.** Table with columns: Package, Current Version, Risk Type (CVE/License/Maintainer/Supply-Chain), Severity, Description, Recommended Action.

## Skill-Specific Checklist

- [ ] All direct dependencies are listed with current version and version constraint.
- [ ] Lock file exists and is committed to version control.
- [ ] All dependencies are pinned to exact versions or reproducible hashes in the lock file.
- [ ] CVE scan results are present for all direct and high-risk transitive dependencies.
- [ ] All Critical and High CVEs have a recommended remediation (upgrade, patch, removal, or accepted risk).
- [ ] All dependency licenses are identified and checked against the project license policy.
- [ ] GPL and AGPL licensed packages are flagged for legal review if the project is not itself GPL/AGPL.
- [ ] Unmaintained packages (no release in 24+ months) are flagged.
- [ ] Packages with a single maintainer and no organizational backing are flagged for bus-factor risk.
- [ ] Package names are checked for typosquatting risk (one-character variants of popular packages).
- [ ] Internal package names are checked against public registries for dependency confusion risk.
- [ ] Container base images are pinned by digest, not only by tag.
- [ ] Container base images are from trusted, official, or organization-approved sources.
- [ ] SBOM generation is assessed: exists, current, or absent with risk noted.
- [ ] Unnecessary dependencies (used in dev only but included in production build) are identified.

## Decision Rules

- If a Critical CVE (CVSS ≥ 9.0) exists for a production dependency, classify the finding as Blocker and do not approve the pull request until the dependency is upgraded, removed, or a risk-acceptance decision is documented.
- If a High CVE (CVSS 7.0–8.9) exists and no upgrade path is available, document an accepted-risk decision with a remediation deadline.
- If a dependency is licensed under GPL or AGPL and the project is not open-source under a compatible license, flag it as a Legal Blocker and escalate to the legal or compliance team.
- If a package has had no release in 24+ months and handles security-critical functions (cryptography, authentication, input parsing), classify the risk as High.
- If a dependency confusion risk exists (internal package name resolvable from a public registry), classify it as High and recommend namespace scoping or registry pinning.
- If a lock file is absent or not committed to VCS, the build is not reproducible; flag it as High and block merge until the lock file is committed.
- Do not recommend removing a dependency without verifying that all its uses in the codebase have a documented replacement.

## DevSecOps Guardrails

- Do not approve a dependency addition without checking its license, CVE status, and maintainer health.
- Do not accept a container image pinned only by tag; tags are mutable and can be silently replaced.
- Do not dismiss a CVE as non-exploitable without verifying that the vulnerable code path is actually unreachable in this application.
- Do not approve a dependency pulled from a non-canonical source (fork, git reference, private mirror without integrity check) without explicit justification and hash verification.
- Do not include actual API keys, credentials, or registry tokens in the dependency review artifact.
- Do not treat an SBOM as a one-time artifact; flag when the SBOM is outdated relative to the current lock file.

## Output Requirements

- **Dependency inventory**: all direct dependencies with version, constraint, and lock-file pin status.
- **CVE findings**: each CVE with package, version, CVSS score, description, exploitability assessment, and recommended action.
- **License compliance report**: each dependency with its license and compliance status against the project policy.
- **Maintainer health report**: flagged packages with last release date, active maintainers, and risk assessment.
- **Supply-chain risk findings**: dependency confusion, typosquatting, and suspicious package findings.
- **Container image assessment**: base image, pin method (tag vs. digest), source trust, and update status.
- **SBOM status**: present/current/absent with recommendation.
- **Risk register**: consolidated table with Package, Risk Type, Severity, and Recommended Action.

## Acceptance Criteria

- All Critical CVEs for production dependencies have a remediation or documented risk-acceptance decision.
- All dependency licenses are checked against the project policy and conflicts are flagged.
- All dependencies are pinned to exact versions or hashes in a committed lock file.
- Container base images are pinned by digest.
- Supply-chain risks (dependency confusion, typosquatting) are assessed for every new dependency.
- Unmaintained packages handling security-critical functions are flagged with risk assessment.

## Anti-Patterns

- **Transitive blindness**: Reviewing only direct dependencies and ignoring transitive dependencies that introduce known vulnerabilities.
- **Tag-pinned images**: Accepting container images pinned by tag instead of by digest, allowing silent image replacement.
- **CVSS-score-only triage**: Dismissing a CVE because the CVSS score is below a threshold without assessing whether the vulnerable code path is reachable.
- **Lock file omission**: Not committing the lock file to VCS, making builds non-reproducible and dependency versions unauditable.
- **License assumption**: Assuming a dependency is MIT or Apache without checking its actual license, creating legal exposure.
- **Dependency hoarding**: Adding multiple packages that provide overlapping functionality, increasing the attack surface without benefit.
- **Upgrade avoidance**: Deferring security upgrades indefinitely because they introduce breaking API changes, without a documented remediation plan and deadline.

## Changelog

### {{.Version}} - {{.LastModified}}

- Initial generated DevSecOps SDLC skill.
`

const ciCdSecurityReviewerBody = `
# {{.Title}}

## Purpose

Review CI/CD pipeline definitions, GitHub Actions workflows, GitLab CI configurations, and build processes for security risks. Identify excessive token permissions, secrets exposure, script injection vulnerabilities, cache poisoning risks, artifact integrity gaps, unprotected deployment gates, and runner trust issues. Produce findings that allow a pipeline maintainer to harden the pipeline before it is used as an attack vector.

## When to Use This Skill

- A CI/CD pipeline configuration file is added or modified.
- A new GitHub Actions workflow, GitLab CI job, or build script is introduced.
- A pipeline handles secrets, deploys to production, or has access to sensitive systems.
- You are routed here by the central agent via $ci-cd-security-reviewer.

## Inputs

- CI/CD configuration files: .github/workflows/*.yml, .gitlab-ci.yml, Jenkinsfile, Dockerfile, Makefile.
- Pipeline secrets and variable declarations (names only, not values).
- Deployment target descriptions and environment separation model.
- Branch protection rules and environment protection rules.
- Runner configuration: self-hosted vs. hosted, runner labels, isolation model.

## Skill-Specific Operating Model

1. **Identify all pipeline triggers.** List every event that can trigger a pipeline execution: push, pull_request, workflow_dispatch, schedule, API trigger. Flag triggers that can be initiated by external contributors without review (e.g., pull_request_target in GitHub Actions, which runs in the context of the base branch).
2. **Audit token and permission scopes.** For each workflow or job, check the declared permissions. Flag any job that uses a broad permission scope (contents: write, id-token: write, packages: write) when a narrower scope would suffice.
3. **Check for script injection.** For every location where a GitHub Actions event-context expression (e.g. the pull-request title, body, branch name, or commit message) or a GitLab CI predefined variable (e.g. CI_COMMIT_MESSAGE) is interpolated directly into a run or script block, check whether the value can be controlled by an external contributor and reach a shell command without sanitization.
4. **Audit secrets usage.** Verify that secrets are referenced via the secrets context, not hardcoded. Check that secrets are not echoed in log output, environment variables printed to the log, or included in artifact contents.
5. **Check third-party action and image versions.** Verify that every third-party action is pinned to a commit SHA, not a mutable tag. Verify that container images used in jobs are pinned by digest.
6. **Assess cache configuration.** Check whether cache keys are based on user-controlled inputs. Check whether cache contents could be poisoned by a malicious pull request targeting the cache used by a privileged workflow.
7. **Assess artifact integrity.** Check whether build artifacts are signed or hashed before upload and verified before use in a downstream job or deployment.
8. **Assess deployment gates.** Verify that deployments to production environments require explicit approval from a named owner, not just a passing CI run. Check that the approval mechanism cannot be bypassed by pushing directly to the deployment branch.
9. **Assess runner trust.** For self-hosted runners, check whether they are isolated per-job, whether they persist state between runs, and whether they have access to sensitive credentials beyond what is needed for the job.
10. **Check OIDC configuration.** If OIDC tokens are used for cloud provider authentication, verify that the audience, subject, and issuer claims are appropriately scoped and that token lifetime is minimal.

## Skill-Specific Checklist

- [ ] Every pipeline trigger is listed and external-contributor-accessible triggers are flagged.
- [ ] pull_request_target in GitHub Actions is reviewed for context escalation risk.
- [ ] Every workflow and job has a declared permission scope at the minimum required level.
- [ ] No job uses write-all permissions or defaults without an explicit permissions block.
- [ ] Every location where event context variables reach a shell command is checked for script injection.
- [ ] No secrets are hardcoded in pipeline configuration files.
- [ ] No secrets are printed to logs via echo, env, or debug steps.
- [ ] Every third-party GitHub Action is pinned to a full commit SHA, not a tag or branch.
- [ ] Every container image used in pipeline jobs is pinned by digest.
- [ ] Cache keys are not based on user-controlled inputs that could be manipulated to poison a shared cache.
- [ ] Build artifacts are signed or hashed before cross-job transfer and verified before use.
- [ ] Production deployments require explicit manual approval from a named approver.
- [ ] Branch protection prevents direct pushes to deployment branches without a passing pipeline and approver.
- [ ] Self-hosted runners are ephemeral (cleaned between jobs) or their persistent state is documented and justified.
- [ ] OIDC token claims (audience, subject, issuer) are scoped to the minimum necessary for the cloud provider action.

## Decision Rules

- If a workflow uses pull_request_target and checks out or runs code from the fork, classify it as Critical (privilege escalation via fork PR); the workflow must be restructured.
- If a job uses contents: write or id-token: write without a clear justification, classify it as High and recommend scoping to the specific operation.
- If an untrusted input (event title, commit message, branch name, PR body) reaches a shell command without sanitization, classify it as High (script injection).
- If a third-party action is pinned to a tag instead of a SHA, classify it as Medium and recommend pinning to the SHA of the current tag.
- If a production deployment has no manual approval gate, classify it as High regardless of how thorough the automated tests are.
- If a self-hosted runner has access to production credentials and is not ephemeral, classify it as High and recommend ephemeral runner configuration or credential isolation.
- If OIDC token subjects are wildcarded (sub: repo:org/*), classify it as Medium and recommend restricting to the specific repository and ref.

## DevSecOps Guardrails

- Do not approve a pipeline that runs untrusted code (from a fork or external contributor) with repository write permissions or access to repository secrets.
- Do not accept mutable action version references (uses: actions/checkout@v4) as secure; require SHA pinning.
- Do not treat a passing test suite as sufficient authorization for a production deployment; require explicit human approval.
- Do not allow pipeline secrets to be accessible in jobs that do not need them; use environment-scoped secrets.
- Do not accept self-hosted runners for security-sensitive jobs without verifying isolation and ephemeral configuration.
- Do not include actual secret values, tokens, or credentials in the pipeline review artifact.

## Output Requirements

- **Trigger risk assessment**: each trigger with its audience, privilege level, and fork-access risk.
- **Permission audit**: each workflow and job with declared permissions, minimum required scope, and gap.
- **Script injection findings**: each location where event context reaches a shell command, with severity and sanitization recommendation.
- **Secrets hygiene report**: hardcoded secrets, log-exposure risks, and scope violations.
- **Action and image pinning report**: unpinned third-party actions and container images.
- **Cache and artifact integrity assessment**: cache key risks and artifact signing status.
- **Deployment gate assessment**: production deployment approval requirements and bypass risks.
- **Runner trust assessment**: runner type, isolation model, and credential access scope.

## Acceptance Criteria

- No workflow runs untrusted code with repository secrets or write permissions without an explicit security justification.
- All third-party actions are pinned to commit SHAs.
- No secrets are printed to pipeline logs or hardcoded in configuration files.
- Production deployments require at least one named, manual approver.
- Script injection risks from event context variables are remediated or documented with risk acceptance.
- Self-hosted runners handling sensitive credentials are ephemeral or their isolation model is documented.

## Anti-Patterns

- **pull_request_target without protection**: Using pull_request_target to access secrets while also checking out and running code from the forked PR branch.
- **Tag-pinned actions**: Using third-party actions pinned to a mutable tag (v3) instead of a commit SHA, allowing the action to be silently replaced.
- **Secrets in env**: Setting secrets as environment variables at the job level, making them accessible to all steps including third-party actions.
- **Wildcard permissions**: Using permissions: write-all or omitting the permissions block, granting every workflow step the maximum token scope.
- **Approval theater**: Adding a manual approval gate that can be bypassed by pushing directly to the deployment branch.
- **Persistent runner secrets**: Storing cloud provider credentials on a persistent self-hosted runner accessible to any job that runs on it.
- **Log-printed secrets**: Using debug steps, env dumps, or echo commands that print secret values to the pipeline log.

## Changelog

### {{.Version}} - {{.LastModified}}

- Initial generated DevSecOps SDLC skill.
`

const containerSecurityReviewerBody = `
# {{.Title}}

## Purpose

Review Dockerfiles, container image configurations, and Kubernetes or container runtime settings for security risks. Identify root runtime, unpinned or riskly base images, secrets baked into images, missing multi-stage build patterns, excessive capabilities, missing read-only filesystem configuration, and image scanning gaps. Produce concrete Dockerfile and runtime configuration improvements.

## When to Use This Skill

- A Dockerfile is added or modified.
- A container image is being onboarded or its base image is being changed.
- A Kubernetes deployment manifest or Helm chart is being reviewed.
- A container image scan has produced findings that need triage.
- You are routed here by the central agent via $container-security-reviewer.

## Inputs

- Dockerfile or Containerfile.
- Container image manifest or digest reference.
- Kubernetes deployment, pod spec, or Helm chart values.
- Container registry configuration and image scanning results.
- Runtime security policy (Pod Security Standards, OPA/Gatekeeper rules).

## Skill-Specific Operating Model

1. **Review base image selection.** Check whether the base image is from a trusted, official, or organization-approved source. Check whether a minimal base (distroless, Alpine, slim) is used where appropriate.
2. **Check base image pinning.** Verify that the base image is pinned by digest, not only by tag. Tags are mutable and can be silently replaced.
3. **Check for non-root runtime.** Verify that the container runs as a non-root user. Check that the USER instruction is present in the Dockerfile and that the user does not have UID 0.
4. **Check for package update hygiene.** Verify that package installations use pinned versions or that the image layer is rebuilt frequently. Check that unnecessary packages are not installed.
5. **Scan for secrets in the image.** Check Dockerfile build arguments, ENV instructions, and RUN steps for hardcoded secrets, tokens, API keys, or credentials. Verify that secrets used during the build are not present in intermediate layers.
6. **Assess multi-stage build usage.** For compiled languages, verify that a multi-stage build is used to separate the build environment from the runtime image. Verify that build tools, source code, and test artifacts are not present in the final image layer.
7. **Check the runtime attack surface.** Identify all ports exposed, all volumes mounted, and all capabilities granted. Flag any capability beyond the minimum required for the container's function.
8. **Review Kubernetes security context.** Check allowPrivilegeEscalation, readOnlyRootFilesystem, capabilities drop/add, runAsNonRoot, runAsUser, seccompProfile, and hostPID/hostNetwork/hostIPC.
9. **Check resource limits.** Verify that CPU and memory limits and requests are set. Containers without limits can be used for resource exhaustion attacks.
10. **Assess health check configuration.** Verify that a liveness probe and readiness probe are defined. Containers without health checks cannot be automatically recovered by the orchestrator.
11. **Check image scanning integration.** Verify that the image is scanned for CVEs before deployment. Check that Critical and High CVE findings block promotion to production.

## Skill-Specific Checklist

- [ ] Base image is from a trusted, official, or organization-approved source.
- [ ] Base image is pinned by digest (SHA256), not only by tag.
- [ ] Container does not run as root (UID 0); USER instruction is present in the Dockerfile.
- [ ] No secrets, API keys, tokens, or passwords appear in ENV instructions, build args exposed in the final image, or RUN commands that leave artifacts in intermediate layers.
- [ ] Multi-stage build is used for compiled binaries; build toolchain is not present in the final image.
- [ ] Final image layer does not contain source code, test files, or build-time credentials.
- [ ] Only required packages are installed; no debugging tools (curl, wget, bash) are present in the production image unless justified.
- [ ] Only required capabilities are granted; all others are dropped (CAP_NET_ADMIN, CAP_SYS_PTRACE flagged unless explicitly justified).
- [ ] readOnlyRootFilesystem: true is set in the Kubernetes security context unless a writable path is explicitly required and mounted.
- [ ] allowPrivilegeEscalation: false is set in the Kubernetes security context.
- [ ] runAsNonRoot: true and a specific runAsUser UID are set in the Kubernetes security context.
- [ ] CPU and memory resource limits and requests are defined in the Kubernetes manifest.
- [ ] Liveness and readiness probes are defined.
- [ ] seccompProfile is set (RuntimeDefault or a custom profile).
- [ ] Image scanning is integrated in CI and Critical/High CVEs block promotion to production.

## Decision Rules

- If the container runs as root and there is no documented justification, classify it as High and require a non-root user to be specified.
- If the base image is pinned only by tag, classify it as Medium and require digest pinning before the image is used in a regulated or production environment.
- If a secret is present in any image layer (including intermediate layers from multi-stage builds), classify it as Blocker and require the secret to be removed and all affected image versions to be treated as compromised.
- If a build-stage tool (compiler, test runner, package manager with full access) is present in the runtime image, classify it as Medium and recommend multi-stage build restructuring.
- If CAP_SYS_ADMIN or CAP_NET_ADMIN is granted without documented justification, classify it as High.
- If no CPU or memory limits are set, classify it as Medium; resource exhaustion is an availability risk.
- If image scanning is not integrated or Critical CVEs are not blocking promotion, classify it as High.

## DevSecOps Guardrails

- Do not accept containers running as root in production without documented justification and compensating controls.
- Do not accept base images that are not from a trusted or organization-approved source.
- Do not accept hardcoded secrets in any image layer, including build stages not present in the final image (they may still exist in the image history).
- Do not accept images with no resource limits in environments where they share a node with other workloads.
- Do not treat image tag pinning as equivalent to digest pinning; tags are mutable.
- Do not skip image scanning for images used in production or staging environments.

## Output Requirements

- **Base image assessment**: image source, trust status, pinning method, CVE scan result.
- **Secret scan results**: any secrets found in Dockerfile instructions or image layers, with remediation.
- **Runtime privilege assessment**: running user, capabilities granted/dropped, privilege escalation configuration.
- **Kubernetes security context assessment**: each security context field with current value and recommended value.
- **Resource limits assessment**: CPU/memory limits and requests with gaps.
- **Multi-stage build assessment**: whether the runtime image is minimal or contains unnecessary build artifacts.
- **Image scanning integration status**: CI scan step present, CVE gate configuration, last scan result.
- **Findings table**: Severity, Component, Finding, Recommended Fix.

## Acceptance Criteria

- Container runs as a non-root user with a specific, non-zero UID.
- Base image is pinned by digest.
- No secrets appear in any image layer, including build stages.
- Multi-stage builds are used for compiled binaries.
- Kubernetes security context sets allowPrivilegeEscalation: false, readOnlyRootFilesystem: true, and runAsNonRoot: true.
- CPU and memory limits and requests are defined.
- Image scanning is integrated in CI with Critical CVEs blocking promotion.

## Anti-Patterns

- **Root runtime**: Running the application process as root because it is simpler than configuring a non-root user.
- **Tag-pinned base images**: Using FROM ubuntu:latest or FROM node:20 without a digest, allowing silent base image substitution.
- **Build-secret leakage**: Passing secrets via ARG or ENV in the build stage without clearing them, leaving them in the image history.
- **Monolithic build image**: Using the same image for building and running, including compilers, test frameworks, and package managers in the production runtime.
- **Capabilities sprawl**: Adding broad capabilities (SYS_ADMIN) instead of investigating which specific capability the application actually needs.
- **No resource limits**: Omitting CPU and memory limits because they require tuning, leaving the container free to consume all node resources.
- **Scan-after-deploy**: Running image scans only after deployment to production rather than as a CI gate before promotion.

## Changelog

### {{.Version}} - {{.LastModified}}

- Initial generated DevSecOps SDLC skill.
`

const iacSecurityReviewerBody = `
# {{.Title}}

## Purpose

Review Infrastructure-as-Code definitions for security misconfigurations, excessive privilege, network exposure, missing encryption, inadequate logging, and Kubernetes runtime risks. Cover Terraform, CloudFormation, Pulumi, Kubernetes manifests, Helm charts, and Kustomize overlays. Every finding must include the specific resource, attribute, and a concrete remediation.

## When to Use This Skill

- A Terraform plan, CloudFormation template, or Kubernetes manifest is added or modified.
- A Helm chart or Kustomize overlay is being reviewed.
- An infrastructure change affects IAM roles, network security groups, storage encryption, or logging configuration.
- A compliance review requires evidence that IaC follows the organization's security baseline.
- You are routed here by the central agent via $iac-security-reviewer.

## Inputs

- Terraform files (.tf), CloudFormation templates (.yaml, .json), Pulumi programs.
- Kubernetes manifests, Helm chart values.yaml, Kustomize overlays.
- IaC linting or policy-as-code scanner output (tfsec, checkov, kube-score).
- Cloud provider IAM policy documents and trust relationships.
- Organization security baseline or compliance control list.

## Skill-Specific Operating Model

1. **Audit IAM and access control.** Check every IAM role, policy, service account, and cluster role binding for least-privilege compliance. Flag wildcards in Action, Resource, or Principal fields.
2. **Assess network exposure.** Check security groups, firewall rules, network policies, and ingress configurations. Flag any rule that allows 0.0.0.0/0 or ::/0 ingress on non-public ports or without justification.
3. **Check for public access.** Identify storage buckets, databases, container registries, and Kubernetes services configured for public access. Flag public access for resources that do not require it.
4. **Verify encryption at rest.** Check that storage volumes, databases, object stores, and message queues have encryption at rest enabled with customer-managed or provider-managed keys as required by policy.
5. **Verify encryption in transit.** Check that load balancers, ingress controllers, service meshes, and database connections enforce TLS and reject unencrypted connections.
6. **Review logging and audit trail.** Check that VPC flow logs, cloud trail / audit logs, Kubernetes audit logs, and load balancer access logs are enabled for all environments. Check that log retention meets policy requirements.
7. **Review backup and recovery.** Check that databases and critical storage have automated backups configured with retention periods that meet recovery time and point objectives.
8. **Assess Kubernetes security contexts.** For every Pod spec, check: runAsNonRoot, allowPrivilegeEscalation, readOnlyRootFilesystem, capabilities, seccompProfile, hostPID, hostNetwork, hostIPC.
9. **Check resource limits and quotas.** Verify that CPU and memory limits and requests are set for every container. Verify that namespace resource quotas and limit ranges are defined.
10. **Assess admission control.** Check whether Pod Security Admission, OPA/Gatekeeper, or Kyverno policies are in place and whether they enforce the organization's baseline.
11. **Check for secrets in IaC.** Scan for hardcoded secrets, passwords, tokens, or sensitive data in IaC files committed to VCS. Verify that secrets are referenced from a secrets manager or vault.
12. **Identify drift risk.** Check whether the IaC is the authoritative source or whether manual changes are possible. Flag resources that are not managed by IaC (manual drift risk).

## Skill-Specific Checklist

- [ ] Every IAM role, policy, and service account is checked for least-privilege (no wildcard Action or Resource).
- [ ] Every network security group, firewall rule, or Kubernetes network policy is checked for overly permissive ingress (0.0.0.0/0 on non-public ports).
- [ ] No storage buckets, databases, or registries are configured with public access unless explicitly required and justified.
- [ ] Encryption at rest is enabled for all databases, object stores, and persistent volumes.
- [ ] Encryption in transit (TLS) is enforced on all load balancers, ingress controllers, and database connections.
- [ ] Cloud audit logs, VPC flow logs, and Kubernetes audit logs are enabled and retained per policy.
- [ ] Automated backups are configured for all stateful resources with documented retention periods.
- [ ] Every Kubernetes Pod spec sets runAsNonRoot: true and allowPrivilegeEscalation: false.
- [ ] Every Kubernetes container spec sets readOnlyRootFilesystem: true or the writable mount is documented.
- [ ] Container capabilities are explicitly dropped; no container is granted SYS_ADMIN or NET_ADMIN without documented justification.
- [ ] CPU and memory limits and requests are set for every container.
- [ ] Namespace-level resource quotas or limit ranges are defined.
- [ ] No secrets, passwords, tokens, or private keys are hardcoded in any IaC file.
- [ ] Admission control policies are in place and enforce the organization's Pod security baseline.
- [ ] All infrastructure resources are managed by IaC; manually created resources are identified as drift risks.

## Decision Rules

- If an IAM role has a wildcard (*) in Action or Resource, classify it as High unless the role is the organization-wide break-glass role with MFA enforcement.
- If a security group or firewall rule allows 0.0.0.0/0 ingress on any port below 1024, classify it as High unless the resource is an intentional public load balancer with documented justification.
- If a storage bucket or database is publicly accessible and does not serve public content, classify it as Critical.
- If encryption at rest is not enabled for a database containing personal data or credentials, classify it as High.
- If a hardcoded secret is found in any IaC file, classify it as Blocker; rotate the secret immediately and remove it from the IaC.
- If a Kubernetes Pod spec does not set runAsNonRoot or allowPrivilegeEscalation, classify it as Medium.
- If cloud audit logging is disabled for any production account or cluster, classify it as High.

## DevSecOps Guardrails

- Do not approve IaC that creates publicly accessible storage or databases without explicit documentation of the business requirement and compensating controls.
- Do not accept wildcards in IAM Action or Resource fields without a documented justification reviewed by the security team.
- Do not accept hardcoded secrets in IaC files; secrets must be sourced from a secrets manager at deploy time.
- Do not accept IaC that disables audit logging for any production environment.
- Do not treat "plan output looks correct" as equivalent to a security review; review the resource definitions, not just the planned changes.
- Do not accept infrastructure resources that are not managed by IaC in production without documented justification and a remediation plan.

## Output Requirements

- **IAM assessment**: each role, policy, and service account with privilege scope and gaps.
- **Network exposure assessment**: each security group, firewall rule, and network policy with ingress rules and public exposure findings.
- **Encryption assessment**: encryption-at-rest and in-transit status per resource type.
- **Logging and audit trail assessment**: enabled log types, retention periods, and gaps.
- **Kubernetes security context assessment**: each Pod spec with security field values and recommendations.
- **Secrets hygiene findings**: any hardcoded secrets or sensitive values in IaC files.
- **Admission control assessment**: policy enforcement mechanism and baseline coverage.
- **Findings table**: Severity, Resource Type, Resource Name, Finding, Recommended Fix.

## Acceptance Criteria

- No IAM policy uses wildcard Action or Resource without documented justification.
- No storage, database, or registry has public access without documented business justification.
- Encryption at rest is enabled for all stateful resources.
- Audit logging is enabled for all production environments.
- No hardcoded secrets appear in any IaC file.
- Every Kubernetes Pod spec sets runAsNonRoot and allowPrivilegeEscalation: false.
- Resource limits are defined for all containers.

## Anti-Patterns

- **Wildcard IAM**: Creating roles with Action: "*" or Resource: "*" because it is easier than enumerating the exact permissions needed.
- **Publicly open security groups**: Allowing 0.0.0.0/0 ingress on database or internal API ports because it simplifies local development.
- **Hardcoded secrets in IaC**: Embedding database passwords or API tokens directly in Terraform variables or Kubernetes manifests.
- **Disabled audit logs**: Disabling cloud audit logs to reduce storage costs without assessing the compliance and incident response impact.
- **Root containers in Kubernetes**: Deploying containers without runAsNonRoot because fixing the application to run as non-root requires effort.
- **Unmanaged drift**: Allowing manual changes to production infrastructure alongside IaC, creating configuration drift that the IaC cannot detect or remediate.
- **Plan-only review**: Reviewing only the Terraform plan output without checking the underlying resource definitions for misconfiguration.

## Changelog

### {{.Version}} - {{.LastModified}}

- Initial generated DevSecOps SDLC skill.
`

const secretsAuditorBody = `
# {{.Title}}

## Purpose

Detect and prevent the exposure of secrets, credentials, tokens, private keys, and other sensitive values in source code, configuration files, test fixtures, CI/CD pipelines, and application logs. When exposure is found, produce an immediate remediation plan: what to rotate, where to revoke, and how to prevent recurrence. A found secret is a production incident until rotated and revoked.

## When to Use This Skill

- A pull request adds or modifies configuration files, environment files, test fixtures, or CI/CD configuration.
- A periodic secrets scan is scheduled or a scanner has produced findings.
- A secret leak has been reported or suspected.
- A new service integration requires credentials and the implementation is being reviewed.
- You are routed here by the central agent via $secrets-auditor.

## Inputs

- Source code files, configuration files (.env, config.yaml, application.properties).
- CI/CD configuration files (.github/workflows/*.yml, .gitlab-ci.yml).
- Test fixtures, example files, documentation with code samples.
- Git history (when scanning for historical exposure).
- Secrets scanner output (trufflehog, gitleaks, detect-secrets).
- List of known secret formats for the project (API key prefixes, token formats).

## Skill-Specific Operating Model

1. **Scan code and configuration files.** Check every file in the change for patterns matching known secret formats: API key prefixes (sk-, ghp_, glpat-, AWS4, AKIA), PEM headers (BEGIN PRIVATE KEY, BEGIN RSA PRIVATE KEY), connection strings with embedded credentials, bearer tokens, and base64-encoded values of unusual length.
2. **Scan CI/CD configuration.** Check pipeline files for secrets passed as environment variables inline, printed to logs via echo or debug steps, or embedded in scripts.
3. **Scan test fixtures and documentation.** Check example files, test data, seed scripts, and documentation code samples for real-looking secrets. Distinguish real secrets from obviously synthetic test values.
4. **Scan git history.** For suspected or reported leaks, scan the commit history of affected files. A secret deleted in a later commit is still exposed in the git history; the entire affected secret must be rotated.
5. **Assess secret storage mechanism.** For every secret reference in code, check whether it is sourced from an approved secrets manager (HashiCorp Vault, AWS Secrets Manager, GCP Secret Manager, Azure Key Vault, CI/CD secrets store) or from a plain environment variable or file.
6. **Assess secret access scope.** Check whether the secret has the minimum permissions required. Flag secrets with broad permissions (admin access, all-read) when narrower scopes would suffice.
7. **Check secret rotation policy.** Verify whether the secret has a rotation policy. Flag secrets with no rotation schedule, especially long-lived tokens and service account keys.
8. **Assess revocation capability.** For each found or suspected exposure, verify whether the secret can be revoked immediately from the issuing system. Document the revocation steps and who has the authority to revoke.
9. **Produce a remediation plan.** For each confirmed exposure: rotation urgency (immediate / within 24h / scheduled), revocation steps, history-scrubbing approach (git-filter-repo, force-push to remove from history with stakeholder approval), and prevention recommendation.
10. **Recommend prevention controls.** For each finding, recommend a specific prevention control: pre-commit hook, CI secrets scanner gate, migration from plain env vars to a secrets manager, or a secrets management policy change.

## Skill-Specific Checklist

- [ ] All files in the change are scanned for known secret format patterns.
- [ ] CI/CD pipeline files are checked for inline secrets and log-printing steps.
- [ ] Test fixtures and example files are checked for real-looking credentials.
- [ ] Git history is checked for secrets deleted in later commits but still present in history.
- [ ] Every secret reference is sourced from an approved secrets manager or CI/CD secrets store, not from a plain file or environment variable committed to VCS.
- [ ] No .env files, credentials.json, or key.pem files are committed to VCS.
- [ ] Secrets are not passed between pipeline steps via unencrypted artifact uploads.
- [ ] Secrets do not appear in pipeline log output.
- [ ] Every discovered secret has an assessed access scope (what systems, what permissions).
- [ ] Every discovered secret has a rotation urgency and revocation plan.
- [ ] Long-lived tokens and service account keys are flagged for rotation policy review.
- [ ] Application logs are checked for patterns that could expose tokens, passwords, or session identifiers.
- [ ] Error responses are checked for connection strings or internal paths that reveal credential structure.
- [ ] A pre-commit hook or CI gate is recommended if none exists.
- [ ] All findings are classified as Confirmed Exposure, Likely Exposure, or Test/Synthetic Value with reasoning.

## Decision Rules

- If a confirmed secret is found in any committed file, classify it as Blocker and treat it as an active incident: rotate immediately, do not wait for the pull request to merge.
- If a secret is in git history (even if deleted), it must still be rotated; the deletion commit does not remove the exposure.
- If a secret appears to be a test value, verify it has zero permissions in any real system before classifying it as Synthetic; do not dismiss based on name alone.
- If a secret is an environment variable passed to a container without a secrets manager, classify it as Medium; plain environment variables can be read from process listing in some configurations.
- If a CI/CD pipeline prints environment variables in a debug step, classify it as High; debug log output is often stored and accessible to pipeline participants.
- If no pre-commit hook or CI secrets scanner is in place, recommend one regardless of whether findings are present; absence of tooling is a process gap.
- If a secret rotation requires coordination with an external provider or stakeholder, document the dependencies and set a deadline.

## DevSecOps Guardrails

- Do not dismiss a finding as "probably a test key" without verifying the key has no permissions in any non-local system.
- Do not treat secret deletion as secret rotation; deleting a committed secret without rotating it leaves the exposed value valid.
- Do not commit .env files, application.properties with credentials, or any file with embedded secrets to VCS, even in a private repository.
- Do not pass secrets between pipeline jobs via artifacts unless the artifacts are encrypted and the keys are managed separately.
- Do not log authentication tokens, session identifiers, or connection strings at any log level in production or staging environments.
- Do not accept "we'll fix it later" for a Confirmed Exposure finding; rotation must happen before the pull request merges or immediately as an out-of-band fix.

## Output Requirements

- **Findings report**: each finding classified as Confirmed Exposure, Likely Exposure, or Synthetic Value; with file, line, secret type, access scope, and remediation urgency.
- **Rotation and revocation plan**: for each confirmed or likely exposure, the rotation steps, revocation mechanism, and who has authority to execute each.
- **Git history assessment**: confirmation that history was checked and whether historical exposure was found.
- **Secret storage assessment**: current storage mechanism for each discovered secret reference vs. recommended storage.
- **Rotation policy assessment**: secrets without rotation policies flagged with recommended policy.
- **Prevention recommendations**: specific tooling, process, or policy changes to prevent recurrence.

## Acceptance Criteria

- All confirmed secret exposures are rotated before the associated change merges or immediately as an incident response.
- All secret references in code source their values from an approved secrets manager or CI/CD secrets store.
- No .env files or credential files are committed to VCS.
- Git history is confirmed clear of secrets, or the exposure is in history and rotation is complete.
- Application log statements do not include tokens, passwords, or session identifiers.
- A pre-commit hook or CI secrets scanner gate is in place or recommended with a timeline.

## Anti-Patterns

- **Rotation avoidance**: Deleting the committed secret file without rotating the actual credential, leaving the exposed value valid.
- **Name-based dismissal**: Marking a secret as "test only" based on its variable name (TEST_API_KEY) without verifying it has no production access.
- **History blindness**: Scanning only the current branch tip without checking whether the secret was committed and removed in an earlier commit.
- **Debug log acceptance**: Accepting pipeline debug steps that print environment variables because "it's only in CI".
- **Plain env var complacency**: Treating plain environment variables as sufficiently secure for production secrets without a secrets manager.
- **Broad-scope acceptance**: Accepting a secret with admin or all-read permissions because "it's easier than creating a scoped token".
- **One-time scan**: Running a secrets scan only at the time of the pull request rather than as a continuous pre-commit and CI gate.

## Changelog

### {{.Version}} - {{.LastModified}}

- Initial generated DevSecOps SDLC skill.
`

const privacyReviewerBody = `
# {{.Title}}

## Purpose

Review features, services, and data flows for data-protection and privacy risks. Identify personal data processing activities, assess legal basis and data minimization, flag retention and logging risks, and produce concrete privacy-by-design recommendations. Output must include an actionable assessment of each personal data flow with legal basis, retention obligation, data-subject rights impact, and DPIA trigger assessment.

## When to Use This Skill

- A feature introduces new processing of personal data, behavioral data, location data, or authentication credentials.
- A data flow is added or modified that involves transferring personal data to a third party or across jurisdictions.
- A retention policy, logging configuration, or analytics pipeline is being changed.
- A Data Protection Impact Assessment (DPIA) is required or its necessity is being assessed.
- You are routed here by the central agent via $privacy-reviewer.

## Inputs

- Feature description, requirements, or design document.
- Data flow diagrams or descriptions showing what personal data is processed and where it flows.
- List of data elements processed: field names, data types, and data subjects.
- Current privacy notices and consent flows, if applicable.
- Applicable regulations and organizational data-protection policies (GDPR, CCPA, HIPAA, LGPD).
- Third-party processor agreements or DPA references.
- Logging and analytics configuration.

## Skill-Specific Operating Model

1. **Identify all personal data elements.** List every data element processed by the feature. Classify each as: ordinary personal data, special category data (health, biometric, political opinion, sexual orientation, religious belief, racial/ethnic origin), or pseudonymous data.
2. **Map data flows.** For each personal data element, trace the flow: collection point, processing system, storage location, retention period, deletion mechanism, and any sharing with third parties or cross-border transfers.
3. **Assess legal basis.** For each processing activity, identify the legal basis (consent, contract, legal obligation, vital interests, public task, legitimate interests). Flag processing activities with no identified legal basis.
4. **Assess data minimization.** For each data element, verify that it is necessary for the stated purpose. Flag data elements that are collected but not used, or used for a purpose broader than the stated purpose.
5. **Assess purpose limitation.** Check whether personal data is used only for the purpose for which it was collected. Flag secondary uses, analytics pipelines, and feature additions that repurpose existing data.
6. **Assess retention.** For each data element, verify that a retention period is defined and enforced. Flag data retained indefinitely or longer than the minimum necessary.
7. **Assess data-subject rights.** For each processing activity, assess the impact on: right of access, right to erasure, right to rectification, right to portability, right to object, and right to restrict processing.
8. **Assess consent mechanisms.** For processing based on consent, verify that consent is freely given, specific, informed, and unambiguous. Check that consent can be withdrawn as easily as it is given.
9. **Assess third-party sharing and cross-border transfers.** For each data transfer to a third party or across jurisdictions, verify that a data processing agreement or transfer mechanism exists.
10. **Assess DPIA necessity.** Check against DPIA trigger criteria: large-scale processing, sensitive data categories, profiling with legal effects, systematic surveillance, new technologies. Recommend a DPIA if any trigger criterion is met.
11. **Assess logging and analytics for privacy risks.** Check whether application logs, analytics pipelines, or monitoring systems capture personal data without explicit justification and retention controls.
12. **Produce privacy-by-design recommendations.** For each gap, propose a concrete design change that implements privacy by default.

## Skill-Specific Checklist

- [ ] Every personal data element processed by the feature is identified and classified (ordinary, special category, pseudonymous).
- [ ] Data flows are mapped for every personal data element: collection, processing, storage, retention, deletion, and sharing.
- [ ] Every processing activity has an identified legal basis.
- [ ] Processing activities without a legal basis are flagged as Blockers.
- [ ] Every data element is assessed for necessity (data minimization).
- [ ] Data elements collected but unused or used beyond stated purpose are flagged.
- [ ] Retention periods are defined and enforced for every personal data element.
- [ ] Indefinite retention or retention beyond legal minimum is flagged.
- [ ] Data-subject rights impact is assessed for each processing activity (access, erasure, portability, objection).
- [ ] Consent-based processing is assessed for validity: freely given, specific, informed, unambiguous, withdrawable.
- [ ] Every third-party data sharing has a data processing agreement reference.
- [ ] Cross-border transfers have an approved transfer mechanism (adequacy decision, SCCs, BCRs).
- [ ] Application logs are checked for personal data capture without justification and retention controls.
- [ ] DPIA necessity is assessed against trigger criteria.
- [ ] Privacy-by-design recommendations are provided for each identified gap.

## Decision Rules

- If a processing activity has no identified legal basis, classify it as a Critical privacy risk; processing without legal basis is a GDPR violation.
- If special category data (health, biometric, political opinion) is processed, verify that an additional legal basis under GDPR Art. 9 exists; absence is a Critical finding.
- If personal data is shared with a third party without a data processing agreement, classify it as High.
- If a cross-border transfer occurs without an approved transfer mechanism, classify it as High.
- If personal data is retained indefinitely without documented justification, classify it as Medium.
- If consent-based processing does not include a withdrawal mechanism of equal ease to consent granting, classify it as High.
- If any DPIA trigger criterion is met and a DPIA has not been initiated, recommend a DPIA and flag it as a blocking requirement before the feature launches.

## DevSecOps Guardrails

- Do not approve a feature that processes personal data without an identified legal basis.
- Do not approve a data flow that shares personal data with a third party without a data processing agreement.
- Do not accept indefinite retention as a default; every personal data element must have a defined and enforced retention period.
- Do not allow application logs to capture personal data (names, email addresses, IP addresses classified as personal in the applicable jurisdiction) without explicit justification, retention controls, and access restrictions.
- Do not treat pseudonymization as anonymization; pseudonymous data remains personal data under GDPR.
- Do not approve a cross-border transfer without a documented transfer mechanism.

## Output Requirements

- **Personal data inventory**: each data element with classification, processing purpose, and legal basis.
- **Data flow map**: per-element flows covering collection, processing, storage, retention, deletion, and third-party sharing.
- **Legal basis assessment**: each processing activity with legal basis and gap findings.
- **Data minimization assessment**: data elements flagged for over-collection or purpose drift.
- **Retention assessment**: defined retention periods, enforcement mechanism, and gaps.
- **Data-subject rights impact assessment**: rights impacted per processing activity with remediation.
- **Third-party and cross-border transfer assessment**: DPA and transfer mechanism status.
- **Logging privacy assessment**: personal data in logs with justification and retention gaps.
- **DPIA trigger assessment**: trigger criteria evaluated with recommendation.
- **Privacy-by-design recommendations**: concrete design changes per gap.

## Acceptance Criteria

- Every processing activity has an identified and documented legal basis.
- All special category data processing has an Art. 9 additional legal basis.
- All personal data elements have defined and enforced retention periods.
- All third-party data sharing has a data processing agreement.
- All cross-border transfers have a documented transfer mechanism.
- Data-subject rights are implementable for all processing activities.
- Application logs do not capture personal data without justification and retention controls.
- DPIA necessity is assessed and, if required, initiated before launch.

## Anti-Patterns

- **Legitimate interest as default**: Applying legitimate interests as the legal basis for all processing without conducting a balancing test.
- **Consent bundling**: Bundling consent for multiple processing purposes into a single checkbox, making consent not specific or granular.
- **Indefinite retention**: Storing personal data indefinitely because defining and enforcing a retention period requires engineering effort.
- **Pseudonymization as anonymization**: Treating pseudonymous data as outside the scope of data-protection obligations.
- **Log everything**: Capturing all request parameters including personal data in application logs for debugging convenience.
- **DPA assumption**: Assuming a data processing agreement exists because a vendor contract was signed, without verifying the DPA terms.
- **DPIA avoidance**: Structuring processing to avoid individual DPIA triggers by splitting a large-scale activity across multiple smaller ones.

## Changelog

### {{.Version}} - {{.LastModified}}

- Initial generated DevSecOps SDLC skill.
`

const releaseReadinessReviewerBody = `
# {{.Title}}

## Purpose

Assess whether a release is ready for deployment to production. Evaluate test results, security findings, rollback capability, database migration safety, monitoring coverage, runbook completeness, feature flag configuration, and support readiness. Produce a clear Go or No-Go recommendation with a prioritized list of blockers and a set of follow-up actions for non-blocking items.

## When to Use This Skill

- A release candidate is ready for production deployment.
- A hotfix needs to be assessed for emergency deployment.
- A release gate review is required before promotion to a regulated or production environment.
- You are routed here by the central agent via $release-readiness-reviewer.

## Inputs

- Release notes or change log for the release candidate.
- CI/CD pipeline results: test suite status, security scan results, SAST/DAST output.
- Open security findings from $secure-code-reviewer, $dependency-risk-reviewer, or $iac-security-reviewer.
- Deployment plan: migration scripts, rollout strategy, feature flag configuration.
- Monitoring dashboard and alert configuration for the affected services.
- Runbooks for the affected services.
- On-call and support readiness status.
- Known issues list and accepted risk register.

## Skill-Specific Operating Model

1. **Verify test status.** Confirm that all required test suites (unit, integration, security, E2E) pass. Flag any failing, skipped, or suppressed tests. Assess the coverage of the changes by the test suite.
2. **Review open security findings.** List all open security findings from prior reviews. Classify each as Blocker (must be resolved before release), High (must have accepted-risk decision), or Carry-over (tracked with a resolution plan and deadline).
3. **Assess rollback capability.** For every change in the release: is it independently rollback-capable? Does rolling back require downtime? Does rolling back undo data migrations? What is the maximum time to rollback?
4. **Assess database migration safety.** Check every migration script for: backward compatibility with the current application version, forward compatibility with the previous application version, reversibility, estimated execution time, and lock risks.
5. **Verify monitoring and alerting.** Check that new endpoints, services, and behaviors are covered by existing or new alerts. Verify that critical user journeys have SLO-based alerting and that on-call will be paged if the SLO is breached.
6. **Verify runbook completeness.** For each new service behavior, failure mode, or operational dependency introduced by the release, verify that a runbook section exists and is current.
7. **Assess feature flag configuration.** For features behind flags: verify that the flag is set to the correct value for the target environment, that the rollback plan includes flag-off as an option, and that the flag definition will be cleaned up after full rollout.
8. **Assess external dependency readiness.** Check whether any external service, third-party API, or infrastructure dependency required by the release has been confirmed ready.
9. **Assess support readiness.** Verify that support teams are informed of user-visible changes, that FAQ updates are ready, and that escalation paths are confirmed.
10. **Produce Go/No-Go decision.** State Go (ready for deployment), No-Go (blocked), or Conditional Go (ready with named conditions and owners). List every blocker explicitly.

## Skill-Specific Checklist

- [ ] All required test suites pass with no failing or suppressed tests.
- [ ] Test coverage of the changed code is assessed and gaps are documented.
- [ ] All Critical and High security findings are either resolved or have a documented accepted-risk decision with a named owner.
- [ ] Every change in the release has a documented rollback procedure stating steps, downtime requirement, and time estimate.
- [ ] Database migrations are backward-compatible with the running application version.
- [ ] Database migration execution time and lock impact are assessed; long-running locks are flagged.
- [ ] Monitoring alerts exist for all new endpoints and critical user journeys affected by the release.
- [ ] SLO alerting is configured and the on-call rotation is confirmed active.
- [ ] Runbooks cover all new failure modes and operational procedures introduced by the release.
- [ ] Feature flags are verified to be set to the correct value in the target environment.
- [ ] Flag cleanup is scheduled for fully-rolled-out features.
- [ ] External dependency readiness is confirmed (third-party APIs, infrastructure dependencies).
- [ ] Release notes are complete and user-visible changes are documented.
- [ ] Support team briefing is confirmed for user-visible changes.
- [ ] The deployment window, deployment owner, and rollback decision owner are named.

## Decision Rules

- If any Critical security finding is unresolved without an accepted-risk decision, the release is No-Go.
- If any required test suite has failures, the release is No-Go unless each failure has a documented justification and a named owner who accepts the risk.
- If a database migration is not backward-compatible with the current running application version, the release is No-Go until a compatible migration strategy is designed.
- If a database migration holds table locks for more than 60 seconds on a high-traffic table, flag it as High risk and require a confirmed maintenance window or online migration strategy.
- If no rollback procedure is documented for a change, the release is No-Go for that change.
- If monitoring alerts do not cover new critical user journeys, flag it as High and require either new alerts before release or an accepted-risk decision with a monitoring gap remediation timeline.
- If the deployment owner or rollback decision owner is not named, the release is Conditional Go until those roles are confirmed.

## DevSecOps Guardrails

- Do not approve a release where a Blocker security finding is unresolved without documented, named risk acceptance.
- Do not approve a release without a tested, documented rollback procedure for every schema-changing migration.
- Do not approve a release that disables or suppresses monitoring or alerting.
- Do not treat "tests pass in CI" as equivalent to "production-ready"; assess coverage and relevance of the test suite.
- Do not approve deployment without naming the rollback decision owner and the rollback time window.
- Do not skip the runbook review because the service is "simple"; operational teams need accurate runbooks for incident response.

## Output Requirements

- **Go/No-Go recommendation**: Go, No-Go, or Conditional Go with explicit rationale.
- **Blocker list**: each blocker with owner, required action, and deadline.
- **Security finding status**: each open finding with severity, resolution status, and accepted-risk owner.
- **Test coverage assessment**: test suite status, coverage of changed code, and gaps.
- **Rollback assessment**: rollback procedure, downtime requirement, and time estimate per change.
- **Migration safety assessment**: backward compatibility, lock impact, and execution time per migration script.
- **Monitoring and alerting assessment**: coverage gaps and SLO alerting status.
- **Runbook assessment**: completeness for new failure modes and operational procedures.
- **Follow-up action list**: non-blocking items with owner and deadline.

## Acceptance Criteria

- All required test suites pass.
- All Critical and High security findings are resolved or have documented risk acceptance.
- Every change has a documented rollback procedure.
- Database migrations are confirmed backward-compatible and lock-safe.
- Monitoring and SLO alerting cover all new critical user journeys.
- Runbooks cover all new failure modes introduced by the release.
- The deployment owner and rollback decision owner are named.
- A clear Go, No-Go, or Conditional Go recommendation is produced with explicit rationale.

## Anti-Patterns

- **Optimistic rollback**: Documenting "roll back by reverting the commit" without addressing whether data migrations are reversible or whether downtime is required.
- **Security finding carry-over**: Accepting all open security findings as carry-overs without assessing their severity relative to the release content.
- **Test pass theater**: Approving a release because CI is green without checking whether the tests cover the changed code paths.
- **Feature flag neglect**: Releasing with feature flags set to incorrect values for the target environment, causing silent activation or deactivation.
- **Runbook debt**: Skipping runbook updates because "there are no new failure modes", then failing to find the operational procedure during an incident.
- **No deployment owner**: Approving a release without naming the person responsible for the deployment decision and rollback authorization.
- **Migration lock blindness**: Approving a migration that holds a table lock on a high-traffic table without assessing the impact on production traffic.

## Changelog

### {{.Version}} - {{.LastModified}}

- Initial generated DevSecOps SDLC skill.
`

const incidentResponseHelperBody = `
# {{.Title}}

## Purpose

Support structured analysis and handling of security incidents and production outages. Guide triage, severity classification, containment, eradication, recovery, evidence preservation, timeline reconstruction, root cause analysis, and postmortem production. Separate confirmed facts from assumptions and open questions throughout the response. Preserve forensic evidence during active response. Produce a postmortem structure that enables blameless learning and prevents recurrence.

## When to Use This Skill

- A security incident, data breach, or suspected compromise is active or has been reported.
- A production outage or severe degradation requires structured incident analysis.
- A postmortem for a past incident needs to be drafted or reviewed.
- A preliminary root cause analysis is needed during an active incident.
- You are routed here by the central agent via $incident-response-helper.

## Inputs

- Incident description: who reported it, when, what was observed.
- Available signals: logs, alerts, metrics, error rates, audit trail entries.
- Systems and services affected or suspected to be affected.
- Timeline of events leading up to and following the incident.
- Recent changes: deployments, configuration changes, dependency updates, access changes.
- Responder list and communication channel for the incident.

## Skill-Specific Operating Model

1. **Triage: establish the known facts.** Separate confirmed observations (from logs, metrics, alerts) from inferred causes and assumptions. Do not conflate symptom with cause at this stage.
2. **Classify severity.** Assign a severity level using the organization's incident severity framework or a standard: SEV-1 (complete outage, confirmed data breach, critical security compromise), SEV-2 (significant degradation, suspected breach, customer-visible security issue), SEV-3 (partial degradation, security anomaly under investigation), SEV-4 (minor issue, investigation required).
3. **Assess impact.** Quantify impact where possible: affected users, affected services, data at risk, revenue impact, compliance implications. Flag any personal data exposure for immediate privacy team notification.
4. **Initiate containment.** Recommend the minimum containment actions needed to stop ongoing harm without destroying evidence. Containment actions must be reversible or have a documented risk of irreversibility.
5. **Preserve forensic evidence.** Before making changes to affected systems, capture: log snapshots, process lists, network connections, memory dumps if applicable, and disk snapshots. Evidence that is overwritten cannot be recovered.
6. **Eradication.** Once the attack vector, vulnerability, or failure mode is identified, recommend specific eradication actions: patch, configuration change, secret rotation, account revocation. Verify eradication is complete before recovery.
7. **Recovery.** Recommend recovery steps: restore from backup, redeploy clean version, re-enable service. Verify recovery is complete and the service is operating normally before declaring resolution.
8. **Reconstruct the timeline.** Build a chronological timeline: first event in logs, first detection, first response action, containment time, eradication time, recovery time. Separate events confirmed by log evidence from events inferred or reported by memory.
9. **Root cause analysis.** Apply a structured RCA method: Five Whys, fishbone diagram, or contributing factors analysis. Identify the immediate cause, the contributing causes, and the systemic or process factors that enabled the incident.
10. **Produce corrective actions.** For each root cause and contributing factor, produce a specific corrective action with an owner, a deadline, and a verifiable completion criterion. Prioritize actions that prevent recurrence over actions that only mitigate impact.
11. **Draft postmortem.** Structure the postmortem: incident summary, timeline, impact, root cause, contributing factors, what went well, what did not go well, corrective actions, and lessons learned. Write for a blameless audience.

## Skill-Specific Checklist

- [ ] Confirmed facts are separated from assumptions and open questions from the first message.
- [ ] Severity is classified with justification; a severity change is documented if the classification changes during the incident.
- [ ] Impact is quantified: affected users, services, data types, and compliance implications.
- [ ] Personal data exposure is flagged immediately for privacy team notification.
- [ ] Containment actions are documented with reversibility assessment before execution.
- [ ] Forensic evidence is captured before any system changes are made.
- [ ] The attack vector, vulnerability, or failure mode is identified before recovery begins.
- [ ] Eradication is verified complete before the recovery phase starts.
- [ ] Recovery is verified complete before the incident is declared resolved.
- [ ] A chronological timeline is built with each event's evidence source noted (log reference, alert, memory).
- [ ] RCA identifies the immediate cause, contributing causes, and systemic factors.
- [ ] Corrective actions each have an owner, a deadline, and a verifiable completion criterion.
- [ ] The postmortem is blameless: findings focus on systems, processes, and tooling, not individual error.
- [ ] "What went well" is documented alongside "what did not go well".
- [ ] Follow-up issues are filed and linked from the postmortem.

## Decision Rules

- If personal data is confirmed or suspected to have been exposed, notify the privacy team immediately; do not wait for full RCA before notification.
- If containment actions would destroy forensic evidence, preserve evidence first, then contain; document the trade-off.
- If the attack vector is unknown, do not begin recovery; recovery into a compromised environment re-exposes the system.
- If a system must be taken offline for containment, confirm the impact on dependent services before action and document the dependency chain.
- If secrets, credentials, or tokens may have been exposed, treat them as compromised and initiate rotation immediately, even if exposure is unconfirmed.
- If RCA reveals a systemic process failure (missing monitoring, no incident runbook, insufficient alerting), add a corrective action addressing the systemic cause, not only the technical cause.
- If a corrective action requires significant engineering effort, document a risk-accepted interim mitigation with a deadline for the full fix.

## DevSecOps Guardrails

- Do not speculate about root cause in external or customer communications; use confirmed facts only.
- Do not make system changes during an active incident without logging the action, the reason, and the person who authorized it.
- Do not declare the incident resolved until recovery is verified by monitoring data, not by the responder's assessment.
- Do not assign blame for individual errors in the postmortem; focus on the systems, processes, and conditions that allowed the error to have impact.
- Do not include actual secret values, credentials, or personal data in incident or postmortem documents.
- Do not close follow-up corrective action items without a verifiable completion criterion; "done" must be verifiable.

## Output Requirements

- **Incident summary**: confirmed facts, affected systems, severity, impact assessment.
- **Confirmed timeline**: chronological events with evidence source for each entry.
- **Assumptions and open questions log**: each open question with owner and urgency.
- **Containment actions**: actions taken, reversibility assessment, and execution authorization.
- **Forensic evidence inventory**: what was captured, when, and where it is stored.
- **Eradication plan**: identified attack vector or failure mode, specific eradication steps, and verification approach.
- **Recovery plan**: recovery steps, monitoring verification, and resolution criteria.
- **Root cause analysis**: immediate cause, contributing causes, systemic factors.
- **Corrective actions register**: each action with owner, deadline, and completion criterion.
- **Postmortem draft**: summary, timeline, impact, RCA, what went well, what did not, corrective actions, lessons learned.

## Acceptance Criteria

- Confirmed facts are separated from assumptions throughout the incident lifecycle.
- Severity is classified with documented justification.
- Personal data exposure is flagged for privacy notification before full RCA is complete.
- Forensic evidence is captured before system changes are made during containment.
- Eradication is verified complete before recovery begins.
- RCA identifies systemic factors, not only technical root cause.
- Every corrective action has an owner, a deadline, and a verifiable completion criterion.
- The postmortem is blameless and documents both successes and failures.

## Anti-Patterns

- **Symptom-as-cause**: Identifying the immediate technical symptom (service crashed) as the root cause without investigating why it crashed and why the crash had that impact.
- **Blame assignment**: Writing a postmortem that identifies an individual as responsible rather than the system conditions that allowed the error to propagate.
- **Recovery before eradication**: Restoring service before confirming the attack vector or failure mode is resolved, re-exposing the system.
- **Evidence destruction**: Making system changes (log rotation, redeployment, instance termination) before capturing forensic evidence.
- **Speculation in communications**: Including unverified hypotheses in external or stakeholder communications that later prove incorrect.
- **Corrective action theater**: Writing corrective actions that are too vague to be verifiable ("improve monitoring") or that are already done before the postmortem is written.
- **Incident fatigue normalization**: Treating recurring low-severity incidents as acceptable background noise rather than signals of systemic problems.

## Changelog

### {{.Version}} - {{.LastModified}}

- Initial generated DevSecOps SDLC skill.
`

const complianceEvidenceCollectorBody = `
# {{.Title}}

## Purpose

Collect, structure, and assess compliance evidence for audits, internal control reviews, and certification assessments. Map observed technical and process artifacts to specific compliance control requirements. Identify evidence gaps, verify evidence quality (completeness, timeliness, source authority), and produce an audit-ready evidence package. Every evidence item must have a source, a timestamp, and a named owner.

## When to Use This Skill

- An audit, certification renewal, or internal control review is approaching.
- A compliance control needs to be demonstrated to an auditor.
- A change is being assessed for its impact on compliance posture.
- Evidence gaps from a prior audit need to be remediated.
- You are routed here by the central agent via $compliance-evidence-collector.

## Inputs

- List of applicable controls from the relevant compliance framework (SOC 2, ISO 27001, PCI DSS, HIPAA, GDPR, NIST CSF, etc.).
- Evidence requests or audit questionnaire from an external or internal auditor.
- Repository artifacts: CI/CD pipeline results, deployment logs, access review records, vulnerability scan reports.
- Policy documents, procedures, and runbooks.
- Change management records, approval logs, and incident records.
- Access control and permission audit exports.

## Skill-Specific Operating Model

1. **Map controls to evidence types.** For each required control, identify the evidence type needed: technical evidence (logs, scan results, configuration exports), process evidence (approvals, meeting minutes, training records), or documentation evidence (policy, procedure, architecture diagram).
2. **Inventory available evidence.** For each evidence type, identify what is available, its source, its timestamp, and whether it covers the full control period required by the audit.
3. **Assess evidence quality.** For each evidence item, assess: completeness (does it cover the full scope?), timeliness (is it current enough for the audit period?), integrity (is it from an authoritative, tamper-evident source?), and relevance (does it directly demonstrate the control, not just an adjacent activity?).
4. **Identify evidence gaps.** List controls with missing, stale, or insufficient evidence. For each gap, state what evidence is needed, who can provide it, and by when.
5. **Collect technical evidence.** For technical controls: export access control configurations, pull CI/CD pipeline results, export security scan reports, retrieve audit log excerpts covering the review period.
6. **Collect process evidence.** For process controls: retrieve change approval records, access review completion records, training completion records, incident response records, and vendor assessment records.
7. **Verify control ownership.** For each control, verify that a named control owner exists, that they are aware of their ownership, and that they can attest to the control's operating effectiveness.
8. **Structure the evidence package.** Organize evidence by control, with each item labeled: Control ID, Control Description, Evidence Type, Evidence Source, Evidence Date Range, Owner, and Gap Status (evidenced/partial/missing).
9. **Assess continuous evidence.** For controls that require continuous evidence over a period (log retention, vulnerability scanning cadence, change management process), verify that evidence is available for the entire audit period, not only point-in-time.
10. **Produce an evidence gap register.** Table listing controls with missing or insufficient evidence, the specific gap, the evidence owner, and the collection deadline.

## Skill-Specific Checklist

- [ ] All required controls from the applicable framework are listed with their evidence requirements.
- [ ] Each control has a named owner who can attest to its operating effectiveness.
- [ ] Evidence is available for the full audit period, not only the current date.
- [ ] Technical evidence is from an authoritative, tamper-evident source (system-generated logs, not manually exported spreadsheets).
- [ ] Process evidence includes approval records with timestamps and named approvers, not just policy documents.
- [ ] Access control evidence covers all privileged access, service accounts, and third-party access.
- [ ] Vulnerability management evidence demonstrates that findings are tracked, prioritized, and remediated within policy-defined timeframes.
- [ ] Change management evidence covers all production changes during the review period with approval records.
- [ ] Incident response evidence covers all incidents during the review period with timeline and remediation.
- [ ] Training completion records cover all required personnel for the review period.
- [ ] Third-party and vendor security assessment evidence is current (not older than the review period).
- [ ] Log retention evidence confirms that logs were retained for the required period.
- [ ] Evidence gaps are documented with owner and collection deadline.
- [ ] The evidence package distinguishes preventive controls from detective controls from corrective controls.
- [ ] The evidence package is organized by control, not by evidence type, for auditor usability.

## Decision Rules

- If a control owner cannot be named, flag it as a Critical governance gap; controls without owners cannot be demonstrated as operating effectively.
- If evidence covers only a point in time (e.g., one access review export) but the control requires continuous operation evidence, flag it as a gap and request additional evidence covering the full period.
- If a control relies on a manual process that has no automated log or record, flag it as Medium risk and recommend adding audit trail capture to the process.
- If vulnerability scan evidence shows findings open beyond the policy remediation deadline, the control is not operating effectively; document the gap and the remediation plan.
- If third-party vendor security assessment evidence is older than the audit period, it does not satisfy the control; flag it and request a current assessment.
- If a required policy document references a procedure that does not exist, flag the missing procedure as a gap; a policy without an implemented procedure does not evidence a control.

## DevSecOps Guardrails

- Do not include personal data, passwords, production credentials, or secret values in evidence packages.
- Do not produce evidence artifacts that selectively cover only periods when the control was operating correctly; auditors require continuous evidence.
- Do not accept manually edited exports as evidence of system-generated controls; evidence must be from the authoritative system.
- Do not submit evidence for controls that were only recently implemented if the audit period pre-dates the implementation; document the gap and the implementation date.
- Do not conflate policy existence with control implementation; a policy document is not evidence that the control operates effectively.

## Output Requirements

- **Control-to-evidence mapping**: each control with required evidence type, available evidence item, source, date range, and gap status.
- **Evidence quality assessment**: completeness, timeliness, integrity, and relevance per evidence item.
- **Evidence gap register**: controls with missing or insufficient evidence, specific gap description, evidence owner, and collection deadline.
- **Control ownership register**: each control with named owner and attestation status.
- **Continuous evidence assessment**: controls requiring period-continuous evidence with coverage status.
- **Evidence package structure**: organized by control ID for auditor navigation.
- **Summary**: percentage of controls with complete evidence, percentage with gaps, and critical risks.

## Acceptance Criteria

- All required controls have a named owner.
- Evidence covers the full audit period for continuous controls.
- All evidence is from authoritative, tamper-evident sources.
- Technical and process evidence are distinguished and both present for applicable controls.
- Evidence gaps are documented with owner and collection deadline.
- No personal data, credentials, or secrets appear in the evidence package.
- The evidence package is organized by control for auditor usability.

## Anti-Patterns

- **Point-in-time evidence for continuous controls**: Submitting a single access review export for a control that requires quarterly access reviews throughout the audit period.
- **Policy-as-evidence**: Submitting a policy document as evidence that a control operates effectively, without demonstrating implementation.
- **Retrospective control creation**: Implementing a control only when an audit is announced and submitting evidence that does not cover the full audit period.
- **Manual evidence fabrication**: Manually creating exports or records that should be system-generated, undermining the integrity of the evidence.
- **Missing control owners**: Treating compliance evidence collection as a documentation exercise without assigning operational ownership to each control.
- **Selective evidence selection**: Choosing evidence periods when performance was best rather than providing representative continuous evidence.
- **Vendor trust without verification**: Accepting vendor-provided security documentation without verifying currency, scope, and applicability to the shared responsibility model.

## Changelog

### {{.Version}} - {{.LastModified}}

- Initial generated DevSecOps SDLC skill.
`

const policyAsCodeReviewerBody = `
# {{.Title}}

## Purpose

Review policy-as-code rules—OPA/Rego policies, Kyverno policies, GitLab Compliance Frameworks, GitHub CODEOWNERS, and Kubernetes admission policies—for correctness, coverage, enforcement mode, test completeness, and bypass risks. Verify that each policy enforces its stated intent under both normal and adversarial inputs. Identify coverage gaps, unsafe exceptions, weak enforcement modes, and missing test cases.

## When to Use This Skill

- An OPA/Rego policy, Kyverno policy, or Kubernetes admission webhook configuration is added or modified.
- A GitLab Compliance Framework, GitHub branch protection rule, or CODEOWNERS file is changed.
- Policy test suites are being reviewed for completeness.
- A policy enforcement audit is required before promoting to a regulated environment.
- You are routed here by the central agent via $policy-as-code-reviewer.

## Inputs

- Policy definitions: Rego files, Kyverno YAML, ValidatingWebhookConfiguration, MutatingWebhookConfiguration.
- Policy test files: OPA unit tests (rego_test.go, _test.rego), Kyverno test cases.
- Enforcement mode configuration: audit vs. enforce vs. warn.
- Policy exception definitions and exception owners.
- CI/CD integration: where policies are evaluated in the pipeline.
- Previous policy audit findings, if available.

## Skill-Specific Operating Model

1. **Clarify policy intent.** For each policy, state in plain language what it is intended to allow, what it is intended to deny, and what the security or compliance objective is.
2. **Verify allow/deny logic.** Read the Rego or policy rule logic and verify that it correctly implements the stated intent. Check for off-by-one errors, missing negations, incorrect set operations, and incorrect data comparisons.
3. **Check for coverage gaps.** Identify inputs that the policy does not cover: edge cases, empty inputs, partial inputs, inputs with fields in unexpected formats, and inputs from non-standard sources.
4. **Review enforcement mode.** Verify that the policy is in the correct enforcement mode for its purpose. Audit mode only logs violations without blocking; it is not a substitute for enforce mode in production.
5. **Review exception definitions.** For each exception or exclusion in the policy, verify: who authorized it, what the business justification is, whether the exception has an expiry date, and whether the exception itself creates a bypass risk.
6. **Review test coverage.** Verify that test cases exist for: every explicit allow case, every explicit deny case, every exception, and at least one edge case (empty input, malformed input, boundary input). Flag policies with tests only for the happy path.
7. **Test for bypass risk.** For each policy, assess whether an attacker or misconfigured system could craft an input that satisfies the allow rules while violating the security intent: field confusion, label overrides, namespace selector misuse, wildcard abuse.
8. **Verify rollout safety.** For enforce-mode policies being introduced into an existing cluster or pipeline, verify that existing compliant workloads are not blocked and that a dry-run or audit-first rollout is planned.
9. **Assess auditability.** Check that policy violations are logged with sufficient context: which resource, which rule, which input field triggered the violation. Verify that logs are accessible to the security team.
10. **Verify policy lifecycle.** Check that deprecated policies have a replacement, that active policies have an owner, and that exception expiry dates are tracked and enforced.

## Skill-Specific Checklist

- [ ] Policy intent is stated in plain language: what is allowed, what is denied, and what objective it serves.
- [ ] Allow and deny logic is verified to match the stated intent.
- [ ] Missing negations or incorrect set operations in Rego logic are checked.
- [ ] Coverage gaps are assessed: edge cases, empty inputs, partial inputs, and non-standard input formats.
- [ ] Enforcement mode is appropriate: enforce for hard requirements, audit for monitoring-only, warn for gradual rollout.
- [ ] Audit-mode-only policies are not presented as security controls.
- [ ] Every exception has a documented business justification, an authorizing owner, and an expiry or review date.
- [ ] Exceptions are not so broad that they negate the policy for the cases it most needs to cover.
- [ ] Test cases cover: every allow case, every deny case, every exception, and at least one edge case.
- [ ] Tests verify both the positive (allow) and negative (deny) outcome of each rule, not only the happy path.
- [ ] Bypass scenarios are assessed: field spoofing, label override, namespace selector misuse, wildcard abuse.
- [ ] Rollout strategy is assessed for enforce-mode policies: dry-run, audit-first, phased enforcement.
- [ ] Policy violations are logged with resource identity, rule name, and violated field.
- [ ] Policy ownership is assigned; no active policy is unowned.
- [ ] Exception expiry dates are tracked and have an enforcement mechanism.

## Decision Rules

- If a policy is in audit mode for a security control that must be enforced (privileged containers, hostNetwork access, root filesystem), classify it as High; audit mode does not block violations.
- If a policy exception is defined without a documented business justification and an expiry date, classify it as Medium and require documentation before the policy is merged.
- If no deny test case exists for a policy, classify it as High; a policy with only allow tests cannot be verified to deny violations correctly.
- If a bypass scenario exists where an input satisfies the allow rules but violates the security intent, classify it as Critical and require the policy logic to be corrected.
- If a policy is being deployed to enforce on an existing cluster without a dry-run phase, classify it as High risk; undocumented policy enforcement can cause unexpected workload disruption.
- If a policy has no owner, classify it as Medium; unowned policies are not maintained, reviewed, or updated when the environment changes.
- If exception expiry is not tracked or enforced, classify it as Medium; permanent exceptions accumulate and can negate the policy over time.

## DevSecOps Guardrails

- Do not classify an audit-mode policy as a security control; it observes but does not enforce.
- Do not merge a policy with test cases only for the happy path; deny logic must be explicitly tested.
- Do not approve an exception without a documented business justification, an authorizing owner, and an expiry or review date.
- Do not deploy enforce-mode policies to an existing cluster or pipeline without a dry-run or audit-first phase to identify unexpected impacts on compliant workloads.
- Do not accept "the policy is simple, tests are not needed" as a justification; even simple policies have edge cases that tests expose.
- Do not allow policy exceptions that are broader than the specific workload or namespace they are intended to exempt.

## Output Requirements

- **Policy intent summary**: plain-language statement of what each policy allows, denies, and what objective it serves.
- **Logic correctness assessment**: analysis of the allow/deny logic with any errors or mismatches identified.
- **Coverage gap assessment**: edge cases and input patterns not covered by the policy logic.
- **Enforcement mode assessment**: current mode, required mode, and gap.
- **Exception review**: each exception with justification status, owner, and expiry assessment.
- **Test coverage assessment**: allow cases, deny cases, exception cases, and edge cases covered vs. missing.
- **Bypass risk assessment**: identified bypass scenarios with severity and recommended fix.
- **Rollout safety assessment**: compliant workload impact analysis and rollout strategy.
- **Findings table**: Severity, Policy, Rule, Finding, Recommended Fix.

## Acceptance Criteria

- Policy intent is documented in plain language.
- Allow and deny logic is verified to correctly implement the stated intent.
- Test cases cover every explicit allow rule, every deny rule, every exception, and at least one edge case.
- No identified bypass scenario allows a violation of security intent while satisfying the allow rules.
- Every exception has a documented justification, owner, and expiry date.
- Enforce-mode policies have a dry-run or audit-first rollout plan for existing environments.
- Policy violations produce logs with sufficient context for security team triage.

## Anti-Patterns

- **Audit mode as enforcement**: Deploying a policy in audit mode for a security control and reporting it as enforced.
- **Happy-path-only tests**: Writing test cases only for inputs that should be allowed, without testing the deny cases the policy is designed to catch.
- **Permanent exceptions**: Adding policy exceptions without expiry dates, allowing them to accumulate and progressively undermine the policy.
- **Overly broad exceptions**: Excepting entire namespaces or label selectors that cover far more workloads than the exception was intended for.
- **Bypass blindness**: Writing policy logic without assessing how an attacker could craft an input that satisfies the allow rules while violating the security intent.
- **Unowned policies**: Merging policies without an assigned owner responsible for maintenance, testing, and updates when the environment changes.
- **Direct enforce rollout**: Switching a policy from audit to enforce in an existing cluster without a dry-run phase to identify unexpected impacts on compliant workloads.

## Changelog

### {{.Version}} - {{.LastModified}}

- Initial generated DevSecOps SDLC skill.
`

const observabilityReadinessReviewerBody = `
# {{.Title}}

## Purpose

Review the observability configuration—logging, metrics, distributed tracing, alerting, SLOs, dashboards, and runbooks—to verify that critical user journeys and operational risks are observable and actionable before and after deployment. Identify missing alerts, untraced operations, sensitive data in logs, undefined SLOs, and runbooks that do not cover current failure modes. Produce a gap analysis with concrete recommendations that connect observability to release readiness and incident response capability.

## When to Use This Skill

- A new service or feature is being released and its observability coverage is being assessed.
- Existing alerting or logging is being reviewed after an incident.
- SLOs are being defined, revised, or assessed for coverage.
- $release-readiness-reviewer has flagged observability gaps as a release blocker.
- You are routed here by the central agent via $observability-readiness-reviewer.

## Inputs

- Service description and critical user journey definitions.
- Logging configuration and log output samples.
- Metrics exported by the service (Prometheus metrics, CloudWatch metrics, custom metrics).
- Alert definitions (Prometheus alerting rules, PagerDuty rules, Grafana alert configuration).
- SLO definitions and error budget configuration.
- Dashboard configuration (Grafana, Datadog, CloudWatch dashboards).
- Runbook documents for the service.
- Recent incident history relevant to the service, if available.

## Skill-Specific Operating Model

1. **Identify critical user journeys.** List every critical user journey supported by the service. For each journey, state the observable signals that indicate whether it is succeeding or failing.
2. **Assess logging completeness.** Verify that critical operations emit structured logs with: operation name, outcome (success/failure), relevant identifiers, duration, and error detail for failures. Check that log levels are appropriate and that debug logs are not enabled in production by default.
3. **Assess sensitive data in logs.** Review log field names and sample output for: passwords, tokens, session identifiers, API keys, full request bodies containing PII, and full response bodies containing sensitive data. Flag any sensitive data found.
4. **Assess metrics coverage.** For each critical user journey, verify that the four golden signals are measurable: latency (p50, p95, p99), traffic (request rate), errors (error rate by type), and saturation (resource utilization). Flag journeys where any golden signal is unobservable.
5. **Assess SLO definition.** For each critical user journey, verify that an SLO exists with a defined SLI, target percentage, and measurement window. Verify that the SLI is derived from actual user experience signals, not from internal health checks alone.
6. **Assess error budget policy.** Verify that an error budget policy exists: what happens when the error budget is exhausted (halt releases, increase reliability work, incident response). Flag services with SLOs but no error budget policy.
7. **Assess alerting coverage and quality.** For each SLO, verify that a burn rate alert fires before the error budget is exhausted. Check alert routing: who is paged, when, and via which channel. Assess alert fatigue risk: are there high-volume low-signal alerts that suppress on-call attention to real issues.
8. **Assess distributed tracing coverage.** For services in a distributed system, verify that traces propagate across service boundaries, that trace sampling is appropriate (not zero), and that slow or erroring traces are identifiable without requiring log correlation.
9. **Assess dashboard usability.** Verify that dashboards exist for each SLO and that they are used for incident triage (not only retrospective analysis). Check that the on-call dashboard shows the current error budget burn rate prominently.
10. **Assess runbook completeness.** For each alert, verify that a runbook section exists that describes: what the alert means, how to investigate, what the likely causes are, how to mitigate, and who to escalate to. Runbooks must reflect current system architecture.
11. **Assess audit log coverage.** For security-relevant operations (authentication, authorization changes, data access, configuration changes), verify that audit logs are produced, retained, and not mixed with application debug logs.
12. **Connect to release readiness.** For each observability gap, state whether it is a release blocker (critical journey is unobservable, SLO alert is missing) or a follow-up item (dashboard improvement, runbook update).

## Skill-Specific Checklist

- [ ] Every critical user journey has defined observable success and failure signals.
- [ ] Structured logging is in place for all critical operations with outcome, identifiers, and duration.
- [ ] Log output does not contain passwords, tokens, session IDs, API keys, or PII.
- [ ] Log levels are appropriate; debug logging is not enabled in production by default.
- [ ] All four golden signals (latency, traffic, errors, saturation) are measurable for each critical journey.
- [ ] p50, p95, and p99 latency metrics are exported for each critical operation.
- [ ] Error rates are broken down by error type, not just total error count.
- [ ] SLOs are defined for every critical user journey with an SLI, target, and measurement window.
- [ ] SLIs are derived from user-facing signals, not only internal health checks.
- [ ] Error budget policy is defined: what actions are triggered when the budget is exhausted.
- [ ] Burn rate alerts are configured to page before the error budget is exhausted.
- [ ] Alert routing is verified: on-call is paged via the correct channel with the correct severity.
- [ ] High-volume low-signal alerts are identified and assessed for alert fatigue risk.
- [ ] Distributed traces propagate across service boundaries with non-zero sampling rate.
- [ ] On-call dashboard shows current SLO status and error budget burn rate prominently.
- [ ] Every active alert has a corresponding runbook section with investigation steps, causes, mitigation, and escalation path.
- [ ] Runbooks reflect current system architecture and have been updated after recent changes.
- [ ] Audit logs cover authentication, authorization changes, data access, and configuration changes.
- [ ] Audit logs are retained separately from application debug logs with appropriate access controls.

## Decision Rules

- If a critical user journey has no SLO and no latency or error rate alert, classify it as High; the journey is invisible to the on-call team during an incident.
- If application logs contain passwords, tokens, or session identifiers, classify it as Critical and treat it as a data exposure finding requiring immediate remediation.
- If SLOs exist but no burn rate alert is configured, classify it as High; SLOs without burn rate alerts do not generate timely pages.
- If an alert has no corresponding runbook section, classify it as Medium; an alert that cannot be triaged without investigation from scratch increases mean time to resolution.
- If distributed trace sampling is zero percent or if traces do not cross service boundaries, classify it as Medium; inter-service latency and errors are invisible.
- If runbooks reference components or procedures that no longer exist in the current architecture, classify it as Medium; incorrect runbooks cause responders to take wrong actions.
- If the error budget policy is undefined, classify it as Medium; teams without a policy default to ignoring error budget burn.

## DevSecOps Guardrails

- Do not approve a service release without SLO-based alerting for every critical user journey.
- Do not accept logs that contain authentication tokens, session identifiers, or personal data without explicit justification and retention controls.
- Do not treat the existence of dashboards as equivalent to operational readiness; verify that dashboards are used during incident response.
- Do not accept SLIs derived only from internal health checks; SLIs must reflect actual user experience.
- Do not accept runbooks that have not been updated after a significant architecture change.
- Do not conflate high alert volume with good observability; alert fatigue reduces the effectiveness of all alerting.

## Output Requirements

- **Critical user journey map**: each journey with observable success and failure signals.
- **Logging assessment**: structured logging coverage, log level configuration, and sensitive data findings.
- **Metrics coverage assessment**: golden signals per critical journey with gaps.
- **SLO assessment**: SLI definition, target, measurement window, and SLI-to-user-experience alignment per SLO.
- **Error budget policy assessment**: policy existence, trigger criteria, and gaps.
- **Alerting assessment**: burn rate alerts, routing, alert fatigue risk, and coverage gaps.
- **Distributed tracing assessment**: cross-service propagation, sampling rate, and usability.
- **Runbook assessment**: coverage per alert, currency (reflects current architecture), and usability.
- **Audit log assessment**: coverage of security-relevant events, retention, and access control.
- **Release readiness impact**: observability gaps classified as release blockers vs. follow-up items.

## Acceptance Criteria

- Every critical user journey has an SLO with a user-facing SLI, a target, and a measurement window.
- Burn rate alerts are configured for every SLO and page the on-call team via the correct channel.
- Application logs do not contain authentication tokens, passwords, session identifiers, or PII.
- All four golden signals are measurable for every critical user journey.
- Every active alert has a corresponding runbook section with investigation steps and escalation path.
- Distributed traces propagate across service boundaries with a non-zero sampling rate.
- Audit logs cover authentication, authorization changes, and data access with appropriate retention and access controls.

## Anti-Patterns

- **Health-check SLIs**: Defining SLIs based on whether the /health endpoint returns 200, rather than on whether actual user operations succeed.
- **Alert without runbook**: Configuring alerts that page on-call without providing any guidance on investigation, likely causes, or mitigation.
- **Sensitive log data**: Logging full request or response bodies containing tokens, passwords, or PII for debugging convenience.
- **Alert fatigue creation**: Configuring dozens of low-signal, high-frequency alerts that train on-call to suppress notifications.
- **Dashboard-only observability**: Creating dashboards for retrospective analysis but not designing the on-call workflow around them for real-time incident response.
- **Zero trace sampling**: Disabling distributed trace sampling in production to reduce storage cost, losing the ability to diagnose inter-service latency and error propagation.
- **Stale runbooks**: Allowing runbooks to reference decommissioned components, outdated procedures, or previous architectural patterns after a system has been significantly changed.

## Changelog

### {{.Version}} - {{.LastModified}}

- Initial generated DevSecOps SDLC skill.
`
