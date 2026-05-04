# Agent Context

**This repo:** `ffreis-lambdas-packer` — CLI that zips Lambda artifacts and syncs them
to S3 with true-sync semantics (upload new/changed, delete stale `.zip` keys under
the prefix).

## Non-obvious facts

- **True sync — deletes stale keys.** The sync deletes `.zip` files under the prefix
  that no longer have a corresponding local artifact. Use `--no-delete` if you need
  additive-only behavior (rare).

- **Artifacts are zipped on-the-fly** — input is compiled binaries (e.g.,
  `*/bootstrap` from cargo lambda); this tool streams them to S3 as `.zip`.

- **Idempotent.** Safe to re-run. Unchanged artifacts are skipped (content hash check).

- **Used by `ffreis-website-lambdas-rust`** via `make upload` (either as Go binary or
  via Docker using `CONTAINER_RUNTIME`).

- **Used by `platform-shared-infra`** to manage Lambda artifact packages for the
  shared form-handling Lambdas.

## Build/run

```bash
go run ./cmd/lambdas-packer \
  --bucket my-bucket \
  --prefix lambdas/prod/ \
  --artifact-dir ./lambdas/target/lambda \
  --dry-run
```

## Keeping this file current

- **If you discover a fact not reflected here:** add it before finishing your task.
- **If something here is wrong or outdated:** correct it in the same commit as the code change.
- **If you rename a file, command, or concept referenced here:** update the reference.
