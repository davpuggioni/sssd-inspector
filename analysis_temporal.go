// analysis_temporal.go
// Phase 3: sliding-window temporal correlation.
//
// Log bursts of the SAME diagnostic event inside a bounded window are the
// fingerprint of retry loops, LDAP/KDC timeouts, offline flapping and
// fail-over storms. A single occurrence is usually benign noise; N
// occurrences within `clusterGapSeconds` of each other are reported as a
// TemporalCluster.
//
// The timeline arrives already sorted chronologically by the single-pass
// scanner (normalizeTimestamp ordering), so clustering is a linear sweep:
// O(n) over the timeline.
package main

import (
	"sort"
	"time"
)

// clusterGapSeconds is the maximum gap between two occurrences of the same
// event for them to belong to the same cluster (retry loops typically fire
// every few seconds; bursts rarely outlast a couple of minutes).
const clusterGapSeconds = 300

// minClusterSize is the smallest number of occurrences worth reporting.
const minClusterSize = 3

// analyzeTemporalClusters groups same-description timeline events into
// bursts separated by more than clusterGapSeconds and returns the clusters
// with at least minClusterSize occurrences, sorted by count (descending).
func analyzeTemporalClusters(timeline []TimelineEvent) []TemporalCluster {
	type clusterKey struct {
		desc   string
		start  int64 // epoch of window start
		end    int64
		count  int
		first  string
		last   string
		sample string
	}

	// Work on a chronologically ordered copy (timeline is already sorted,
	// but do not rely on the caller).
	sorted := make([]TimelineEvent, len(timeline))
	copy(sorted, timeline)
	sort.SliceStable(sorted, func(i, j int) bool {
		ti := normalizeTimestamp(sorted[i].Timestamp)
		tj := normalizeTimestamp(sorted[j].Timestamp)
		if ti == "" || tj == "" {
			return ti != ""
		}
		return ti < tj
	})

	var clusters []clusterKey
	epoch := func(ev TimelineEvent) (int64, bool) {
		ts := normalizeTimestamp(ev.Timestamp)
		if ts == "" {
			return 0, false
		}
		t, err := time.ParseInLocation("2006-01-02 15:04:05", ts, time.Local)
		if err != nil {
			return 0, false
		}
		return t.Unix(), true
	}

	for _, ev := range sorted {
		secs, ok := epoch(ev)
		if !ok {
			continue
		}
		// Aggregated rows carry the repeat volume in Occurrences (P7): a
		// burst collapsed into one row must still count as N occurrences.
		n := ev.Occurrences
		if n < 1 {
			n = 1
		}
		appended := false
		// Extend the last cluster of this description if within the gap.
		for i := len(clusters) - 1; i >= 0; i-- {
			c := &clusters[i]
			if c.desc != ev.Message {
				continue
			}
			if secs-c.end <= clusterGapSeconds {
				c.end = secs
				c.count += n
				c.last = normalizeTimestamp(ev.Timestamp)
				appended = true
			}
			break
		}
		if !appended {
			clusters = append(clusters, clusterKey{
				desc:   ev.Message,
				start:  secs,
				end:    secs,
				count:  n,
				first:  normalizeTimestamp(ev.Timestamp),
				last:   normalizeTimestamp(ev.Timestamp),
				sample: ev.RawLog,
			})
		}
	}

	var out []TemporalCluster
	for _, c := range clusters {
		if c.count >= minClusterSize {
			out = append(out, TemporalCluster{
				Description:  c.desc,
				EventCount:   c.count,
				WindowStart:  c.first,
				WindowEnd:    c.last,
				SampleRawLog: c.sample,
			})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].EventCount > out[j].EventCount })
	return out
}
