package templates

import "embed"

// Templates provides a collection of all views and templates. This is to load all
// predefined templates.
//go:embed *
var Templates embed.FS
