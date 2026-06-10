// types.go - Type aliases for backward compatibility
// All types have been moved to pkg/types/types.go
// These aliases allow existing code in package main to compile
// without modifications during the incremental refactoring.
package main

import "sssd-inspector/pkg/types"

type SSSDLogError = types.SSSDLogError
type TIDArticle = types.TIDArticle
type ReportData = types.ReportData
type TimelineEvent = types.TimelineEvent
