package static

import "embed"

// DistFS embeds the frontend build output.
// The dist directory is created by running `bun run build` in the frontend directory.
//
//go:embed dist/*
var DistFS embed.FS
