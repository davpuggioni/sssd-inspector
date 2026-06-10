// cache.go - Legacy wrapper for fileutil caching
// All actual implementations have moved to pkg/fileutil.
package main

import "sssd-inspector/pkg/fileutil"

// Global cache instances (backward compatibility)
var (
	globalRegexCache = fileutil.GlobalRegexCache
	globalFileCache  = fileutil.GlobalFileCache
)
