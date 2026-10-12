# Coding standards

Mechanical rules are enforced by `scripts/check.sh` (`.golangci.yml`, `.swiftlint.yml`). This file holds only what a linter cannot check.

- A change to paths, commands, build steps or requirements updates `README.md` and `docs/instruction.md` in the same diff.
- Docs state only what was verified. Remove claims you did not check.
- Comments in Go are doc comments on declarations. Put the explanation in the doc comment, not inline.
- Every new file is referenced from code, docs or the build. Delete files nothing uses.
