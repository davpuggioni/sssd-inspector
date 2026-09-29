// KbArticlesList — knowledge base articles whose patterns matched the report.
//
// Evidence is the log line(s) that triggered the match (TIDArticle.Evidence,
// capped at three by the matcher). It is the difference between "this TID might
// apply" and "this TID applied because of this line".
import type { TIDArticle } from '../../api/backend';

export interface KbArticlesListProps {
  articles: TIDArticle[];
}

export function KbArticlesList({ articles }: KbArticlesListProps) {
  return (
    <>
      {articles.map((article, index) => (
        <div className="kb-article" key={`${article.tid_id}-${index}`}>
          <h4>
            <a className="tid-link" href={article.url} target="_blank" rel="noreferrer">
              {article.title}
            </a>
          </h4>
          <p className="tid-id">TID: {article.tid_id}</p>
          <p className="kb-description">{article.description}</p>
          {article.evidence && article.evidence.length > 0 ? (
            <details>
              <summary>Evidence Found</summary>
              <div className="kb-evidence">
                {article.evidence.map((line, lineIndex) => (
                  <div key={`${article.tid_id}-${lineIndex}`}>{line}</div>
                ))}
              </div>
            </details>
          ) : null}
        </div>
      ))}
    </>
  );
}

export default KbArticlesList;
