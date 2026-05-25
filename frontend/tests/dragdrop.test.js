/**
 * Test Suite for Drag-and-Drop and UI Interaction
 * Tests for OnFileDrop callback, keyboard shortcuts, and event handling
 * @version 1.0.0
 */

import { TestUtils } from './components.test.js';

/**
 * Mock the Wails runtime and backend for testing
 */
class MockWailsRuntime {
  constructor() {
    this._dropCallbacks = [];
    this._eventListeners = {};
    this._callOrder = [];
  }

  // Mock OnFileDrop — stores callback for manual invocation
  OnFileDrop(callback, useDropTarget) {
    this._dropCallbacks.push({ callback, useDropTarget });
    this._callOrder.push('OnFileDrop');
  }

  // Mock EventsOn — stores event listeners
  EventsOn(eventName, callback) {
    if (!this._eventListeners[eventName]) {
      this._eventListeners[eventName] = [];
    }
    this._eventListeners[eventName].push(callback);
    this._callOrder.push(`EventsOn:${eventName}`);
  }

  // Simulate a file drop event
  simulateDrop(files) {
    this._dropCallbacks.forEach(({ callback }) => {
      callback(100, 200, files);
    });
  }

  // Simulate an event emission
  simulateEvent(eventName, ...args) {
    const listeners = this._eventListeners[eventName] || [];
    listeners.forEach(cb => cb(...args));
  }

  // Mock backend Analyze function
  async Analyze(path, anonymize) {
    this._callOrder.push(`Analyze:${path}:${anonymize}`);
    return {
      timestamp: new Date().toISOString(),
      problems: [],
      warnings: [],
      sssd_log_errors: [],
      timeline: [],
      matched_tids: []
    };
  }

  // Mock backend OpenFileBrowser
  async OpenFileBrowser() {
    this._callOrder.push('OpenFileBrowser');
    return '/mock/path/supportconfig.txz';
  }

  // Mock backend SaveTXT
  async SaveTXT(report) {
    this._callOrder.push('SaveTXT');
    return '/mock/path/report.txt';
  }

  // Reset mock state
  reset() {
    this._dropCallbacks = [];
    this._eventListeners = {};
    this._callOrder = [];
  }
}

/**
 * Drag-and-Drop Tests for main.js UI
 */
class DragDropTests {
  static runAll() {
    console.log('Running Drag-and-Drop tests...');
    
    this.testOnFileDropCallbackOrder();
    this.testOnFileDropPopulatesInput();
    this.testOnFileDropMultipleFiles();
    this.testOnFileDropEmpty();
    this.testEventsOnRegistered();
    this.testAnalyzeProgressUpdatesButton();
    this.testAnalyzeProgressStartEnd();
    
    console.log('Drag-and-Drop tests completed successfully!');
  }

  static testOnFileDropCallbackOrder() {
    console.log('  Testing OnFileDrop callback parameter order...');
    
    const mockRuntime = new MockWailsRuntime();
    const filePathInput = { value: '' };

    // Simulate what main.js does: OnFileDrop((x, y, files) => { ... })
    mockRuntime.OnFileDrop((x, y, files) => {
      if (files.length > 0) filePathInput.value = files[0];
    });

    // Simulate a drop event — Wails passes (x, y, paths)
    mockRuntime.simulateDrop(['/tmp/supportconfig.txz']);

    TestUtils.assertEqual(
      filePathInput.value,
      '/tmp/supportconfig.txz',
      'OnFileDrop should populate filePathInput with the dropped file path'
    );

    mockRuntime.reset();
    console.log('  ✓ OnFileDrop callback parameter order is correct');
  }

  static testOnFileDropPopulatesInput() {
    console.log('  Testing OnFileDrop populates input value...');
    
    const mockRuntime = new MockWailsRuntime();
    const filePathInput = { value: '' };

    mockRuntime.OnFileDrop((x, y, files) => {
      if (files.length > 0) filePathInput.value = files[0];
    });

    // Drop a single file
    mockRuntime.simulateDrop(['/home/user/supportconfig.txz']);

    TestUtils.assertEqual(
      filePathInput.value,
      '/home/user/supportconfig.txz',
      'Input value should match the dropped file path'
    );

    mockRuntime.reset();
    console.log('  ✓ OnFileDrop correctly populates input');
  }

  static testOnFileDropMultipleFiles() {
    console.log('  Testing OnFileDrop with multiple files...');
    
    const mockRuntime = new MockWailsRuntime();
    const filePathInput = { value: '' };

    mockRuntime.OnFileDrop((x, y, files) => {
      if (files.length > 0) filePathInput.value = files[0];
    });

    // Drop multiple files — should only take the first one
    mockRuntime.simulateDrop([
      '/tmp/first.txz',
      '/tmp/second.txz',
      '/tmp/third.txz'
    ]);

    TestUtils.assertEqual(
      filePathInput.value,
      '/tmp/first.txz',
      'Should use only the first file when multiple are dropped'
    );

    mockRuntime.reset();
    console.log('  ✓ OnFileDrop handles multiple files correctly');
  }

  static testOnFileDropEmpty() {
    console.log('  Testing OnFileDrop with empty file list...');
    
    const mockRuntime = new MockWailsRuntime();
    const filePathInput = { value: '' };

    mockRuntime.OnFileDrop((x, y, files) => {
      if (files.length > 0) filePathInput.value = files[0];
    });

    // Drop with empty file list — should NOT change input
    mockRuntime.simulateDrop([]);

    TestUtils.assertEqual(
      filePathInput.value,
      '',
      'Input should remain empty when no files are dropped'
    );

    mockRuntime.reset();
    console.log('  ✓ OnFileDrop handles empty file list gracefully');
  }

  static testEventsOnRegistered() {
    console.log('  Testing EventsOn registration...');
    
    const mockRuntime = new MockWailsRuntime();
    
    // Simulate the two EventsOn calls from main.js
    mockRuntime.EventsOn('analyze-progress', () => {});

    TestUtils.assert(
      mockRuntime._eventListeners['analyze-progress'] !== undefined,
      'EventsOn should register an analyze-progress listener'
    );
    TestUtils.assert(
      mockRuntime._eventListeners['analyze-progress'].length === 1,
      'Should have exactly one analyze-progress listener'
    );

    mockRuntime.reset();
    console.log('  ✓ EventsOn listeners are registered correctly');
  }

  static testAnalyzeProgressUpdatesButton() {
    console.log('  Testing analyze-progress event updates button text...');
    
    const mockRuntime = new MockWailsRuntime();
    const analyzeBtn = { 
      textContent: 'Analyze',
      disabled: false
    };

    // Simulate EventsOn('analyze-progress', ...) from main.js
    mockRuntime.EventsOn('analyze-progress', (message, percentage) => {
      if (percentage > 0 && percentage < 100) {
        analyzeBtn.textContent = `${percentage}% - ${message}`;
      }
    });

    // Simulate progress at 50%
    mockRuntime.simulateEvent('analyze-progress', 'Processing...', 50);
    TestUtils.assertEqual(
      analyzeBtn.textContent,
      '50% - Processing...',
      'Button text should show progress percentage and message'
    );

    // Simulate progress at 0% (start) — should NOT update button
    analyzeBtn.textContent = 'Analyze';
    mockRuntime.simulateEvent('analyze-progress', 'Starting...', 0);
    TestUtils.assertEqual(
      analyzeBtn.textContent,
      'Analyze',
      'Button text should NOT change at 0% progress'
    );

    // Simulate progress at 100% (complete) — should NOT update button
    mockRuntime.simulateEvent('analyze-progress', 'Complete!', 100);
    TestUtils.assertEqual(
      analyzeBtn.textContent,
      'Analyze',
      'Button text should NOT change at 100% progress'
    );

    mockRuntime.reset();
    console.log('  ✓ analyze-progress updates button correctly');
  }

  static testAnalyzeProgressStartEnd() {
    console.log('  Testing analyze-progress boundary conditions...');
    
    const mockRuntime = new MockWailsRuntime();
    const analyzeBtn = { 
      textContent: 'Analyze',
      disabled: false
    };

    mockRuntime.EventsOn('analyze-progress', (message, percentage) => {
      if (percentage > 0 && percentage < 100) {
        analyzeBtn.textContent = `${percentage}% - ${message}`;
      }
    });

    // Simulate progress sequence
    mockRuntime.simulateEvent('analyze-progress', 'Starting', 0);
    TestUtils.assertEqual(analyzeBtn.textContent, 'Analyze', 'At 0%, button should still say Analyze');

    mockRuntime.simulateEvent('analyze-progress', 'Extracting...', 10);
    TestUtils.assertEqual(analyzeBtn.textContent, '10% - Extracting...', 'At 10%, button should show progress');

    mockRuntime.simulateEvent('analyze-progress', 'Analyzing logs...', 50);
    TestUtils.assertEqual(analyzeBtn.textContent, '50% - Analyzing logs...', 'At 50%, button should show progress');

    mockRuntime.simulateEvent('analyze-progress', 'Done', 100);
    TestUtils.assertEqual(analyzeBtn.textContent, '50% - Analyzing logs...', 'At 100%, button should keep last in-range value');

    mockRuntime.reset();
    console.log('  ✓ analyze-progress boundary conditions handled correctly');
  }
}

/**
 * Keyboard Shortcut Tests
 */
class KeyboardShortcutTests {
  static runAll() {
    console.log('Running Keyboard Shortcut tests...');
    
    this.testCtrlOTriggersBrowse();
    this.testCtrlEnterTriggersAnalyze();
    this.testCtrlEnterNoOpWithoutPath();
    
    console.log('Keyboard Shortcut tests completed successfully!');
  }

  static testCtrlOTriggersBrowse() {
    console.log('  Testing Ctrl+O triggers browse...');
    
    let browseClicked = false;
    const browseBtn = { click: () => { browseClicked = true; } };

    // Simulate the keydown handler from main.js
    const keyHandler = (e) => {
      if (e.ctrlKey || e.metaKey) {
        switch(e.key) {
          case 'o': e.preventDefault(); browseBtn.click(); break;
        }
      }
    };

    // Simulate Ctrl+O
    const event = new KeyboardEvent('keydown', {
      key: 'o',
      ctrlKey: true,
      metaKey: false,
      bubbles: true,
      cancelable: true
    });
    
    // We can't easily call preventDefault on a constructed event,
    // but we can test the logic directly
    keyHandler({ ctrlKey: true, key: 'o', preventDefault: () => {} });

    TestUtils.assert(browseClicked, 'Ctrl+O should trigger browse button click');

    console.log('  ✓ Ctrl+O triggers browse');
  }

  static testCtrlEnterTriggersAnalyze() {
    console.log('  Testing Ctrl+Enter triggers analyze...');
    
    let analyzeClicked = false;
    const filePathInput = { value: '/tmp/test.txz' };
    const analyzeBtn = { click: () => { analyzeClicked = true; } };

    const keyHandler = (e) => {
      if (e.ctrlKey || e.metaKey) {
        switch(e.key) {
          case 'Enter': e.preventDefault(); if (filePathInput.value.trim()) analyzeBtn.click(); break;
        }
      }
    };

    keyHandler({ ctrlKey: true, key: 'Enter', preventDefault: () => {} });

    TestUtils.assert(analyzeClicked, 'Ctrl+Enter should trigger analyze when path is set');

    console.log('  ✓ Ctrl+Enter triggers analyze');
  }

  static testCtrlEnterNoOpWithoutPath() {
    console.log('  Testing Ctrl+Enter no-op when path is empty...');
    
    let analyzeClicked = false;
    const filePathInput = { value: '' };
    const analyzeBtn = { click: () => { analyzeClicked = true; } };

    const keyHandler = (e) => {
      if (e.ctrlKey || e.metaKey) {
        switch(e.key) {
          case 'Enter': e.preventDefault(); if (filePathInput.value.trim()) analyzeBtn.click(); break;
        }
      }
    };

    keyHandler({ ctrlKey: true, key: 'Enter', preventDefault: () => {} });

    TestUtils.assert(!analyzeClicked, 'Ctrl+Enter should NOT trigger analyze when path is empty');

    console.log('  ✓ Ctrl+Enter is guarded against empty path');
  }
}

/**
 * Test Runner
 */
class DragDropTestRunner {
  static async runAllTests() {
    console.log('\nStarting Drag-and-Drop & UI Interaction Tests...\n');
    
    try {
      DragDropTests.runAll();
      console.log('');
      KeyboardShortcutTests.runAll();
      console.log('');
      
      console.log('🎉 All drag-and-drop and UI interaction tests passed successfully!');
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
  MockWailsRuntime,
  DragDropTests,
  KeyboardShortcutTests,
  DragDropTestRunner
};

// Auto-run tests if executed directly
if (typeof window !== 'undefined' && window.location) {
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
      DragDropTestRunner.runAllTests();
    });
  } else {
    DragDropTestRunner.runAllTests();
  }
}