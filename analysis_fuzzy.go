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

// kbSynonyms maps domain-specific token variants to a single canonical token so
// that the lexical (TF-IDF/BM25) matching treats "kdc"/"krb5", "nsupdate"/"ddns",
// "skew"/"clock" etc. as one concept. Without this, a KB article that only says
// "KDC unreachable" gets a low similarity against a log line that says
// "krb5_child timeout", even though they describe the same root cause. This is
// a cheap, model-free recall booster driven purely by the SSSD/AD lexicon
// (mirrors the cn= the logPatternCategory classification buckets).
var kbSynonyms = map[string]string{
	"krb5":         "kerberos",
	"krb":          "kerberos",
	"kdc":          "kerberos",
	"keytab":       "kerberos",
	"tgt":          "ticket",
	"ntlm":         "kerberos",
	"preauth":      "preauth",
	"nsupdate":     "dns",
	"ddns":         "dns",
	"skew":         "clock",
	"timeout":      "timeout",
	"connect":      "connect",
	"reconnect":    "connect",
	"srv":          "srv",
	"offline":      "offline",
	"disconnected": "disconnected",
	"starttls":     "tls",
	"ssl":          "tls",
	"ldaps":        "tls",
}

// normalizeKBTerm maps a token to its canonical form via kbSynonyms; tokens
// without a registered synonym are returned unchanged.
func normalizeKBTerm(term string) string {
	if syn, ok := kbSynonyms[term]; ok {
		return syn
	}
	return term
}

// tokenizeKB lowercases and splits text into normalized alphanumeric tokens,
// dropping stopwords and tokens shorter than 3 characters. Each kept token is
// passed through the domain synonym dictionary so equivalent vocabulary from
// the SSSD world maps to a single term before TF-IDF/BM25 scoring.
func tokenizeKB(text string) []string {
	var tokens []string
	cur := strings.Builder{}
	flush := func() {
		if cur.Len() >= 3 {
			t := cur.String()
			if _, stop := kbStopwords[t]; !stop {
				tokens = append(tokens, normalizeKBTerm(t))
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

// BM25 ranking parameters: k1 controls term-frequency saturation (>0, usually
// 1.2) and b the document-length normalisation (0=off, 1=full). Values chosen
// for short SSSD log lines against longer KB articles.
const (
	bm25K1 = 1.2
	bm25B  = 0.75
)

// kbBM25 scores a document against a query using the Okapi BM25 ranking
// function. query and doc are raw term-frequency vectors (see kbVector). It
// improves on raw TF-IDF cosine for ranking: it applies document-length
// normalisation and saturates raw term frequency, so a verbose KB article is
// not unfairly downgraded because it is long.
func kbBM25(query, doc map[string]float64, docLen, avgDocLen, corpusN int, df map[string]int) float64 {
	if len(query) == 0 || len(doc) == 0 {
		return 0
	}
	if avgDocLen <= 0 {
		avgDocLen = 1
	}
	var score float64
	for term := range query {
		n, inCorpus := df[term]
		if !inCorpus {
			continue
		}
		idf := math.Log(1 + (float64(corpusN)-float64(n)+0.5)/(float64(n)+0.5))
		if idf <= 0 {
			continue
		}
		f, ok := doc[term]
		if !ok {
			continue
		}
		denom := f + bm25K1*(1-bm25B+bm25B*float64(docLen)/float64(avgDocLen))
		if denom == 0 {
			continue
		}
		score += idf * f * (bm25K1 + 1) / denom
	}
	return score
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
	docLen := make([]int, corpusSize)
	var totalLen int
	rawVectors := make([]map[string]float64, corpusSize)
	for i, toks := range docs {
		docVectors[i] = kbTfIdf(toks, df, corpusSize, false)
		rawVectors[i] = kbVector(toks)
		docLen[i] = len(toks)
		totalLen += len(toks)
	}
	avgDocLen := 0
	if corpusSize > 0 {
		avgDocLen = totalLen / corpusSize
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
		score float64 // TF-IDF cosine similarity (kept for 0..1 report compat)
		bm25  float64 // Okapi BM25 score (superior lexical ranking signal)
		line  string
	}
	best := make(map[string]*agg)

	for _, q := range queries {
		qt := tokenizeKB(q)
		qv := kbTfIdf(qt, df, corpusSize, true)
		if len(qv) == 0 {
			continue
		}
		queryVec := kbVector(qt)
		for i := range kbArticles {
			if _, skip := matched[kbArticles[i].TIDID]; skip {
				continue
			}
			score := cosineSimilarity(qv, docVectors[i])
			if score < kbSuggestionMinScore {
				continue
			}
			bm := kbBM25(queryVec, rawVectors[i], docLen[i], avgDocLen, corpusSize, df)
			cur := best[kbArticles[i].TIDID]
			if cur == nil || (bm > cur.bm25 || (bm == cur.bm25 && score > cur.score)) {
				best[kbArticles[i].TIDID] = &agg{score: score, bm25: bm, line: q}
			}
		}
	}

	// Build suggestion candidates, carrying both similarity scores.
	type candidate struct {
		tid, title, url, line string
		score, bm25           float64
	}
	cands := make([]candidate, 0, len(best))
	for tid, a := range best {
		for i := range kbArticles {
			if kbArticles[i].TIDID == tid {
				cands = append(cands, candidate{
					tid: tid, title: kbArticles[i].Title, url: kbArticles[i].URL,
					line: a.line, score: a.score, bm25: a.bm25,
				})
				break
			}
		}
	}
	// Rank primarily by BM25 (document-length-aware, frequency-saturating),
	// breaking ties by the comparable cosine similarity.
	sort.SliceStable(cands, func(i, j int) bool {
		if cands[i].bm25 != cands[j].bm25 {
			return cands[i].bm25 > cands[j].bm25
		}
		if cands[i].score != cands[j].score {
			return cands[i].score > cands[j].score
		}
		return cands[i].tid < cands[j].tid
	})
	if len(cands) > kbSuggestionMax {
		cands = cands[:kbSuggestionMax]
	}
	suggestions := make([]KBSuggestion, 0, len(cands))
	for _, c := range cands {
		suggestions = append(suggestions, KBSuggestion{
			TIDID:      c.tid,
			Title:      c.title,
			URL:        c.url,
			Score:      c.score, // cosine similarity in 0..1 for report display
			SampleLine: c.line,
		})
	}
	return suggestions
}
