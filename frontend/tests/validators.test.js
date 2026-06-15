/**
 * Test Suite for Frontend Validators
 * Tests for FileValidator, InputValidator, and FormValidator
 * @version 2.0.0
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { FileValidator, InputValidator, FormValidator } from '../src/utils/validators.js';

/**
 * Test utilities
 */
function createMockFile(name = 'test.txz', size = 1024) {
  return {
    name,
    size,
    type: 'application/x-xz',
    lastModified: Date.now()
  };
}

function createMockForm() {
  const form = document.createElement('form');
  const fileInput = document.createElement('input');
  fileInput.type = 'text';
  fileInput.id = 'filePath';
  fileInput.value = '/path/to/test.txz';
  const checkbox = document.createElement('input');
  checkbox.type = 'checkbox';
  checkbox.id = 'anonymizeCheck';
  checkbox.checked = false;
  
  form.appendChild(fileInput);
  form.appendChild(checkbox);
  return form;
}

/**
 * FileValidator Tests
 */
describe('FileValidator', () => {
  describe('validateFilePath', () => {
    it('accepts valid absolute path', () => {
      const result = FileValidator.validateFilePath('/path/to/file.txz');
      expect(result.isValid).toBe(true);
    });

    it('accepts valid Windows path', () => {
      const result = FileValidator.validateFilePath('C:\\path\\to\\file.txz');
      expect(result.isValid).toBe(true);
    });

    it('rejects empty path', () => {
      const result = FileValidator.validateFilePath('');
      expect(result.isValid).toBe(false);
    });

    it('rejects null path', () => {
      const result = FileValidator.validateFilePath(null);
      expect(result.isValid).toBe(false);
    });

    it('rejects undefined path', () => {
      const result = FileValidator.validateFilePath(undefined);
      expect(result.isValid).toBe(false);
    });

    it('rejects path that is too long', () => {
      const result = FileValidator.validateFilePath('a'.repeat(1001));
      expect(result.isValid).toBe(false);
    });

    it('rejects path with invalid characters', () => {
      const result = FileValidator.validateFilePath('/path/with<invalid>chars');
      expect(result.isValid).toBe(false);
    });
  });

  describe('validateFileExtension', () => {
    it('accepts .txz extension', () => {
      const result = FileValidator.validateFileExtension('test.txz');
      expect(result.isValid).toBe(true);
    });

    it('accepts .tar.xz extension', () => {
      const result = FileValidator.validateFileExtension('test.tar.xz');
      expect(result.isValid).toBe(true);
    });

    it('accepts uppercase extension', () => {
      const result = FileValidator.validateFileExtension('TEST.TXZ');
      expect(result.isValid).toBe(true);
    });

    it('rejects .txt extension', () => {
      const result = FileValidator.validateFileExtension('test.txt');
      expect(result.isValid).toBe(false);
    });

    it('rejects file without extension', () => {
      const result = FileValidator.validateFileExtension('test');
      expect(result.isValid).toBe(false);
    });

    it('rejects empty filename', () => {
      const result = FileValidator.validateFileExtension('');
      expect(result.isValid).toBe(false);
    });

    it('rejects null filename', () => {
      const result = FileValidator.validateFileExtension(null);
      expect(result.isValid).toBe(false);
    });
  });

  describe('validateFileSize', () => {
    it('accepts valid file size', () => {
      const result = FileValidator.validateFileSize(1024);
      expect(result.isValid).toBe(true);
    });

    it('accepts zero file size', () => {
      const result = FileValidator.validateFileSize(0);
      expect(result.isValid).toBe(true);
    });

    it('accepts 50MB file', () => {
      const result = FileValidator.validateFileSize(50 * 1024 * 1024);
      expect(result.isValid).toBe(true);
    });

    it('rejects file that is too large', () => {
      const result = FileValidator.validateFileSize(150 * 1024 * 1024);
      expect(result.isValid).toBe(false);
    });

    it('rejects negative file size', () => {
      const result = FileValidator.validateFileSize(-1);
      expect(result.isValid).toBe(false);
    });

    it('rejects non-numeric file size', () => {
      const result = FileValidator.validateFileSize('not a number');
      expect(result.isValid).toBe(false);
    });

    it('rejects null file size', () => {
      const result = FileValidator.validateFileSize(null);
      expect(result.isValid).toBe(false);
    });
  });

  describe('validateFile', () => {
    it('accepts valid file object', () => {
      const validFile = {
        path: '/path/to/test.txz',
        name: 'test.txz',
        size: 1024
      };
      const result = FileValidator.validateFile(validFile);
      expect(result.isValid).toBe(true);
      expect(result.messages.length).toBe(0);
    });

    it('rejects file with invalid extension', () => {
      const invalidExtFile = {
        path: '/path/to/test.txt',
        name: 'test.txt',
        size: 1024
      };
      const result = FileValidator.validateFile(invalidExtFile);
      expect(result.isValid).toBe(false);
      expect(result.messages.length).toBeGreaterThan(0);
    });

    it('rejects file that is too large', () => {
      const largeFile = {
        path: '/path/to/large.txz',
        name: 'large.txz',
        size: 150 * 1024 * 1024
      };
      const result = FileValidator.validateFile(largeFile);
      expect(result.isValid).toBe(false);
    });

    it('rejects file with multiple issues', () => {
      const problematicFile = {
        path: '',
        name: 'test.txt',
        size: 200 * 1024 * 1024
      };
      const result = FileValidator.validateFile(problematicFile);
      expect(result.isValid).toBe(false);
      expect(result.messages.length).toBeGreaterThanOrEqual(2);
    });
  });
});

/**
 * InputValidator Tests
 */
describe('InputValidator', () => {
  describe('validateTextInput', () => {
    it('accepts valid text input', () => {
      const result = InputValidator.validateTextInput('valid text');
      expect(result.isValid).toBe(true);
      expect(result.value).toBe('valid text');
    });

    it('trims whitespace', () => {
      const result = InputValidator.validateTextInput('  spaced text  ');
      expect(result.isValid).toBe(true);
      expect(result.value).toBe('spaced text');
    });

    it('rejects empty string', () => {
      const result = InputValidator.validateTextInput('');
      expect(result.isValid).toBe(false);
    });

    it('rejects whitespace-only string', () => {
      const result = InputValidator.validateTextInput('   ');
      expect(result.isValid).toBe(false);
    });

    it('rejects null input', () => {
      const result = InputValidator.validateTextInput(null);
      expect(result.isValid).toBe(false);
    });

    it('rejects undefined input', () => {
      const result = InputValidator.validateTextInput(undefined);
      expect(result.isValid).toBe(false);
    });

    it('rejects non-string input', () => {
      const result = InputValidator.validateTextInput(123);
      expect(result.isValid).toBe(false);
    });
  });

  describe('validateCheckbox', () => {
    it('accepts checked checkbox', () => {
      const result = InputValidator.validateCheckbox(true, false);
      expect(result.isValid).toBe(true);
      expect(result.value).toBe(true);
    });

    it('accepts unchecked checkbox', () => {
      const result = InputValidator.validateCheckbox(false, false);
      expect(result.isValid).toBe(true);
      expect(result.value).toBe(false);
    });

    it('rejects unchecked required checkbox', () => {
      const result = InputValidator.validateCheckbox(false, true);
      expect(result.isValid).toBe(false);
    });

    it('accepts checked required checkbox', () => {
      const result = InputValidator.validateCheckbox(true, true);
      expect(result.isValid).toBe(true);
    });
  });

  describe('createDebouncedValidator', () => {
    beforeEach(() => {
      vi.useFakeTimers();
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it('delays callback execution', () => {
      let callbackCalled = false;
      let callbackValue = null;

      const mockCallback = (result) => {
        callbackCalled = true;
        callbackValue = result;
      };

      const debouncedValidator = InputValidator.createDebouncedValidator('test', mockCallback, 50);

      // Call the validator multiple times quickly
      debouncedValidator('test1');
      debouncedValidator('test2');
      debouncedValidator('test3');

      // Should not have called callback immediately due to debounce
      expect(callbackCalled).toBe(false);

      // Fast-forward past the debounce delay
      vi.advanceTimersByTime(100);

      // Should have called callback with the last value
      expect(callbackCalled).toBe(true);
      expect(callbackValue.isValid).toBe(true);
      expect(callbackValue.value).toBe('test3');
    });
  });
});

/**
 * FormValidator Tests
 */
describe('FormValidator', () => {
  describe('validateAnalysisForm', () => {
    it('accepts valid form', () => {
      const validForm = createMockForm();
      const result = FormValidator.validateAnalysisForm(validForm);
      expect(result.isValid).toBe(true);
      expect(result.messages.length).toBe(0);
    });

    it('rejects form with empty file path', () => {
      const emptyPathForm = createMockForm();
      emptyPathForm.querySelector('#filePath').value = '';
      const result = FormValidator.validateAnalysisForm(emptyPathForm);
      expect(result.isValid).toBe(false);
      expect(result.messages.length).toBeGreaterThan(0);
    });

    it('rejects form with invalid file extension', () => {
      const invalidExtForm = createMockForm();
      invalidExtForm.querySelector('#filePath').value = '/path/to/test.txt';
      const result = FormValidator.validateAnalysisForm(invalidExtForm);
      expect(result.isValid).toBe(false);
    });
  });

  describe('displayErrors', () => {
    it('creates error elements for each message', () => {
      const container = document.createElement('div');
      document.body.appendChild(container);

      FormValidator.displayErrors(container, ['Error 1', 'Error 2']);

      const errorElements = container.querySelectorAll('.validation-error');
      expect(errorElements.length).toBe(2);
      expect(errorElements[0].textContent).toContain('Error 1');
      expect(errorElements[1].textContent).toContain('Error 2');

      document.body.removeChild(container);
    });
  });

  describe('clearErrors', () => {
    it('removes all error elements', () => {
      const container = document.createElement('div');
      const error1 = document.createElement('div');
      error1.className = 'validation-error';
      error1.textContent = 'Error 1';
      const error2 = document.createElement('div');
      error2.className = 'validation-error';
      error2.textContent = 'Error 2';

      container.appendChild(error1);
      container.appendChild(error2);
      document.body.appendChild(container);

      FormValidator.clearErrors(container);

      const errorElements = container.querySelectorAll('.validation-error');
      expect(errorElements.length).toBe(0);

      document.body.removeChild(container);
    });
  });
});