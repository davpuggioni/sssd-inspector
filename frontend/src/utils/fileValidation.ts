// fileValidation — is this something the analyser can actually read?
//
// Replaces utils/validators.js (FileValidator), which the legacy frontend used
// for exactly this check. The old implementation took the substring from the
// LAST dot in the name and compared it against ['.txz', '.tar.xz'], so
// "supportconfig.tar.xz" reduced to ".xz" and was rejected — the GUI refused
// the most common supportconfig filename that its own file dialog offers
// (constants.PatternSupportconfig in Go is "*.txz;*.tar.xz"). Matching the
// longest supported suffix instead fixes that and keeps the check honest.
import { SUPPORTED_EXTENSIONS } from '../config/ui';

export interface FileValidationResult {
  isValid: boolean;
  message: string;
}
/**
 * True when the path is something the analyser can open: a supported archive,
 * or a path with no file extension — an already-extracted supportconfig
 * directory, which the analysis supports (`analyzeData` works on a directory
 * and `extractIfArchive` returns it unchanged) and which the CLI accepts too.
 * Anything else carries a different extension and is rejected with a message
 * that says what is supported.
 */
export function validateArchivePath(path: string): FileValidationResult {
  if (typeof path !== 'string' || path.trim() === '') {
    return { isValid: false, message: 'File name is required' };
  }

  const name = path.trim().toLowerCase();
  // Longest first, so ".tar.xz" is tested before ".xz" could ever match.
  const supported = [...SUPPORTED_EXTENSIONS].sort((a, b) => b.length - a.length);
  if (supported.some((extension) => name.endsWith(extension))) {
    return { isValid: true, message: 'File format is supported' };
  }

  const lastSegment = name.split('/').pop() ?? name;
  if (!lastSegment.includes('.')) {
    return { isValid: true, message: 'Assuming an already-extracted supportconfig directory' };
  }

  return {
    isValid: false,
    message: `Unsupported file format. Supported formats: ${SUPPORTED_EXTENSIONS.join(', ')} (or an already-extracted supportconfig directory)`,
  };
}
