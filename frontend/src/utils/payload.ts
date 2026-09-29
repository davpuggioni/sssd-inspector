// payload.ts — reading a backend payload safely.
//
// It lives here, not in api/backend.ts, on purpose: api/backend.ts is the
// module tests replace wholesale with a stub, and a pure data helper must not
// disappear with the bridge it is meant to protect.

/**
 * Read an array field out of a backend payload.
 *
 * `encoding/json` renders a nil Go slice as `null`, so a payload can carry
 * `"rules": null` even though the generated model declares `rules: RuleInfo[]` —
 * which is exactly what an installation with no custom rules sends, and it
 * crashed the Definitions Studio before this helper existed. Every array read
 * from a payload goes through it, so a null field can never take a panel down:
 * the cost is one function call, the benefit is that the UI does not depend on
 * the backend happening to allocate an empty slice.
 */
export function listOf<T>(value: T[] | null | undefined): T[] {
  return Array.isArray(value) ? value : [];
}
