// Package plugins imports all built-in plugins for registration
package plugins

import (
	// Import plugins for side-effect registration
	_ "github.com/es6kr/pmux/internal/plugins/theia"
	_ "github.com/es6kr/pmux/internal/plugins/ttyd"
)
