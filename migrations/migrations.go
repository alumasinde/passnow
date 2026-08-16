package migrations

import "embed"

// FS contains the versioned SQL migrations shipped with PassNow.
// Migration files live beside this package so they can be embedded into the
// application binary and cannot be accidentally omitted from a deployment.
//
// Only forward migrations are currently embedded because the existing PassNow
// migration history contains .up.sql files only. Rollbacks will be introduced
// deliberately after reverse migrations have been reviewed and added.
//
//go:embed *.up.sql
var FS embed.FS
