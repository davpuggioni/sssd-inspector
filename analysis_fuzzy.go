// analysis_fuzzy.go
// Phase 3: TF-IDF cosine-similarity matching between unmatched log lines and
// Knowledge Base articles.
//
// The literal/regex engines only detect what they know. When a log line is
// SSSD-related but matches NO known pattern, this module still connects it to
// the most similar KB articles, so the support engineer gets a hint instead
// of nothing. Cosine similarity over a TF-IDF vector space: each article is
// a document (title + description + article text), each unmatched log line
// is a query document.
package main

import (
	"math"
	"sort"
	"strings"
	"unicode"
)

// kbSuggestionMinScore is the minimum cosine similarity for a suggestion.
const kbSuggestionMinScore = 0.12

// kbSuggestionMax is the maximum number of suggested articles in the report.
const kbSuggestionMax = 3

// kbFuzzyMaxQueries caps how many unmatched lines are evaluated (cost guard
// for huge archives; the first occurrences are the most representative).
const kbFuzzyMaxQueries = 200

// kbStopwords are tokens with no discriminative power in SSSD logs.
var kbStopwords = map[string]struct{}{
	"the": {}, "and": {}, "for": {}, "not": {}, "with": {}, "from": {},
	"this": {}, "that": {}, "was": {}, "are": {}, "but": {}, "has": {},
	"have": {}, "will": {}, "its": {}, "into": {}, "out": {}, "all": {},
	"can": {}, "may": {}, "one": {}, "use": {}, "using": {}, "when": {},
	"sssd": {}, "child": {}, "line": {}, "file": {}, "because": {},
	"unable": {}, "could": {},
}

// tokenizeKB lowercases and splits text into normalized alphanumeric tokens,
// dropping stopwords and tokens shorter than 3 characters.
func tokenizeKB(text string) []string {
	var tokens []string
	cur := strings.Builder{}
	flush := func() {
		if cur.Len() >= 3 {
			t := cur.String()
			if _, stop := kbStopwords[t]; !stop {
				tokens = append(tokens, t)
			}
		}
		cur.Reset()
	}
	for _, r := range strings.ToLower(text) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cur.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return tokens
}

// kbVector builds a term-frequency vector from tokens.
func kbVector(tokens []string) map[string]float64 {
	v := make(map[string]float64, len(tokens))
	for _, t := range tokens {
		v[t]++
	}
	return v
}

// kbTfIdf builds the TF-IDF vector of a document against the corpus
// document-frequency table (df). Terms absent from the corpus are ignored
// for articles, but kept for queries with a high default IDF so rare log
// tokens (option names, error codes) dominate the similarity.
func kbTfIdf(tokens []string, df map[string]int, corpusSize int, isQuery bool) map[string]float64 {
	tf := kbVector(tokens)
	v := make(map[string]float64, len(tf))
	for term, f := range tf {
		var idf float64
		if n, ok := df[term]; ok && n > 0 {
			idf = 1 + float64(corpusSize)/float64(n)
		} else if isQuery {
			idf = 1 + float64(corpusSize)
		} else {
			continue
		}
		v[term] = (1 + f) * idf
	}
	return v
}

// cosineSimilarity computes the cosine of two sparse vectors:
// dot(a,b) / (||a|| * ||b||).
func cosineSimilarity(a, b map[string]float64) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	var dot, na, nb float64
	for t, va := range a {
		na += va * va
		if vb, ok := b[t]; ok {
			dot += va * vb
		}
	}
	for _, vb := range b {
		nb += vb * vb
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

// kbDocumentTokens returns the token list of an article's document.
func kbDocumentTokens(a *TIDArticle) []string {
	text := a.Title + "\n" + a.Description + "\n" + a.Situation + "\n" + a.Cause + "\n" + a.PlainText
	return tokenizeKB(text)
}

// buildKBCorpus tokenizes all articles and returns (docs, df).
func buildKBCorpus(articles []TIDArticle) ([][]string, map[string]int) {
	docs := make([][]string, 0, len(articles))
	df := make(map[string]int)
	for i := range articles {
		toks := kbDocumentTokens(&articles[i])
		docs = append(docs, toks)
		s := make(map[string]struct{}, len(toks))
		for _, t := range toks {
			if _, ok := s[t]; !ok {
				s[t] = struct{}{}
				df[t]++
			}
		}
	}
	return docs, df
}

// analyzeKBSuggestions links unmatched, SSSD-related log lines to the most
// similar KB articles via TF-IDF cosine similarity. Articles already matched
// by the deterministic engines are skipped: suggestions must ADD information.
func analyzeKBSuggestions(timeline []TimelineEvent, kbArticles []TIDArticle, matchedTIDs []TIDArticle) []KBSuggestion {
	if len(kbArticles) == 0 || len(timeline) == 0 {
		return nil
	}

	matched := make(map[string]struct{}, len(matchedTIDs))
	for _, t := range matchedTIDs {
		matched[t.TIDID] = struct{}{}
	}

	docs, df := buildKBCorpus(kbArticles)
	corpusSize := len(docs)
	docVectors := make([]map[string]float64, corpusSize)
	for i, toks := range docs {
		docVectors[i] = kbTfIdf(toks, df, corpusSize, false)
	}

	// Collect unique raw log lines (bounded).
	seen := make(map[string]struct{})
	var queries []string
	for _, ev := range timeline {
		line := strings.TrimSpace(ev.RawLog)
		if line == "" {
			continue
		}
		if _, dup := seen[line]; dup {
			continue
		}
		seen[line] = struct{}{}
		queries = append(queries, line)
		if len(queries) >= kbFuzzyMaxQueries {
			break
		}
	}

	type agg struct {
		score float64
		line  string
	}
	best := make(map[string]*agg)

	for _, q := range queries {
		qv := kbTfIdf(tokenizeKB(q), df, corpusSize, true)
		if len(qv) == 0 {
			continue
		}
		for i := range kbArticles {
			if _, skip := matched[kbArticles[i].TIDID]; skip {
				continue
			}
			score := cosineSimilarity(qv, docVectors[i])
			if score < kbSuggestionMinScore {
				continue
			}
			cur := best[kbArticles[i].TIDID]
			if cur == nil || score > cur.score {
				best[kbArticles[i].TIDID] = &agg{score: score, line: q}
			}
		}
	}

	suggestions := make([]KBSuggestion, 0, len(best))
	for tid, a := range best {
		for i := range kbArticles {
			if kbArticles[i].TIDID == tid {
				suggestions = append(suggestions, KBSuggestion{
					TIDID:      tid,
					Title:      kbArticles[i].Title,
					URL:        kbArticles[i].URL,
					Score:      a.score,
					SampleLine: a.line,
				})
				break
			}
		}
	}
	sort.SliceStable(suggestions, func(i, j int) bool {
		if suggestions[i].Score != suggestions[j].Score {
			return suggestions[i].Score > suggestions[j].Score
		}
		return suggestions[i].TIDID < suggestions[j].TIDID
	})
	if len(suggestions) > kbSuggestionMax {
		suggestions = suggestions[:kbSuggestionMax]
	}
	return suggestions
}
