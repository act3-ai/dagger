---
name: New Module Checklist
about: Checklist for adding a new module
title: "[MODULE] Add module <name>"
labels: enhancement
assignees: ''

---

- [ ] Create top-level directory for module using module name
- [ ] Initialize module `dagger init --sdk=go` (in module dir)
- [ ] Implement module
- [ ] Add `tests` subdirectory to module dir, refer to [dagger testing docs](https://docs.dagger.io/reference/best-practices/modules/#module-tests)
- [ ] Copy and existing cliff.toml to module directory, update "include paths" regex at bottom of file to reflect module name, refer to [example](https://github.com/act3-ai/dagger/blob/main/release/cliff.toml#L118)
- [ ] Add module to README.md
