/**
 * Test Suite for Frontend Validators
 * Tests for FileValidator, InputValidator, and FormValidator
 * @version 2.0.0
 */

import { FileValidator, InputValidator, FormValidator } from '../src/utils/validators.js';

/**
 * Test utilities
 */
class TestUtils {
  static assert(condition, message) {
    if (!condition) {
      throw new Error(`Assertion failed: ${message}`);
    }
  }

  static assertEqual(actual, expected, message) {
    if (actual !== expected) {
      throw new Error(`Assertion failed: ${message}. Expected: ${expected}, Actual: ${actual}`);
    }
  }

  static assertDeepEqual(actual, expected, message) {
    if (JSON.stringify(actual) !== JSON.stringify(expected)) {
      throw new Error(`Assertion failed: ${message}. Expected: ${JSON.stringify(expected)}, Actual: ${JSON.stringify(actual)}`);
    }
  }

  static createMockFile(name = 'test.txz', size = 1024) {
    return {
      name,
      size,
      type: 'application/x-xz',
      lastModified: Date.now()
    };
  }

  static createMockFileInput() {
    const input = document.createElement('input');
    input.type = 'text';
    input.id = 'filePath';
    input.value = '/path/to/test.txz';
    return input;
  }

  static createMockForm() {
    const form = document.createElement('form');
    const fileInput = this.createMockFileInput();
    const checkbox = document.createElement('input');
    checkbox.type = 'checkbox';
    checkbox.id = 'anonymizeCheck';
    checkbox.checked = false;
    
    form.appendChild(fileInput);
    form.appendChild(checkbox);
    return form;
  }
}

/**
 * FileValidator Tests
 */
class FileValidatorTests {
  static runAll() {
    console.log('Running FileValidator tests...');
    
    this.testValidateFilePath();
    this.testValidateFileExtension();
    this.testValidateFileSize();
    this.testValidateFile();
    
    console.log('FileValidator tests completed successfully!');
  }

  static testValidateFilePath() {
    console.log('Testing validateFilePath...');
    
    // Test valid paths
    let result = FileValidator.validateFilePath('/path/to/file.txz');
    TestUtils.assert(result.isValid, 'Should accept valid absolute path');
    
    result = FileValidator.validateFilePath('C:\\path\\to\\file.txz');
    TestUtils.assert(result.isValid, 'Should accept valid Windows path');
    
    // Test invalid paths
    result = FileValidator.validateFilePath('');
    TestUtils.assert(!result.isValid, 'Should reject empty path');
    
    result = FileValidator.validateFilePath(null);
    TestUtils.assert(!result.isValid, 'Should reject null path');
    
    result = FileValidator.validateFilePath(undefined);
    TestUtils.assert(!result.isValid, 'Should reject undefined path');
    
    result = FileValidator.validateFilePath('a'.repeat(1001));
    TestUtils.assert(!result.isValid, 'Should reject path that is too long');
    
    result = FileValidator.validateFilePath('/path/with<invalid>chars');
    TestUtils.assert(!result.isValid, 'Should reject path with invalid characters');
    
    console.log('validateFilePath tests passed');
  }

  static testValidateFileExtension() {
    console.log('Testing validateFileExtension...');
    
    // Test valid extensions
    let result = FileValidator.validateFileExtension('test.txz');
    TestUtils.assert(result.isValid, 'Should accept .txz extension');
    
    result = FileValidator.validateFileExtension('test.tar.xz');
    TestUtils.assert(result.isValid, 'Should accept .tar.xz extension');
    
    result = FileValidator.validateFileExtension('TEST.TXZ');
    TestUtils.assert(result.isValid, 'Should accept uppercase extension');
    
    // Test invalid extensions
    result = FileValidator.validateFileExtension('test.txt');
    TestUtils.assert(!result.isValid, 'Should reject .txt extension');
    
    result = FileValidator.validateFileExtension('test');
    TestUtils.assert(!result.isValid, 'Should reject file without extension');
    
    result = FileValidator.validateFileExtension('');
    TestUtils.assert(!result.isValid, 'Should reject empty filename');
    
    result = FileValidator.validateFileExtension(null);
    TestUtils.assert(!result.isValid, 'Should reject null filename');
    
    console.log('validateFileExtension tests passed');
  }

  static testValidateFileSize() {
    console.log('Testing validateFileSize...');
    
    // Test valid sizes
    let result = FileValidator.validateFileSize(1024);
    TestUtils.assert(result.isValid, 'Should accept valid file size');
    
    result = FileValidator.validateFileSize(0);
    TestUtils.assert(result.isValid, 'Should accept zero file size');
    
    result = FileValidator.validateFileSize(50 * 1024 * 1024); // 50MB
    TestUtils.assert(result.isValid, 'Should accept 50MB file');
    
    // Test invalid sizes
    result = FileValidator.validateFileSize(150 * 1024 * 1024); // 150MB
    TestUtils.assert(!result.isValid, 'Should reject file that is too large');
    
    result = FileValidator.validateFileSize(-1);
    TestUtils.assert(!result.isValid, 'Should reject negative file size');
    
    result = FileValidator.validateFileSize('not a number');
    TestUtils.assert(!result.isValid, 'Should reject non-numeric file size');
    
    result = FileValidator.validateFileSize(null);
    TestUtils.assert(!result.isValid, 'Should reject null file size');
    
    console.log('validateFileSize tests passed');
  }

  static testValidateFile() {
    console.log('Testing validateFile...');
    
    // Test valid file
    const validFile = {
      path: '/path/to/test.txz',
      name: 'test.txz',
      size: 1024
    };
    
    let result = FileValidator.validateFile(validFile);
    TestUtils.assert(result.isValid, 'Should accept valid file object');
    TestUtils.assertEqual(result.messages.length, 0, 'Should have no error messages');
    
    // Test file with invalid extension
    const invalidExtFile = {
      path: '/path/to/test.txt',
      name: 'test.txt',
      size: 1024
    };
    
    result = FileValidator.validateFile(invalidExtFile);
    TestUtils.assert(!result.isValid, 'Should reject file with invalid extension');
    TestUtils.assert(result.messages.length > 0, 'Should have error messages');
    
    // Test file that is too large
    const largeFile = {
      path: '/path/to/large.txz',
      name: 'large.txz',
      size: 150 * 1024 * 1024
    };
    
    result = FileValidator.validateFile(largeFile);
    TestUtils.assert(!result.isValid, 'Should reject file that is too large');
    
    // Test file with multiple issues
    const problematicFile = {
      path: '',
      name: 'test.txt',
      size: 200 * 1024 * 1024
    };
    
    result = FileValidator.validateFile(problematicFile);
    TestUtils.assert(!result.isValid, 'Should reject file with multiple issues');
    TestUtils.assert(result.messages.length >= 2, 'Should have multiple error messages');
    
    console.log('validateFile tests passed');
  }
}

/**
 * InputValidator Tests
 */
class InputValidatorTests {
  static runAll() {
    console.log('Running InputValidator tests...');
    
    this.testValidateTextInput();
    this.testValidateCheckbox();
    this.testCreateDebouncedValidator();
    
    console.log('InputValidator tests completed successfully!');
  }

  static testValidateTextInput() {
    console.log('Testing validateTextInput...');
    
    // Test valid inputs
    let result = InputValidator.validateTextInput('valid text');
    TestUtils.assert(result.isValid, 'Should accept valid text input');
    TestUtils.assertEqual(result.value, 'valid text', 'Should trim whitespace');
    
    result = InputValidator.validateTextInput('  spaced text  ');
    TestUtils.assert(result.isValid, 'Should accept text with spaces');
    TestUtils.assertEqual(result.value, 'spaced text', 'Should trim leading/trailing spaces');
    
    // Test invalid inputs
    result = InputValidator.validateTextInput('');
    TestUtils.assert(!result.isValid, 'Should reject empty string');
    
    result = InputValidator.validateTextInput('   ');
    TestUtils.assert(!result.isValid, 'Should reject whitespace-only string');
    
    result = InputValidator.validateTextInput(null);
    TestUtils.assert(!result.isValid, 'Should reject null input');
    
    result = InputValidator.validateTextInput(undefined);
    TestUtils.assert(!result.isValid, 'Should reject undefined input');
    
    result = InputValidator.validateTextInput(123);
    TestUtils.assert(!result.isValid, 'Should reject non-string input');
    
    console.log('validateTextInput tests passed');
  }

  static testValidateCheckbox() {
    console.log('Testing validateCheckbox...');
    
    // Test checkbox states
    let result = InputValidator.validateCheckbox(true, false);
    TestUtils.assert(result.isValid, 'Should accept checked checkbox');
    TestUtils.assertEqual(result.value, true, 'Should return checked state');
    
    result = InputValidator.validateCheckbox(false, false);
    TestUtils.assert(result.isValid, 'Should accept unchecked checkbox');
    TestUtils.assertEqual(result.value, false, 'Should return unchecked state');
    
    // Test required checkbox
    result = InputValidator.validateCheckbox(false, true);
    TestUtils.assert(!result.isValid, 'Should reject unchecked required checkbox');
    
    result = InputValidator.validateCheckbox(true, true);
    TestUtils.assert(result.isValid, 'Should accept checked required checkbox');
    
    console.log('validateCheckbox tests passed');
  }

  static testCreateDebouncedValidator() {
    console.log('Testing createDebouncedValidator...');
    
    let callbackCalled = false;
    let callbackValue = null;
    
    const mockCallback = (result) => {
      callbackCalled = true;
      callbackValue = result;
    };
    
    const debouncedValidator = InputValidator.createDebouncedValidator('test', mockCallback, 10);
    
    // Call the validator multiple times quickly
    debouncedValidator('test1');
    debouncedValidator('test2');
    debouncedValidator('test3');
    
    // Should not have called callback immediately due to debounce
    TestUtils.assert(!callbackCalled, 'Should not call callback immediately due to debounce');
    
    // Wait for debounce delay
    setTimeout(() => {
      TestUtils.assert(callbackCalled, 'Should call callback after debounce delay');
      TestUtils.assertEqual(callbackValue.isValid, true, 'Should validate the last value');
      TestUtils.assertEqual(callbackValue.value, 'test3', 'Should use the last input value');
    }, 20);
    
    console.log('createDebouncedValidator tests passed');
  }
}

/**
 * FormValidator Tests
 */
class FormValidatorTests {
  static runAll() {
    console.log('Running FormValidator tests...');
    
    this.testValidateAnalysisForm();
    this.testDisplayErrors();
    this.testClearErrors();
    
    console.log('FormValidator tests completed successfully!');
  }

  static testValidateAnalysisForm() {
    console.log('Testing validateAnalysisForm...');
    
    // Create mock form with valid data
    const validForm = TestUtils.createMockForm();
    validForm.querySelector('#filePath').value = '/path/to/test.txz';
    validForm.querySelector('#anonymizeCheck').checked = true;
    
    let result = FormValidator.validateAnalysisForm(validForm);
    TestUtils.assert(result.isValid, 'Should accept valid form');
    TestUtils.assertEqual(result.messages.length, 0, 'Should have no error messages');
    
    // Test form with empty file path
    const emptyPathForm = TestUtils.createMockForm();
    emptyPathForm.querySelector('#filePath').value = '';
    
    result = FormValidator.validateAnalysisForm(emptyPathForm);
    TestUtils.assert(!result.isValid, 'Should reject form with empty file path');
    TestUtils.assert(result.messages.length > 0, 'Should have error messages');
    
    // Test form with invalid file extension
    const invalidExtForm = TestUtils.createMockForm();
    invalidExtForm.querySelector('#filePath').value = '/path/to/test.txt';
    
    result = FormValidator.validateAnalysisForm(invalidExtForm);
    TestUtils.assert(!result.isValid, 'Should reject form with invalid file extension');
    
    console.log('validateAnalysisForm tests passed');
  }

  static testDisplayErrors() {
    console.log('Testing displayErrors...');
    
    // Create mock container
    const container = document.createElement('div');
    document.body.appendChild(container);
    
    // Display errors
    FormValidator.displayErrors(container, ['Error 1', 'Error 2']);
    
    const errorElements = container.querySelectorAll('.validation-error');
    TestUtils.assertEqual(errorElements.length, 2, 'Should create error elements for each message');
    
    TestUtils.assert(errorElements[0].textContent.includes('Error 1'), 'Should display first error message');
    TestUtils.assert(errorElements[1].textContent.includes('Error 2'), 'Should display second error message');
    
    // Clean up
    document.body.removeChild(container);
    
    console.log('displayErrors tests passed');
  }

  static testClearErrors() {
    console.log('Testing clearErrors...');
    
    // Create mock container with errors
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
    
    // Clear errors
    FormValidator.clearErrors(container);
    
    const errorElements = container.querySelectorAll('.validation-error');
    TestUtils.assertEqual(errorElements.length, 0, 'Should remove all error elements');
    
    // Clean up
    document.body.removeChild(container);
    
    console.log('clearErrors tests passed');
  }
}

/**
 * Test Runner
 */
class TestRunner {
  static async runAllTests() {
    console.log('Starting Frontend Validator Tests...\n');
    
    try {
      // Run all test suites
      FileValidatorTests.runAll();
      console.log('');
      
      InputValidatorTests.runAll();
      console.log('');
      
      FormValidatorTests.runAll();
      console.log('');
      
      console.log('🎉 All frontend validator tests passed successfully!');
      return true;
      
    } catch (error) {
      console.error('❌ Test failed:', error.message);
      console.error(error.stack);
      return false;
    }
  }
}

// Export for use in test runner
export {
  TestUtils,
  FileValidatorTests,
  InputValidatorTests,
  FormValidatorTests,
  TestRunner
};

// Auto-run tests if this file is executed directly
if (typeof window !== 'undefined' && window.location) {
  // Browser environment - wait for DOM
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
      TestRunner.runAllTests();
    });
  } else {
    TestRunner.runAllTests();
  }
} else if (typeof module !== 'undefined' && module.exports) {
  // Node.js environment
  module.exports = {
    TestUtils,
    FileValidatorTests,
    InputValidatorTests,
    FormValidatorTests,
    TestRunner
  };
}
