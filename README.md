# 🧱 act3-ai/dagger

## A monorepo of reusable Dagger modules

This repository contains a growing collection of Dagger modules for use in common CI/CD, release, security, and automation workflows. Each module is designed to be **composable**, and **versioned**, making it easy to reuse across many repositories without duplication.

---

## 📦 Modules

Each module can be used independently and is versioned separately.

|Module|Version|Description|
|------|-------|-----------|
|**datatool**|![v](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Fact3-ai%2Fdagger%2Fmain%2Fdata-tool%2FVERSION&search=^(\d%2B\.\d%2B\.\d%2B)\s*%24&label=)|General tooling for data workflows|
|**docker**|![v](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Fact3-ai%2Fdagger%2Fmain%2Fdocker%2FVERSION&search=^(\d%2B\.\d%2B\.\d%2B)\s*%24&label=)|Builds, tags, and publishes container images|
|**git-cliff**|![v](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Fact3-ai%2Fdagger%2Fmain%2Fgit-cliff%2FVERSION&search=^(\d%2B\.\d%2B\.\d%2B)\s*%24&label=)|Changelog and version generation using Git-Cliff|
|**gocoverage**|![v](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Fact3-ai%2Fdagger%2Fmain%2Fgocoverage%2FVERSION&search=^(\d%2B\.\d%2B\.\d%2B)\s*%24&label=)|Go coverage tooling and reporting|
|**goreleaser**|![v](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Fact3-ai%2Fdagger%2Fmain%2Fgoreleaser%2FVERSION&search=^(\d%2B\.\d%2B\.\d%2B)\s*%24&label=)|Helpers for GoReleaser automation|
|**govulncheck**|![v](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Fact3-ai%2Fdagger%2Fmain%2Fgovulncheck%2FVERSION&search=^(\d%2B\.\d%2B\.\d%2B)\s*%24&label=)|Scan Go dependencies for vulnerabilities|
|**markdownlint**|![v](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Fact3-ai%2Fdagger%2Fmain%2Fmarkdownlint%2FVERSION&search=^(\d%2B\.\d%2B\.\d%2B)\s*%24&label=)|Lint Markdown files|
|**python**|![v](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Fact3-ai%2Fdagger%2Fmain%2Fpython%2FVERSION&search=^(\d%2B\.\d%2B\.\d%2B)\s*%24&label=)|Python tooling workflows|
|**release**|![v](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Fact3-ai%2Fdagger%2Fmain%2Frelease%2FVERSION&search=^(\d%2B\.\d%2B\.\d%2B)\s*%24&label=)|Release automation and versioning|
|**renovate**|![v](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Fact3-ai%2Fdagger%2Fmain%2Frenovate%2FVERSION&search=^(\d%2B\.\d%2B\.\d%2B)\s*%24&label=)|Run Renovate in Dagger|
|**yamllint**|![v](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Fact3-ai%2Fdagger%2Fmain%2Fyamllint%2FVERSION&search=^(\d%2B\.\d%2B\.\d%2B)\s*%24&label=)|Lint YAML files|

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
