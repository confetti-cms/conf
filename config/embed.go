package config

import (
	"embed"
    "src/resources/templates"
)

// Embed contains fields of all resources to be loaded during compile time.
var Embed = struct {
    Template embed.FS
}{
    Template: templates.Templates,
}
