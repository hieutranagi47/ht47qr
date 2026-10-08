---
name: golang-ddd-new-project
description: "Scaffold a new Go DDD HTTP service or add a DDD module with ready-to-run common HTTP, logging, error, UUID, and module-lifecycle foundations. Use when creating a service from scratch, starting a bounded-context module, or standardizing a new Go service on this repository's services/auth and services/ht47agent conventions."
---

# Go DDD New Project

Create a small, compilable starting point. Keep it generic: add a database,
migrations, external clients, and real use cases only when the request needs them.

## Workflow

1. Read `references/conventions.md` before creating code.
2. Confirm the target directory is new and choose the Go module import path and
   first module name. Use lowercase snake_case for Go package names and a
   hyphen-free HTTP path segment.
3. For a service, run:

   ```sh
   scripts/scaffold.sh project <target-directory> <go-module-path> <first-module>
   ```

   The script copies the maintained `services/auth/common` foundation, creates
   a composition root, and generates the first module. Run it from this
   repository or pass `--template-root <path-to-services/auth>`.
4. To add another module, run:

   ```sh
   scripts/scaffold.sh module <project-directory> <module-name>
   ```

   Then add `<module>.NewModule()` to `cmd/main.go`'s `service.New` call.
5. Replace the example route and domain value with a business use case. Keep
   transport serialization in `api/http`, orchestration in `app`, invariants in
   `domain`, and databases/clients in `adapters`.
6. Run `go test ./...` and `go vet ./...` from the generated project. Do not
   add CQRS, a repository, or migrations merely because the scaffold has room
   for them.

## Guardrails

- Keep `common` cross-cutting only; never put module business rules there.
- Implement `module.Module` fully (`Name`, `Init`, `RegisterContracts`, and
  `RegisterHttp`) for every module.
- Add cross-module APIs to `common/module/contracts`, and validate required
  contracts in `Verify`; do not import one module directly from another.
- Keep `cmd/main.go` as composition and process lifecycle only.

## Resource

- `scripts/scaffold.sh` creates the skeleton. Read it before changing its
  output or supporting another transport.
- `references/conventions.md` specifies the generated dependency direction and
  what to customize first.
