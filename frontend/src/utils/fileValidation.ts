// fileValidation — typed view of the legacy validators module.
//
// validators.js is plain JavaScript with JSDoc-free signatures, so TypeScript
// infers `Object` for its return values. The wrappers below state the contract
// the UI relies on while keeping a single implementation (the same file the
// browser test-runner exercises), so the two can never disagree.
import { FileValidator } from '../utils/validators.js';

export interface FileValidationResult {
  isValid: boolean;
  message: string;
}

/** True when the path looks like a supportconfig archive the analyser accepts. */
export function validateArchivePath(path: string): FileValidationResult {
  const result = FileValidator.validateFileExtension(path) as unknown as FileValidationResult;
  return { isValid: Boolean(result?.isValid), message: String(result?.message ?? '') };
}
