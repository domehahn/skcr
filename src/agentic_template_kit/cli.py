from __future__ import annotations

from pathlib import Path
from typing import Optional

import typer
from rich.console import Console
from rich.table import Table

from .bake import build_initial_config, dump_bake_file, load_bake_file, resolve_target
from .lockfile import write_lockfile
from .models import parse_platforms
from .renderer import render_files
from .validator import validate_project

app = typer.Typer(help="Generate agentic DevSecOps SDLC templates for multiple agent platforms.")
console = Console()


@app.command()
def init(
    target: Path = typer.Option(Path("."), "--target", "-t", help="Target repository path."),
    platform: Optional[str] = typer.Option(
        None,
        "--platform",
        help='Comma-separated platforms, e.g. "gitlab-duo,codex,github-copilot".',
    ),
    preset: Optional[str] = typer.Option(None, "--preset", help="Preset: minimal, gitlab, enterprise, local-ai, all."),
    project_name: Optional[str] = typer.Option(None, "--project-name", help="Project name used in rendered templates."),
    owner_team: str = typer.Option("platform-engineering", "--owner-team", help="Owning team."),
    language: str = typer.Option("de", "--language", help="Default documentation language."),
    governance_level: str = typer.Option("standard", "--governance-level", help="relaxed, standard, strict."),
    force: bool = typer.Option(False, "--force", help="Overwrite existing agentic.bake.yaml."),
) -> None:
    target = target.resolve()
    target.mkdir(parents=True, exist_ok=True)
    bake_path = target / "agentic.bake.yaml"

    if bake_path.exists() and not force:
        raise typer.BadParameter(f"{bake_path} already exists. Use --force to overwrite.")

    platforms = parse_platforms(platform)
    if project_name is None:
        project_name = target.name

    config = build_initial_config(
        platforms=platforms,
        project_name=project_name,
        owner_team=owner_team,
        language=language,
        governance_level=governance_level,
        preset=preset,
    )
    dump_bake_file(config, bake_path)
    console.print(f"[green]Created[/green] {bake_path}")


@app.command("list-targets")
def list_targets(target: Path = typer.Option(Path("."), "--target", "-t", help="Target repository path.")) -> None:
    config = load_bake_file(target / "agentic.bake.yaml")
    table = Table(title="Agentic Bake Targets")
    table.add_column("Target")
    table.add_column("Description")
    table.add_column("Platforms / Inherits")

    for name, cfg in config.targets.items():
        value = ", ".join(cfg.platforms or cfg.inherits)
        table.add_row(name, cfg.description, value)

    console.print(table)


@app.command()
def bake(
    name: str = typer.Argument("default", help="Target name from agentic.bake.yaml."),
    target: Path = typer.Option(Path("."), "--target", "-t", help="Target repository path."),
    dry_run: bool = typer.Option(False, "--dry-run", help="Show generated files without writing."),
    write: bool = typer.Option(False, "--write", help="Write files to target repository."),
) -> None:
    if not dry_run and not write:
        dry_run = True

    target = target.resolve()
    config = load_bake_file(target / "agentic.bake.yaml")
    resolved = resolve_target(config, name)
    files = render_files(config, resolved)

    table = Table(title=f"Bake target: {name}")
    table.add_column("Action")
    table.add_column("Platform")
    table.add_column("Path")

    for rendered in files:
        path = target / rendered.destination
        action = "create"
        if path.exists():
            existing = path.read_text(encoding="utf-8")
            action = "unchanged" if existing == rendered.content else "update"
        table.add_row(action, rendered.platform, str(rendered.destination))

    console.print(table)

    if dry_run:
        console.print("[yellow]Dry run only. Use --write to write files.[/yellow]")
        return

    for rendered in files:
        path = target / rendered.destination
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(rendered.content, encoding="utf-8")

    write_lockfile(target, files, name)
    console.print(f"[green]Wrote {len(files)} files and .agentic-template.lock[/green]")


@app.command()
def validate(target: Path = typer.Option(Path("."), "--target", "-t", help="Target repository path.")) -> None:
    errors = validate_project(target.resolve())
    if errors:
        for error in errors:
            console.print(f"[red]ERROR[/red] {error}")
        raise typer.Exit(1)

    console.print("[green]Validation passed[/green]")


if __name__ == "__main__":
    app()
