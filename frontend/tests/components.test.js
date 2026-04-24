/**
 * Test Suite for Frontend UI Components
 * Tests for Button, ProgressBar, StatusMessage, and FileInput components
 * @version 2.0.0
 */

import { Button, ProgressBar, StatusMessage, FileInput } from '../src/components/UIComponents.js';

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

  static assertContains(actual, expected, message) {
    if (!actual.includes(expected)) {
      throw new Error(`Assertion failed: ${message}. Expected "${actual}" to contain "${expected}"`);
    }
  }

  static createMockContainer() {
    const container = document.createElement('div');
    container.id = 'test-container';
    document.body.appendChild(container);
    return container;
  }

  static cleanupMockContainer(container) {
    if (container && container.parentNode) {
      container.parentNode.removeChild(container);
    }
  }

  static waitForElement(selector, timeout = 1000) {
    return new Promise((resolve) => {
      const element = document.querySelector(selector);
      if (element) {
        resolve(element);
        return;
      }
      
      const observer = new MutationObserver(() => {
        const element = document.querySelector(selector);
        if (element) {
          observer.disconnect();
          resolve(element);
        }
      });
      
      observer.observe(document.body, { childList: true, subtree: true });
      
      setTimeout(() => {
        observer.disconnect();
        resolve(null);
      }, timeout);
    });
  }

  static triggerEvent(element, eventType, eventData = {}) {
    const event = new Event(eventType, { bubbles: true, ...eventData });
    element.dispatchEvent(event);
  }

  static triggerClick(element) {
    this.triggerEvent(element, 'click');
  }
}

/**
 * Button Component Tests
 */
class ButtonTests {
  static runAll() {
    console.log('Running Button tests...');
    
    this.testButtonCreation();
    this.testButtonStates();
    this.testButtonEvents();
    this.testButtonDestruction();
    
    console.log('Button tests completed successfully!');
  }

  static testButtonCreation() {
    console.log('Testing button creation...');
    
    const button = new Button({
      id: 'test-button',
      text: 'Test Button',
      className: 'custom-button'
    });
    
    const element = button.getElement();
    TestUtils.assert(element !== null, 'Button element should be created');
    TestUtils.assertEqual(element.id, 'test-button', 'Button should have correct ID');
    TestUtils.assertEqual(element.textContent, 'Test Button', 'Button should have correct text');
    TestUtils.assert(element.classList.contains('custom-button'), 'Button should have custom class');
    
    // Test default properties
    TestUtils.assert(!button.disabled, 'Button should be enabled by default');
    
    // Clean up
    button.destroy();
    
    console.log('Button creation tests passed');
  }

  static testButtonStates() {
    console.log('Testing button states...');
    
    const button = new Button({
      id: 'state-test-button',
      text: 'State Test'
    });
    
    const element = button.getElement();
    
    // Test disabled state
    button.setDisabled(true);
    TestUtils.assert(button.disabled, 'Button should be disabled');
    TestUtils.assert(element.disabled, 'Button element should be disabled');
    TestUtils.assertEqual(element.style.cursor, 'not-allowed', 'Button should show not-allowed cursor');
    
    // Test enabled state
    button.setDisabled(false);
    TestUtils.assert(!button.disabled, 'Button should be enabled');
    TestUtils.assert(!element.disabled, 'Button element should be enabled');
    TestUtils.assertEqual(element.style.cursor, 'pointer', 'Button should show pointer cursor');
    
    // Test text update
    button.setText('New Text');
    TestUtils.assertEqual(element.textContent, 'New Text', 'Button text should be updated');
    
    // Clean up
    button.destroy();
    
    console.log('Button states tests passed');
  }

  static testButtonEvents() {
    console.log('Testing button events...');
    
    let clickCount = 0;
    const clickHandler = () => {
      clickCount++;
    };
    
    const button = new Button({
      id: 'event-test-button',
      text: 'Event Test',
      onClick: clickHandler
    });
    
    const element = button.getElement();
    
    // Test click event
    TestUtils.triggerClick(element);
    TestUtils.assertEqual(clickCount, 1, 'Click handler should be called once');
    
    TestUtils.triggerClick(element);
    TestUtils.assertEqual(clickCount, 2, 'Click handler should be called twice');
    
    // Test disabled button doesn't trigger events
    button.setDisabled(true);
    TestUtils.triggerClick(element);
    TestUtils.assertEqual(clickCount, 2, 'Disabled button should not trigger click handler');
    
    // Clean up
    button.destroy();
    
    console.log('Button events tests passed');
  }

  static testButtonDestruction() {
    console.log('Testing button destruction...');
    
    const button = new Button({
      id: 'destroy-test-button',
      text: 'Destroy Test'
    });
    
    const element = button.getElement();
    TestUtils.assert(document.body.contains(element), 'Button element should be in DOM');
    
    button.destroy();
    TestUtils.assert(!document.body.contains(element), 'Button element should be removed from DOM');
    
    console.log('Button destruction tests passed');
  }
}

/**
 * ProgressBar Component Tests
 */
class ProgressBarTests {
  static runAll() {
    console.log('Running ProgressBar tests...');
    
    this.testProgressBarCreation();
    this.testProgressBarValues();
    this.testProgressBarStates();
    this.testProgressBarDestruction();
    
    console.log('ProgressBar tests completed successfully!');
  }

  static testProgressBarCreation() {
    console.log('Testing progress bar creation...');
    
    const progressBar = new ProgressBar({
      id: 'test-progress',
      min: 0,
      max: 100,
      value: 0,
      showPercentage: true,
      showStatus: true
    });
    
    const element = progressBar.getElement();
    TestUtils.assert(element !== null, 'Progress bar element should be created');
    TestUtils.assertEqual(element.id, 'test-progress', 'Progress bar should have correct ID');
    
    // Check for sub-elements
    const fillElement = element.querySelector('.progress-bar-fill');
    TestUtils.assert(fillElement !== null, 'Progress bar should have fill element');
    
    const percentageElement = element.querySelector('.progress-percentage');
    TestUtils.assert(percentageElement !== null, 'Progress bar should have percentage element');
    
    // Clean up
    progressBar.destroy();
    
    console.log('Progress bar creation tests passed');
  }

  static testProgressBarValues() {
    console.log('Testing progress bar values...');
    
    const progressBar = new ProgressBar({
      id: 'value-test-progress',
      showPercentage: true,
      showStatus: true
    });
    
    const element = progressBar.getElement();
    const fillElement = element.querySelector('.progress-bar-fill');
    const percentageElement = element.querySelector('.progress-percentage');
    
    // Test value setting
    progressBar.setValue(50, 'Processing...');
    TestUtils.assertEqual(progressBar.value, 50, 'Progress bar value should be set');
    TestUtils.assertEqual(fillElement.style.width, '50%', 'Fill width should match percentage');
    TestUtils.assertEqual(percentageElement.textContent, '50%', 'Percentage text should be updated');
    
    // Test boundary values
    progressBar.setValue(0, 'Starting...');
    TestUtils.assertEqual(progressBar.value, 0, 'Progress bar should handle 0 value');
    TestUtils.assertEqual(fillElement.style.width, '0%', 'Fill width should be 0%');
    
    progressBar.setValue(100, 'Complete!');
    TestUtils.assertEqual(progressBar.value, 100, 'Progress bar should handle 100 value');
    TestUtils.assertEqual(fillElement.style.width, '100%', 'Fill width should be 100%');
    
    // Test value clamping
    progressBar.setValue(150, 'Over 100');
    TestUtils.assertEqual(progressBar.value, 100, 'Progress bar should clamp to max value');
    
    progressBar.setValue(-10, 'Under 0');
    TestUtils.assertEqual(progressBar.value, 0, 'Progress bar should clamp to min value');
    
    // Clean up
    progressBar.destroy();
    
    console.log('Progress bar values tests passed');
  }

  static testProgressBarStates() {
    console.log('Testing progress bar states...');
    
    const progressBar = new ProgressBar({
      id: 'state-test-progress'
    });
    
    const element = progressBar.getElement();
    
    // Test visibility
    TestUtils.assertEqual(element.style.display, 'none', 'Progress bar should be hidden initially');
    
    progressBar.setValue(10);
    TestUtils.assertEqual(element.style.display, 'block', 'Progress bar should be visible when value > 0');
    
    progressBar.setValue(0);
    TestUtils.assertEqual(element.style.display, 'none', 'Progress bar should be hidden when value = 0');
    
    // Test reset
    progressBar.setValue(50);
    progressBar.reset();
    TestUtils.assertEqual(progressBar.value, 0, 'Progress bar value should be reset');
    TestUtils.assertEqual(element.style.display, 'none', 'Progress bar should be hidden after reset');
    
    // Clean up
    progressBar.destroy();
    
    console.log('Progress bar states tests passed');
  }

  static testProgressBarDestruction() {
    console.log('Testing progress bar destruction...');
    
    const progressBar = new ProgressBar({
      id: 'destroy-test-progress'
    });
    
    const element = progressBar.getElement();
    TestUtils.assert(document.body.contains(element), 'Progress bar element should be in DOM');
    
    progressBar.destroy();
    TestUtils.assert(!document.body.contains(element), 'Progress bar element should be removed from DOM');
    
    console.log('Progress bar destruction tests passed');
  }
}

/**
 * StatusMessage Component Tests
 */
class StatusMessageTests {
  static runAll() {
    console.log('Running StatusMessage tests...');
    
    this.testStatusMessageCreation();
    this.testStatusMessageTypes();
    this.testStatusMessageUpdates();
    this.testStatusMessageDestruction();
    
    console.log('StatusMessage tests completed successfully!');
  }

  static testStatusMessageCreation() {
    console.log('Testing status message creation...');
    
    const statusMessage = new StatusMessage({
      id: 'test-status',
      type: 'info',
      message: 'Test message',
      dismissible: true,
      autoHide: 0 // Don't auto-hide for testing
    });
    
    const element = statusMessage.getElement();
    TestUtils.assert(element !== null, 'Status message element should be created');
    TestUtils.assertEqual(element.id, 'test-status', 'Status message should have correct ID');
    TestUtils.assert(element.classList.contains('status-info'), 'Status message should have type class');
    
    // Check for sub-elements
    const contentElement = element.querySelector('.status-content');
    TestUtils.assert(contentElement !== null, 'Status message should have content element');
    TestUtils.assert(contentElement.innerHTML.includes('Test message'), 'Content should contain message');
    
    const dismissElement = element.querySelector('.status-dismiss');
    TestUtils.assert(dismissElement !== null, 'Status message should have dismiss button');
    
    // Clean up
    statusMessage.destroy();
    
    console.log('Status message creation tests passed');
  }

  static testStatusMessageTypes() {
    console.log('Testing status message types...');
    
    const types = ['info', 'success', 'warning', 'error'];
    
    types.forEach(type => {
      const statusMessage = new StatusMessage({
        id: `${type}-test-status`,
        type: type,
        message: `${type} message`,
        dismissible: false,
        autoHide: 0
      });
      
      const element = statusMessage.getElement();
      TestUtils.assert(element.classList.contains(`status-${type}`), `Status message should have ${type} class`);
      
      // Check dismiss button
      const dismissElement = element.querySelector('.status-dismiss');
      if (type === 'info') {
        TestUtils.assert(dismissElement !== null, 'Dismissible status should have dismiss button');
      } else {
        TestUtils.assert(dismissElement === null, 'Non-dismissible status should not have dismiss button');
      }
      
      statusMessage.destroy();
    });
    
    console.log('Status message types tests passed');
  }

  static testStatusMessageUpdates() {
    console.log('Testing status message updates...');
    
    const statusMessage = new StatusMessage({
      id: 'update-test-status',
      type: 'info',
      message: 'Initial message',
      dismissible: true,
      autoHide: 0
    });
    
    const element = statusMessage.getElement();
    
    // Test update
    statusMessage.update('Updated message', 'success');
    
    const contentElement = element.querySelector('.status-content');
    TestUtils.assert(contentElement.innerHTML.includes('Updated message'), 'Content should be updated');
    TestUtils.assert(element.classList.contains('status-success'), 'Type should be updated');
    TestUtils.assert(!element.classList.contains('status-info'), 'Old type should be removed');
    
    // Clean up
    statusMessage.destroy();
    
    console.log('Status message updates tests passed');
  }

  static testStatusMessageDestruction() {
    console.log('Testing status message destruction...');
    
    const statusMessage = new StatusMessage({
      id: 'destroy-test-status',
      message: 'Destroy test'
    });
    
    const element = statusMessage.getElement();
    TestUtils.assert(document.body.contains(element), 'Status message element should be in DOM');
    
    statusMessage.destroy();
    TestUtils.assert(!document.body.contains(element), 'Status message element should be removed from DOM');
    
    console.log('Status message destruction tests passed');
  }
}

/**
 * FileInput Component Tests
 */
class FileInputTests {
  static runAll() {
    console.log('Running FileInput tests...');
    
    this.testFileInputCreation();
    this.testFileInputEvents();
    this.testFileInputValidation();
    this.testFileInputDestruction();
    
    console.log('FileInput tests completed successfully!');
  }

  static testFileInputCreation() {
    console.log('Testing file input creation...');
    
    const fileInput = new FileInput({
      id: 'test-file-input',
      placeholder: 'Select file...',
      accept: '.txz,.tar.xz'
    });
    
    const element = fileInput.getElement();
    TestUtils.assert(element !== null, 'File input element should be created');
    TestUtils.assertEqual(element.id, 'test-file-input', 'File input should have correct ID');
    
    // Check for sub-elements
    const textInput = element.querySelector('.file-path-input');
    TestUtils.assert(textInput !== null, 'File input should have text input');
    TestUtils.assertEqual(textInput.placeholder, 'Select file...', 'Text input should have correct placeholder');
    
    const browseButton = element.querySelector('.browse-button');
    TestUtils.assert(browseButton !== null, 'File input should have browse button');
    
    const hiddenInput = element.querySelector('input[type="file"]');
    TestUtils.assert(hiddenInput !== null, 'File input should have hidden file input');
    TestUtils.assertEqual(hiddenInput.accept, '.txz,.tar.xz', 'Hidden input should have correct accept attribute');
    
    // Clean up
    fileInput.destroy();
    
    console.log('File input creation tests passed');
  }

  static testFileInputEvents() {
    console.log('Testing file input events...');
    
    let fileSelectCount = 0;
    let lastFileData = null;
    
    const fileSelectHandler = (fileData) => {
      fileSelectCount++;
      lastFileData = fileData;
    };
    
    const fileInput = new FileInput({
      id: 'event-test-file-input',
      onFileSelect: fileSelectHandler
    });
    
    const element = fileInput.getElement();
    const textInput = element.querySelector('.file-path-input');
    
    // Test text input change
    textInput.value = '/path/to/test.txz';
    TestUtils.triggerEvent(textInput, 'input');
    
    TestUtils.assertEqual(fileSelectCount, 1, 'File select handler should be called');
    TestUtils.assertEqual(lastFileData.path, '/path/to/test.txz', 'File data should contain path');
    TestUtils.assertEqual(lastFileData.name, 'test.txz', 'File data should contain name');
    
    // Test empty input
    textInput.value = '';
    TestUtils.triggerEvent(textInput, 'input');
    
    // Handler should not be called for empty input (implementation dependent)
    
    // Clean up
    fileInput.destroy();
    
    console.log('File input events tests passed');
  }

  static testFileInputValidation() {
    console.log('Testing file input validation...');
    
    const fileInput = new FileInput({
      id: 'validation-test-file-input'
    });
    
    // Test value setting and getting
    fileInput.setValue('/path/to/test.txz');
    TestUtils.assertEqual(fileInput.getValue(), '/path/to/test.txz', 'File input should return correct value');
    
    fileInput.setValue('');
    TestUtils.assertEqual(fileInput.getValue(), '', 'File input should return empty value');
    
    // Test clear
    fileInput.setValue('/path/to/test.txz');
    fileInput.clear();
    TestUtils.assertEqual(fileInput.getValue(), '', 'File input should be cleared');
    
    // Test disabled state
    fileInput.setDisabled(true);
    const textInput = fileInput.getElement().querySelector('.file-path-input');
    TestUtils.assert(textInput.disabled, 'Text input should be disabled');
    
    fileInput.setDisabled(false);
    TestUtils.assert(!textInput.disabled, 'Text input should be enabled');
    
    // Clean up
    fileInput.destroy();
    
    console.log('File input validation tests passed');
  }

  static testFileInputDestruction() {
    console.log('Testing file input destruction...');
    
    const fileInput = new FileInput({
      id: 'destroy-test-file-input'
    });
    
    const element = fileInput.getElement();
    TestUtils.assert(document.body.contains(element), 'File input element should be in DOM');
    
    fileInput.destroy();
    TestUtils.assert(!document.body.contains(element), 'File input element should be removed from DOM');
    
    console.log('File input destruction tests passed');
  }
}

/**
 * Test Runner
 */
class TestRunner {
  static async runAllTests() {
    console.log('Starting Frontend Component Tests...\n');
    
    try {
      // Run all test suites
      ButtonTests.runAll();
      console.log('');
      
      ProgressBarTests.runAll();
      console.log('');
      
      StatusMessageTests.runAll();
      console.log('');
      
      FileInputTests.runAll();
      console.log('');
      
      console.log('🎉 All frontend component tests passed successfully!');
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
  ButtonTests,
  ProgressBarTests,
  StatusMessageTests,
  FileInputTests,
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
    ButtonTests,
    ProgressBarTests,
    StatusMessageTests,
    FileInputTests,
    TestRunner
  };
}
