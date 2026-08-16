# PassNow database migrations

PassNow uses versioned SQL migrations for the shared MySQL schema.

Migration files are forward-only in the current repository history (`*.up.sql`).
The migration runner embeds these files into the Go binary so deployments do not
need a separate migrations directory.

## Commands

```bash
go run ./cmd/passnow migrate
go run ./cmd/passnow migrate status
```

The normal application connection keeps MySQL `multiStatements` disabled. The
migration connection enables it only for applying the versioned migration files.

Rollback commands are intentionally not exposed yet because the existing
migration history has no reviewed `.down.sql` files. Add and review reverse
migrations before introducing rollback support.
