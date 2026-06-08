# Skill/Agentic Template Creator (`skcr`)

`skcr` ist eine Go-CLI zum Generieren von agentischen Projekt- und Plattform-Templates. `skcr` rendert Dateien aus `agentic.bake.yaml`, schreibt `.agentic-template.lock` und kann installierte Skills aus `skpm` nur lesen.

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
go install github.com/domehahn/skcr/cmd/skcr@latest
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
- `skcr clean`: entfernt nur von `skcr` verwaltete Dateien aus `.agentic-template.lock`
- `skcr scaffold skill <name>`: erstellt ein versionierbares Skill-Skeleton
- `skcr version`: zeigt Version, Commit und Build-Zeit

## Abgrenzung zu `skpm`

`skcr` besitzt:

- `agentic.bake.yaml`
- `.agentic-template.lock`
- generierte Plattformdateien
- Rendering-Templates und Plattform-Mappings

`skcr` darf lesen:

- `agent-skills.lock`
- installierte Skill-Verzeichnisse
- `skill.yaml`
- `SKILL.md`

`skcr` schreibt niemals `agent-skills.yaml`, `agent-skills.lock`, Registry-Konfigurationen oder Paketmanager-Caches. Skill-Auflösung, Installation, Updates, Publishing und Verifikation bleiben Aufgabe von `skpm`.

## Skill-Scaffolding

`skcr` kann neue Skill-Strukturen erzeugen. Danach übernimmt `skpm` den Lifecycle wie Validierung, Versionierung, Packaging und Publishing.

```bash
skcr scaffold skill secure-code-review \
  --version 0.1.0 \
  --description "Security-focused code review skill" \
  --owner platform-engineering \
  --platform codex \
  --platform claude-code \
  --platform gitlab-duo
```

Erzeugte Struktur:

```text
secure-code-review/
├── SKILL.md
├── skill.yaml
├── VERSION
├── CHANGELOG.md
├── README.md
├── LICENSE
└── tests/
    └── README.md
```

Wichtige Flags:

- `--output-dir <path>`: Zielordner für das Skill-Verzeichnis
- `--version <semver>`: Startversion, Standard `0.1.0`
- `--description <text>`: Beschreibung für `skill.yaml`
- `--owner <owner>`: Skill Owner
- `--platform <platform>`: kompatible Plattform, mehrfach nutzbar
- `--license <license>`: Lizenz, Standard `MIT`
- `--force`: vorhandene Scaffold-Dateien überschreiben
- `--dry-run`: geplante Dateien anzeigen, ohne zu schreiben

Gültige Skill-Namen bestehen nur aus Kleinbuchstaben, Ziffern und Bindestrichen, ohne führenden oder abschließenden Bindestrich.

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

### Empfohlener Workflow mit `skpm`

```bash
# 1. Skills mit skpm verwalten
skpm init
skpm add secure-code-review@^1.0.0
skpm lock
skpm install
skpm verify

# 2. Plattformdateien mit skcr rendern
skcr init --platform codex,claude-code --project-name MyProject
skcr bake default --skills-from agent-skills.lock --plan
skcr bake default --skills-from agent-skills.lock --write
skcr validate --against-lock agent-skills.lock
```

## Skill-Integration

`agentic.bake.yaml` kann eine optionale Skill-Integration enthalten:

```yaml
skills:
  source: agent-skills.lock
  mode: reference
  platforms:
    - codex
    - claude-code
    - gitlab-duo
```

Modi:

- `reference`: generierte Agent-Dateien referenzieren installierte Skill-Pfade. Das ist der Standard, weil `skpm` den Installationszustand besitzt.
- `copy`: kopiert installierte `SKILL.md`-Dateien in plattformspezifische Ausgabeordner.
- `link`: erstellt Symlinks, wo das Dateisystem es unterstützt.
- `embed`: stellt Skill-Metadaten im Render-Kontext bereit, ohne installierte Skill-Dateien zu übernehmen.

CLI-Beispiele:

```bash
skcr bake default --skills-from agent-skills.lock --skills-mode reference --plan
skcr bake default --skills-from agent-skills.lock --skills-mode copy --write
skcr list-targets --with-skills --skills-from agent-skills.lock
skcr validate --against-lock agent-skills.lock
skcr clean --plan
skcr clean --write
```
