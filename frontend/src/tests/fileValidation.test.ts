// fileValidation.test.ts — the archive check that decides whether the GUI
// will even start an analysis.
//
// The cases mirror the legacy validators.test.js suite, plus the regression
// that suite could not have caught: ".tar.xz" must be accepted, because the Go
// file dialog offers "*.txz;*.tar.xz" and the analysis handles it.
import { describe, expect, it } from 'vitest';
import { validateArchivePath } from '../utils/fileValidation';

describe('validateArchivePath', () => {
  it('accepts the supportconfig archives the backend can open', () => {
    expect(validateArchivePath('/tmp/supportconfig.txz').isValid).toBe(true);
    expect(validateArchivePath('/home/u/Downloads/supportconfig.tar.xz').isValid).toBe(true);
  });

  it('is case-insensitive', () => {
    expect(validateArchivePath('/tmp/SUPPORTCONFIG.TXZ').isValid).toBe(true);
    expect(validateArchivePath('/tmp/SupportConfig.Tar.Xz').isValid).toBe(true);
  });

  it('accepts an already-extracted supportconfig directory', () => {
    expect(validateArchivePath('/tmp/supportconfig').isValid).toBe(true);
    // A dot in a parent directory must not be mistaken for an extension.
    expect(validateArchivePath('/home/u/v1.2/supportconfig').isValid).toBe(true);
  });

  it('rejects a file with an unsupported extension', () => {
    const result = validateArchivePath('/tmp/notes.txt');
    expect(result.isValid).toBe(false);
    expect(result.message).toContain('.txz');
    expect(result.message).toContain('.tar.xz');
  });

  it('rejects a file with no recognisable name', () => {
    expect(validateArchivePath('/tmp/sssd.conf').isValid).toBe(false);
    expect(validateArchivePath('/tmp/report.pdf').isValid).toBe(false);
  });

  it('rejects an empty path', () => {
    expect(validateArchivePath('').isValid).toBe(false);
    expect(validateArchivePath('   ').isValid).toBe(false);
    expect(validateArchivePath(undefined as unknown as string).isValid).toBe(false);
  });

  it('does not accept a partial suffix such as .xz alone', () => {
    expect(validateArchivePath('/tmp/supportconfig.xz').isValid).toBe(false);
  });
});
