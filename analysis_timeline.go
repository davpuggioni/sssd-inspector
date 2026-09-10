// analysis_timeline.go
//
// Bounded timeline aggregation (P7).
//
// The pattern scanners can fire on every line of a huge retry-loop log, so
// appending one TimelineEvent per match would let the timeline (and the JSON
// export) grow unbounded. The timelineAggregator below collapses repeats of
// the same diagnostic event into a single row that keeps the FIRST timestamp
// and sample plus a running count (Occurrences). Burst consumers (temporal
// clusters, sequence chains, fuzzy KB queries) still see the volume via the
// count without paying per-occurrence memory.
package main

import "sort"

// Timeline aggregation bounds.
const (
	// maxTimelineEvents caps the number of distinct timeline rows kept per
	// scan. Beyond it, further distinct events are dropped but counted so
	// the report can say so.
	maxTimelineEvents = 5000
	// maxTimelineSamples caps how many raw-log samples a single aggregated
	// row keeps for drill-down. The total occurrence count is unbounded.
	maxTimelineSamples = 3
)

// timelineAggregator incrementally aggregates scan matches into bounded
// timeline rows, keyed by (timestamp, message).
type timelineAggregator struct {
	events  []TimelineEvent
	index   map[string]int // aggregation key -> position in events
	dropped int            // distinct events dropped past maxTimelineEvents
}

// timelineAggKey groups repeats of the same diagnostic event at the same
// timestamp. Raw logs differ (PIDs, counters) while the timestamp+message
// identify the diagnostic signal.
func timelineAggKey(ts, msg string) string {
	return ts + "\x00" + msg
}

// newTimelineAggregator pre-allocates for the common case (< 1000 rows).
func newTimelineAggregator() *timelineAggregator {
	return &timelineAggregator{
		events: make([]TimelineEvent, 0, 1000),
		index:  make(map[string]int),
	}
}

// add records one match. Identical (timestamp, message) repeats bump the
// Occurrences counter of the existing row (keeping the first sample) and at
// most maxTimelineSamples distinct raw-log samples are retained.
func (a *timelineAggregator) add(ts, msg, raw string) {
	key := timelineAggKey(ts, msg)
	if pos, ok := a.index[key]; ok {
		ev := &a.events[pos]
		ev.Occurrences++
		if len(ev.Samples) < maxTimelineSamples && raw != ev.RawLog {
			dup := false
			for _, s := range ev.Samples {
				if s == raw {
					dup = true
					break
				}
			}
			if !dup {
				ev.Samples = append(ev.Samples, raw)
			}
		}
		return
	}
	if len(a.events) >= maxTimelineEvents {
		a.dropped++
		return
	}
	a.index[key] = len(a.events)
	a.events = append(a.events, TimelineEvent{
		Timestamp:   ts,
		Message:     msg,
		RawLog:      raw,
		Occurrences: 1,
	})
}

// sortTimelineChronological orders rows oldest-first, unparseable timestamps
// last (same semantics as the previous inline sorts).
func sortTimelineChronological(events []TimelineEvent) {
	sort.SliceStable(events, func(i, j int) bool {
		ti := normalizeTimestamp(events[i].Timestamp)
		tj := normalizeTimestamp(events[j].Timestamp)
		if ti == "" && tj == "" {
			return false
		}
		if ti == "" {
			return false
		}
		if tj == "" {
			return true
		}
		return ti < tj
	})
}

// timelineTotalOccurrences returns the total observed matches behind the
// (possibly aggregated) rows. Rows built before aggregation carry
// Occurrences == 0 and count as 1 for backward compatibility.
func timelineTotalOccurrences(events []TimelineEvent) int {
	total := 0
	for _, ev := range events {
		n := ev.Occurrences
		if n < 1 {
			n = 1
		}
		total += n
	}
	return total
}

// timelinePreviewRows returns the first N timeline rows for the HTML
// report timeline table (P6). The full volume stays available via
// timelineTotalOccurrences; the table itself stays printable.
func timelinePreviewRows(events []TimelineEvent) []TimelineEvent {
	const maxTimelinePreviewRows = 50
	if len(events) <= maxTimelinePreviewRows {
		return events
	}
	return events[:maxTimelinePreviewRows]
}

// timelineShownCount reports how many rows the preview table shows.
func timelineShownCount(events []TimelineEvent) int {
	return len(timelinePreviewRows(events))
}

// timelineRowOccurrences reports the occurrence count of one row,
// defaulting to 1 for pre-aggregation rows (Occurrences == 0).
func timelineRowOccurrences(ev TimelineEvent) int {
	if ev.Occurrences < 1 {
		return 1
	}
	return ev.Occurrences
}
