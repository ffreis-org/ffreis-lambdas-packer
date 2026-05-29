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

- **Used by the website Lambdas repo** via `make upload` (either as Go binary or
  via Docker using `CONTAINER_RUNTIME`).

- **Used by a private platform repo** to manage Lambda artifact packages for the
  shared form-handling Lambdas.

## Build/run

**Sync mode** (Rust/Go multi-Lambda artifact directory):
```bash
go run ./cmd/lambdas-packer \
  --bucket my-bucket \
  --prefix lambdas/prod/ \
  --artifact-dir ./lambdas/target/lambda \
  --dry-run
```

**Single-file mode** (Python/pre-built zip, e.g. `monitor_evaluator.zip`):
```bash
go run ./cmd/lambdas-packer \
  --bucket my-bucket \
  --file dist/monitor_evaluator.zip \
  --key monitor-lambda/monitor_evaluator.zip
```
Use `--dry-run` to preview without uploading. No deletion logic runs in single-file mode.

## Keeping this file current

- **If you discover a fact not reflected here:** add it before finishing your task.
- **If something here is wrong or outdated:** correct it in the same commit as the code change.
- **If you rename a file, command, or concept referenced here:** update the reference.
