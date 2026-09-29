// LogErrorsList — aggregated sssd log pattern matches with sample lines.
import type { SSSDLogError } from '../../api/backend';

export interface LogErrorsListProps {
  errors: SSSDLogError[];
}

export function LogErrorsList({ errors }: LogErrorsListProps) {
  return (
    <>
      {errors.map((error, index) => (
        <div className="error-item" key={`${index}-${error.description}`}>
          <h3 className="error-title">{error.description}</h3>
          {error.examples && error.examples.length > 0 && (
            <div className="log-block">
              {error.examples.map((example, exampleIndex) => (
                <span key={`${index}-${exampleIndex}`}>
                  {example}
                  {exampleIndex < error.examples.length - 1 ? <br /> : null}
                </span>
              ))}
            </div>
          )}
        </div>
      ))}
    </>
  );
}

export default LogErrorsList;
