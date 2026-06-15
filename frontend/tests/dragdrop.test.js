/**
 * Test Suite for Drag-and-Drop and UI Interaction
 * Tests for OnFileDrop callback, keyboard shortcuts, and event handling
 * @version 1.0.0
 */

import { describe, it, expect } from 'vitest';

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

  // Mock backend functions
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

  async OpenFileBrowser() {
    this._callOrder.push('OpenFileBrowser');
    return '/mock/path/supportconfig.txz';
  }

  async SaveTXT(report) {
    this._callOrder.push('SaveTXT');
    return '/mock/path/report.txt';
  }

  reset() {
    this._dropCallbacks = [];
    this._eventListeners = {};
    this._callOrder = [];
  }
}

/**
 * Drag-and-Drop Tests
 */
describe('Drag-and-Drop', () => {
  it('OnFileDrop callback populates input with dropped file', () => {
    const mockRuntime = new MockWailsRuntime();
    const filePathInput = { value: '' };

    mockRuntime.OnFileDrop((x, y, files) => {
      if (files.length > 0) filePathInput.value = files[0];
    });

    mockRuntime.simulateDrop(['/tmp/supportconfig.txz']);

    expect(filePathInput.value).toBe('/tmp/supportconfig.txz');
  });

  it('OnFileDrop handles multiple files by taking the first one', () => {
    const mockRuntime = new MockWailsRuntime();
    const filePathInput = { value: '' };

    mockRuntime.OnFileDrop((x, y, files) => {
      if (files.length > 0) filePathInput.value = files[0];
    });

    mockRuntime.simulateDrop(['/tmp/first.txz', '/tmp/second.txz', '/tmp/third.txz']);

    expect(filePathInput.value).toBe('/tmp/first.txz');
  });

  it('OnFileDrop handles empty file list gracefully', () => {
    const mockRuntime = new MockWailsRuntime();
    const filePathInput = { value: '' };

    mockRuntime.OnFileDrop((x, y, files) => {
      if (files.length > 0) filePathInput.value = files[0];
    });

    mockRuntime.simulateDrop([]);

    expect(filePathInput.value).toBe('');
  });

  it('EventsOn registers analyze-progress listener', () => {
    const mockRuntime = new MockWailsRuntime();

    mockRuntime.EventsOn('analyze-progress', () => {});

    expect(mockRuntime._eventListeners['analyze-progress']).toBeDefined();
    expect(mockRuntime._eventListeners['analyze-progress'].length).toBe(1);
  });
});

/**
 * Window Events Tests (analyze-progress)
 */
describe('Analysis Progress Events', () => {
  it('updates button text with progress percentage and message', () => {
    const mockRuntime = new MockWailsRuntime();
    const analyzeBtn = { textContent: 'Analyze', disabled: false };

    mockRuntime.EventsOn('analyze-progress', (message, percentage) => {
      if (percentage > 0 && percentage < 100) {
        analyzeBtn.textContent = `${percentage}% - ${message}`;
      }
    });

    mockRuntime.simulateEvent('analyze-progress', 'Processing...', 50);
    expect(analyzeBtn.textContent).toBe('50% - Processing...');
  });

  it('does not update button at 0% progress', () => {
    const mockRuntime = new MockWailsRuntime();
    const analyzeBtn = { textContent: 'Analyze', disabled: false };

    mockRuntime.EventsOn('analyze-progress', (message, percentage) => {
      if (percentage > 0 && percentage < 100) {
        analyzeBtn.textContent = `${percentage}% - ${message}`;
      }
    });

    mockRuntime.simulateEvent('analyze-progress', 'Starting...', 0);
    expect(analyzeBtn.textContent).toBe('Analyze');
  });

  it('does not update button at 100% progress', () => {
    const mockRuntime = new MockWailsRuntime();
    const analyzeBtn = { textContent: 'Analyze', disabled: false };

    mockRuntime.EventsOn('analyze-progress', (message, percentage) => {
      if (percentage > 0 && percentage < 100) {
        analyzeBtn.textContent = `${percentage}% - ${message}`;
      }
    });

    mockRuntime.simulateEvent('analyze-progress', 'Complete!', 100);
    expect(analyzeBtn.textContent).toBe('Analyze');
  });

  it('handles full progress sequence correctly', () => {
    const mockRuntime = new MockWailsRuntime();
    const analyzeBtn = { textContent: 'Analyze', disabled: false };

    mockRuntime.EventsOn('analyze-progress', (message, percentage) => {
      if (percentage > 0 && percentage < 100) {
        analyzeBtn.textContent = `${percentage}% - ${message}`;
      }
    });

    mockRuntime.simulateEvent('analyze-progress', 'Starting', 0);
    expect(analyzeBtn.textContent).toBe('Analyze');

    mockRuntime.simulateEvent('analyze-progress', 'Extracting...', 10);
    expect(analyzeBtn.textContent).toBe('10% - Extracting...');

    mockRuntime.simulateEvent('analyze-progress', 'Analyzing logs...', 50);
    expect(analyzeBtn.textContent).toBe('50% - Analyzing logs...');

    mockRuntime.simulateEvent('analyze-progress', 'Done', 100);
    expect(analyzeBtn.textContent).toBe('50% - Analyzing logs...');
  });
});

/**
 * Keyboard Shortcut Tests
 */
describe('Keyboard Shortcuts', () => {
  it('Ctrl+O triggers browse', () => {
    let browseClicked = false;
    const browseBtn = { click: () => { browseClicked = true; } };

    const keyHandler = (e) => {
      if (e.ctrlKey || e.metaKey) {
        switch(e.key) {
          case 'o': e.preventDefault(); browseBtn.click(); break;
        }
      }
    };

    keyHandler({ ctrlKey: true, key: 'o', preventDefault: () => {} });

    expect(browseClicked).toBe(true);
  });

  it('Ctrl+Enter triggers analyze when path is set', () => {
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

    expect(analyzeClicked).toBe(true);
  });

  it('Ctrl+Enter is no-op when path is empty', () => {
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

    expect(analyzeClicked).toBe(false);
  });
});