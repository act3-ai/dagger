# 🧱 act3-ai/dagger

## A monorepo of reusable Dagger modules

This repository contains a growing collection of Dagger modules for use in common CI/CD, release, security, and automation workflows. Each module is designed to be **composable**, and **versioned**, making it easy to reuse across many repositories without duplication.

---

## 📦 Modules

Each module can be used independently and is versioned separately.

|Module|Version|Description|Documentation|
|------|-------|-----------|-----------|
|**datatool**|![v](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Fact3-ai%2Fdagger%2Fmain%2Fdata-tool%2FVERSION&search=^(\d%2B\.\d%2B\.\d%2B)\s*%24&label=)|General tooling for data workflows|[daggerverse](https://daggerverse.dev/mod/github.com/act3-ai/dagger/data-tool)|
|**docker**|![v](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Fact3-ai%2Fdagger%2Fmain%2Fdocker%2FVERSION&search=^(\d%2B\.\d%2B\.\d%2B)\s*%24&label=)|Builds, tags, and publishes container images|[daggerverse](https://daggerverse.dev/mod/github.com/act3-ai/dagger/docker)|
|**git-cliff**|![v](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Fact3-ai%2Fdagger%2Fmain%2Fgit-cliff%2FVERSION&search=^(\d%2B\.\d%2B\.\d%2B)\s*%24&label=)|Changelog and version generation using Git-Cliff|[daggerverse](https://daggerverse.dev/mod/github.com/act3-ai/dagger/git-cliff)|
|**gocoverage**|![v](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Fact3-ai%2Fdagger%2Fmain%2Fgocoverage%2FVERSION&search=^(\d%2B\.\d%2B\.\d%2B)\s*%24&label=)|Go coverage tooling and reporting|[daggerverse](https://daggerverse.dev/mod/github.com/act3-ai/dagger/gocoverage)|
|**goreleaser**|![v](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Fact3-ai%2Fdagger%2Fmain%2Fgoreleaser%2FVERSION&search=^(\d%2B\.\d%2B\.\d%2B)\s*%24&label=)|Helpers for GoReleaser automation|[daggerverse](https://daggerverse.dev/mod/github.com/act3-ai/dagger/goreleaser)|
|**govulncheck**|![v](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Fact3-ai%2Fdagger%2Fmain%2Fgovulncheck%2FVERSION&search=^(\d%2B\.\d%2B\.\d%2B)\s*%24&label=)|Scan Go dependencies for vulnerabilities|[daggerverse](https://daggerverse.dev/mod/github.com/act3-ai/dagger/govulncheck)|
|**markdownlint**|![v](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Fact3-ai%2Fdagger%2Fmain%2Fmarkdownlint%2FVERSION&search=^(\d%2B\.\d%2B\.\d%2B)\s*%24&label=)|Lint Markdown files|[daggerverse](https://daggerverse.dev/mod/github.com/act3-ai/dagger/markdownlint)|
|**python**|![v](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Fact3-ai%2Fdagger%2Fmain%2Fpython%2FVERSION&search=^(\d%2B\.\d%2B\.\d%2B)\s*%24&label=)|Python tooling workflows|[daggerverse](https://daggerverse.dev/mod/github.com/act3-ai/dagger/python)|
|**release**|![v](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Fact3-ai%2Fdagger%2Fmain%2Frelease%2FVERSION&search=^(\d%2B\.\d%2B\.\d%2B)\s*%24&label=)|Release automation and versioning|[daggerverse](https://daggerverse.dev/mod/github.com/act3-ai/dagger/release)|
|**renovate**|![v](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Fact3-ai%2Fdagger%2Fmain%2Frenovate%2FVERSION&search=^(\d%2B\.\d%2B\.\d%2B)\s*%24&label=)|Run Renovate in Dagger|[daggerverse](https://daggerverse.dev/mod/github.com/act3-ai/dagger/renovate)|
|**shields**|![v](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Fact3-ai%2Fdagger%2Fmain%2Fshields%2FVERSION&search=^(\d%2B\.\d%2B\.\d%2B)\s*%24&label=)|Run Shields in Dagger|[daggerverse](https://daggerverse.dev/mod/github.com/act3-ai/dagger/shields)|
|**yamllint**|![v](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Fact3-ai%2Fdagger%2Fmain%2Fyamllint%2FVERSION&search=^(\d%2B\.\d%2B\.\d%2B)\s*%24&label=)|Lint YAML files|[daggerverse](https://daggerverse.dev/mod/github.com/act3-ai/dagger/yamllint)|

---

## 🚀 Getting Started

### Install the Dagger CLI

You must have the Dagger CLI installed to use these modules:

```bash
curl -fsSL https://dagger.io/install.sh | sh
```

Verify installation:

```bash
dagger version
```

---

## 🔌 Using a Module

### Example

```bash
dagger call \
  -m github.com/act3-ai/dagger/markdownlint@v0.2.1 \
  --src=. \
  lint
```

Inspect available functions and arguments:

```bash
dagger call -m github.com/act3-ai/dagger/markdownlint@v0.2.1 --help
```

---
