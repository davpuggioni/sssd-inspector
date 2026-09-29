// KbArticlesList — knowledge base articles whose patterns matched the report.
//
// NOTE: TIDArticle.Evidence is tagged `json:"-"` in types.go (never shipped to
// the frontend), so the legacy "Evidence Found" block could never render. The
// React port does not fake it; surfacing that evidence needs a backend change.
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
        </div>
      ))}
    </>
  );
}

export default KbArticlesList;
