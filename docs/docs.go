// Package docs embeds the specification of Breadcrumb SDD, so that
// bcr carries the one it was built with (TD-0029).
package docs

import _ "embed"

// Spec is docs/breadcrumb-sdd.md, the specification (RQ-0035).
//
//go:embed breadcrumb-sdd.md
var Spec string
