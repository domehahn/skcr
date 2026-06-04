# Skill Creator (`skcr`)

`skcr` ist eine Go-CLI zum Generieren von agentischen Projekt-Templates für:

- Codex
- GitLab Duo
- GitHub Copilot
- Claude Code
- OpenHands
- OpenCode
- Ollama

## Voraussetzungen

- Go `>= 1.23`
- `make` (optional, für `make build` und `make install`)

## Installation

### 1. Lokal aus dem Repo installieren (wie bei `sctl`)

```bash
make install
```

Das installiert `skcr` nach `$(go env GOBIN)` oder alternativ nach `$(go env GOPATH)/bin`.

### 2. Direkt mit Go (lokales Repo)

```bash
go install ./cmd/skcr
```

### 3. Global über Modulpfad (z. B. GitHub)

```bash
go install github.com/skcr/cmd/skcr@latest
```

### 4. Nur Binary bauen

```bash
make build
```

Ergebnis: `dist/skcr`

## Verifikation

```bash
skcr version
skcr --help
```

## Schnellstart

```bash
skcr init --target /path/to/repo --project-name MyProject
skcr bake default --target /path/to/repo --plan
skcr bake default --target /path/to/repo --write
skcr validate --target /path/to/repo
```

## Typische Commands

- `skcr init`: erstellt `agentic.bake.yaml`
- `skcr list-targets`: listet Targets aus `agentic.bake.yaml`
- `skcr bake`: rendert Dateien als Plan oder schreibt sie mit `--write`
- `skcr validate`: prüft Konsistenz und Plattform-Outputs
- `skcr version`: zeigt Version, Commit und Build-Zeit

## Häufige Workflows

### Neues Projekt initialisieren

```bash
skcr init --target . --project-name MyProject
skcr bake default --target . --write
```

### Nur bestimmte Plattformen aktivieren

```bash
skcr init --target . --platform "gitlab-duo,codex,github-copilot" --project-name MyProject
skcr bake default --target . --write
```

### Preset verwenden

```bash
skcr init --target . --preset all --project-name MyProject
```
