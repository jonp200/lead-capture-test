// Package web embeds the compiled Vue frontend.
package web

import "embed"

// Assets contains the frontend built by npm run build.
//
//go:embed dist
var Assets embed.FS
