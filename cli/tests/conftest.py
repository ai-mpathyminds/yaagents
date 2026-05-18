"""pytest configuration — points tests at the shared conformance corpus."""

from __future__ import annotations

import pathlib

import pytest

# The corpus lives at yaagents/spec/examples/v0.1/ — two dirs up from cli/.
_CORPUS = pathlib.Path(__file__).parent.parent.parent / "spec" / "examples" / "v0.1"


def pytest_configure(config: pytest.Config) -> None:
    if not _CORPUS.is_dir():
        raise RuntimeError(f"Conformance corpus not found at {_CORPUS}")


@pytest.fixture(scope="session")
def corpus_dir() -> pathlib.Path:
    return _CORPUS


@pytest.fixture(scope="session")
def valid_fixtures(corpus_dir: pathlib.Path) -> list[pathlib.Path]:
    return sorted(corpus_dir.glob("*.valid*.json"))


@pytest.fixture(scope="session")
def invalid_fixtures(corpus_dir: pathlib.Path) -> list[pathlib.Path]:
    return sorted(corpus_dir.glob("*.invalid*.json"))
