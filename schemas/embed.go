// Package schemas embeds the official GitHub Actions schema for internal use.
package schemas

import _ "embed"

// RelPath is the repository path to the action schema.
const RelPath = "schemas/action.schema.json"

// ActionSchema holds the embedded GitHub Actions schema JSON.
//
//go:embed action.schema.json
var ActionSchema []byte
