// KbSuggestionsList — fuzzy (TF-IDF) matches for log lines no rule matched.
import type { KBSuggestion } from '../../api/backend';

export interface KbSuggestionsListProps {
  suggestions: KBSuggestion[];
}

export function KbSuggestionsList({ suggestions }: KbSuggestionsListProps) {
  return (
    <>
      <p className="kb-description">
        Log lines that no known pattern matched, correlated with similar Knowledge Base articles
        (TF-IDF similarity):
      </p>
      {suggestions.map((suggestion, index) => (
        <div className="kb-article" key={`${suggestion.tid_id}-${index}`}>
          <h4>
            <a className="tid-link" href={suggestion.url} target="_blank" rel="noreferrer">
              {suggestion.tid_id}: {suggestion.title}
            </a>
          </h4>
          <p className="tid-id">Similarity: {Math.round(suggestion.score * 100)}%</p>
          {suggestion.sample_line ? <div className="log-block">{suggestion.sample_line}</div> : null}
        </div>
      ))}
    </>
  );
}

export default KbSuggestionsList;
