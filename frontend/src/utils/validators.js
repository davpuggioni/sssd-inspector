/**
 * Validation Utilities for SSSD Inspector Frontend
 * Provides input validation and file format checking
 * @version 2.0.0
 */

import { VALIDATION, FILES } from '../config/constants.js';

/**
 * FileValidator class for validating file inputs
 */
export class FileValidator {
  /**
   * Validates if a file path is properly formatted
   * @param {string} filePath - The file path to validate
   * @returns {Object} Validation result with isValid and message properties
   */
  static validateFilePath(filePath) {
    if (!filePath || typeof filePath !== 'string') {
      return {
        isValid: false,
        message: 'File path is required and must be a string'
      };
    }

    if (filePath.length < VALIDATION.FILE_PATH.MIN_LENGTH) {
      return {
        isValid: false,
        message: 'File path is too short'
      };
    }

    if (filePath.length > VALIDATION.FILE_PATH.MAX_LENGTH) {
      return {
        isValid: false,
        message: 'File path is too long'
      };
    }

    if (!VALIDATION.FILE_PATH.PATTERN.test(filePath)) {
      return {
        isValid: false,
        message: 'File path contains invalid characters'
      };
    }

    return {
      isValid: true,
      message: 'File path is valid'
    };
  }

  /**
   * Validates if a file has a supported extension
   * @param {string} fileName - The file name to check
   * @returns {Object} Validation result with isValid and message properties
   */
  static validateFileExtension(fileName) {
    if (!fileName || typeof fileName !== 'string') {
      return {
        isValid: false,
        message: 'File name is required'
      };
    }

    const lowerName = fileName.toLowerCase();
    // Check for multi-part extension first (.tar.xz) then single (.txz)
    const isSupported = FILES.SUPPORTED_EXTENSIONS.some(ext => lowerName.endsWith(ext));

    return {
      isValid: isSupported,
      message: isSupported 
        ? 'File format is supported' 
        : `Unsupported file format. Supported formats: ${FILES.SUPPORTED_EXTENSIONS.join(', ')}`
    };
  }

  /**
   * Validates file size against limits
   * @param {number} fileSize - File size in bytes
   * @returns {Object} Validation result with isValid and message properties
   */
  static validateFileSize(fileSize) {
    if (typeof fileSize !== 'number' || fileSize < 0) {
      return {
        isValid: false,
        message: 'Invalid file size'
      };
    }

    if (fileSize > FILES.LIMITS.MAX_FILE_SIZE) {
      const maxSizeMB = FILES.LIMITS.MAX_FILE_SIZE / (1024 * 1024);
      return {
        isValid: false,
        message: `File size exceeds maximum limit of ${maxSizeMB}MB`
      };
    }

    return {
      isValid: true,
      message: 'File size is within limits'
    };
  }

  /**
   * Comprehensive file validation
   * @param {Object} fileOptions - File validation options
   * @param {string} fileOptions.path - File path
   * @param {string} fileOptions.name - File name
   * @param {number} fileOptions.size - File size in bytes
   * @returns {Object} Comprehensive validation result
   */
  static validateFile({ path, name, size }) {
    const results = {
      path: this.validateFilePath(path),
      extension: this.validateFileExtension(name),
      size: this.validateFileSize(size),
    };

    const isValid = Object.values(results).every(result => result.isValid);
    const messages = Object.values(results)
      .filter(result => !result.isValid)
      .map(result => result.message);

    return {
      isValid,
      messages,
      details: results
    };
  }
}

/**
 * InputValidator class for form input validation
 */
export class InputValidator {
  /**
   * Validates text input with debouncing
   * @param {string} input - Input text to validate
   * @param {Function} callback - Callback function for validation result
   * @param {number} delay - Debounce delay in milliseconds
   * @returns {Function} Debounced validation function
   */
  static createDebouncedValidator(input, callback, delay = VALIDATION.INPUT.DEBOUNCE_DELAY) {
    let timeoutId;
    
    return function(value) {
      clearTimeout(timeoutId);
      timeoutId = setTimeout(() => {
        const result = InputValidator.validateTextInput(value);
        callback(result);
      }, delay);
    };
  }

  /**
   * Validates text input
   * @param {string} input - Input text to validate
   * @returns {Object} Validation result
   */
  static validateTextInput(input) {
    if (!input || typeof input !== 'string') {
      return {
        isValid: false,
        message: 'Input is required'
      };
    }

    const trimmed = input.trim();
    if (trimmed.length === 0) {
      return {
        isValid: false,
        message: 'Input cannot be empty'
      };
    }

    return {
      isValid: true,
      message: 'Input is valid',
      value: trimmed
    };
  }

  /**
   * Validates checkbox state
   * @param {boolean} isChecked - Checkbox state
   * @param {boolean} required - Whether checkbox is required
   * @returns {Object} Validation result
   */
  static validateCheckbox(isChecked, required = false) {
    if (required && !isChecked) {
      return {
        isValid: false,
        message: 'This field is required'
      };
    }

    return {
      isValid: true,
      message: 'Checkbox state is valid',
      value: isChecked
    };
  }
}

/**
 * FormValidator class for complete form validation
 */
export class FormValidator {
  /**
   * Validates the complete analysis form
   * @param {HTMLFormElement} form - The form element to validate
   * @returns {Object} Form validation result
   */
  static validateAnalysisForm(form) {
    const filePathInput = form.querySelector('#filePath');
    const anonymizeCheckbox = form.querySelector('#anonymizeCheck');
    
    const results = {
      filePath: InputValidator.validateTextInput(filePathInput?.value),
      anonymize: InputValidator.validateCheckbox(anonymizeCheckbox?.checked, false)
    };

    // Additional file validation if path looks like a file
    if (results.filePath.isValid && results.filePath.value) {
      const fileName = results.filePath.value.split('/').pop();
      const fileValidation = FileValidator.validateFileExtension(fileName);
      results.fileExtension = fileValidation;
    }

    const isValid = Object.values(results).every(result => result.isValid);
    const messages = Object.values(results)
      .filter(result => !result.isValid)
      .map(result => result.message);

    return {
      isValid,
      messages,
      details: results
    };
  }

  /**
   * Displays validation errors in the UI
   * @param {HTMLElement} container - Container element for error messages
   * @param {Array<string>} messages - Error messages to display
   */
  static displayErrors(container, messages) {
    if (!container) return;

    // Clear existing errors
    const existingErrors = container.querySelectorAll('.validation-error');
    existingErrors.forEach(error => error.remove());

    // Add new errors
    messages.forEach(message => {
      const errorElement = document.createElement('div');
      errorElement.className = 'validation-error';
      errorElement.style.cssText = `
        color: #dc3545;
        font-size: 0.875em;
        margin-top: 0.25rem;
        padding: 0.25rem 0.5rem;
        background-color: #f8d7da;
        border: 1px solid #f5c6cb;
        border-radius: 0.25rem;
      `;
      errorElement.textContent = message;
      container.appendChild(errorElement);
    });
  }

  /**
   * Clears validation errors from the UI
   * @param {HTMLElement} container - Container element with error messages
   */
  static clearErrors(container) {
    if (!container) return;
    
    const existingErrors = container.querySelectorAll('.validation-error');
    existingErrors.forEach(error => error.remove());
  }
}

/**
 * Export validation utilities for easy import
 */
export default {
  FileValidator,
  InputValidator,
  FormValidator
};
