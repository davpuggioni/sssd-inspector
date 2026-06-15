/**
 * UI Components for SSSD Inspector Frontend
 * Reusable UI components with clean architecture
 * @version 2.0.0
 */

import { UI, CSS_CLASSES, COLORS } from '../config/constants.js';
import { DOMHelper, FormatHelper } from '../utils/helpers.js';

/**
 * Button component with consistent styling and behavior
 */
export class Button {
  /**
   * Creates a new button instance
   * @param {Object} options - Button configuration
   */
  constructor(options = {}) {
    this.id = options.id || `btn-${Date.now()}`;
    this.text = options.text || '';
    this.className = options.className || CSS_CLASSES.COMPONENTS.BUTTON;
    this.onClick = options.onClick || null;
    this.disabled = options.disabled || false;
    this.styles = options.styles || {};
    this.attributes = options.attributes || {};
    
    this.element = this.createElement();
  }

  /**
   * Creates the button DOM element
   * @returns {HTMLElement} Button element
   */
  createElement() {
    const attrs = {
      id: this.id,
      type: 'button',
      ...this.attributes
    };
    if (this.disabled) {
      attrs.disabled = true;
    }
    const button = DOMHelper.createElement('button', {
      attributes: attrs,
      className: this.className,
      textContent: this.text,
      styles: {
        padding: '10px 15px',
        fontSize: '14px',
        cursor: this.disabled ? 'not-allowed' : 'pointer',
        backgroundColor: COLORS.PRIMARY.MAIN,
        color: COLORS.TEXT.WHITE,
        border: 'none',
        borderRadius: '4px',
        fontWeight: 'bold',
        opacity: this.disabled ? '0.6' : '1',
        ...this.styles
      }
    });

    this.element = button;

    if (this.onClick) {
      this.addClickListener();
    }

    return button;
  }

  /**
   * Adds click event listener
   */
  addClickListener() {
    if (!this.disabled && this.onClick) {
      this.cleanup = DOMHelper.addEventListener(this.element, 'click', this.onClick);
    }
  }

  /**
   * Updates button text
   * @param {string} text - New button text
   */
  setText(text) {
    this.text = text;
    DOMHelper.updateContent(this.element, text);
  }

  /**
   * Updates button disabled state
   * @param {boolean} disabled - Disabled state
   */
  setDisabled(disabled) {
    this.disabled = disabled;
    this.element.disabled = disabled;
    this.element.style.cursor = disabled ? 'not-allowed' : 'pointer';
    this.element.style.opacity = disabled ? '0.6' : '1';
    
    // Re-add click listener if enabling
    if (!disabled && this.onClick && !this.cleanup) {
      this.addClickListener();
    }
    
    // Remove click listener if disabling
    if (disabled && this.cleanup) {
      this.cleanup();
      this.cleanup = null;
    }
  }

  /**
   * Shows or hides the button
   * @param {boolean} show - Whether to show the button
   */
  toggle(show) {
    DOMHelper.toggleElement(this.element, show, { duration: 200 });
  }

  /**
   * Gets the button element
   * @returns {HTMLElement} Button element
   */
  getElement() {
    return this.element;
  }

  /**
   * Destroys the button and cleans up event listeners
   */
  destroy() {
    if (this.cleanup) {
      this.cleanup();
    }
    if (this.element && this.element.parentNode) {
      this.element.parentNode.removeChild(this.element);
    }
  }
}

/**
 * ProgressBar component for showing analysis progress
 */
export class ProgressBar {
  /**
   * Creates a new progress bar instance
   * @param {Object} options - Progress bar configuration
   */
  constructor(options = {}) {
    this.id = options.id || `progress-${Date.now()}`;
    this.min = options.min || 0;
    this.max = options.max || 100;
    this.value = options.value || 0;
    this.showPercentage = options.showPercentage !== false;
    this.showStatus = options.showStatus !== false;
    this.status = options.status || '';
    
    this.element = this.createElement();
  }

  /**
   * Creates the progress bar DOM element
   * @returns {HTMLElement} Progress bar element
   */
  createElement() {
    const container = DOMHelper.createElement('div', {
      attributes: { id: this.id },
      className: 'progress-container',
      styles: {
        width: '100%',
        margin: '10px 0',
        display: this.value > 0 ? 'block' : 'none'
      }
    });

    // Status text
    if (this.showStatus) {
      this.statusElement = DOMHelper.createElement('div', {
        className: 'progress-status',
        styles: {
          fontSize: '0.875em',
          color: COLORS.TEXT.SECONDARY,
          marginBottom: '5px',
          textAlign: 'center'
        },
        textContent: this.status
      });
      container.appendChild(this.statusElement);
    }

    // Progress bar wrapper
    const barWrapper = DOMHelper.createElement('div', {
      className: 'progress-bar-wrapper',
      styles: {
        width: '100%',
        height: '20px',
        backgroundColor: COLORS.BACKGROUND.LIGHT,
        borderRadius: '10px',
        overflow: 'hidden',
        border: '1px solid #dee2e6'
      }
    });

    // Progress bar fill
    this.fillElement = DOMHelper.createElement('div', {
      className: 'progress-bar-fill',
      styles: {
        height: '100%',
        backgroundColor: COLORS.PRIMARY.MAIN,
        width: '0%',
        transition: 'width 0.3s ease-in-out',
        borderRadius: '8px'
      }
    });

    barWrapper.appendChild(this.fillElement);
    container.appendChild(barWrapper);

    // Percentage text
    if (this.showPercentage) {
      this.percentageElement = DOMHelper.createElement('div', {
        className: 'progress-percentage',
        styles: {
          fontSize: '0.875em',
          color: COLORS.TEXT.SECONDARY,
          marginTop: '5px',
          textAlign: 'center'
        },
        textContent: '0%'
      });
      container.appendChild(this.percentageElement);
    }

    return container;
  }

  /**
   * Updates progress value
   * @param {number} value - New progress value
   * @param {string} status - Status message (optional)
   */
  setValue(value, status = '') {
    this.value = Math.max(this.min, Math.min(this.max, value));
    
    // Update fill width
    const percentage = ((this.value - this.min) / (this.max - this.min)) * 100;
    this.fillElement.style.width = `${percentage}%`;

    // Update percentage text
    if (this.percentageElement) {
      this.percentageElement.textContent = `${Math.round(percentage)}%`;
    }

    // Update status text
    if (this.statusElement && status) {
      this.statusElement.textContent = status;
      this.status = status;
    }

    // Show/hide based on value
    if (this.value > 0) {
      this.element.style.display = 'block';
    } else {
      this.element.style.display = 'none';
    }
  }

  /**
   * Shows or hides the progress bar
   * @param {boolean} show - Whether to show the progress bar
   */
  toggle(show) {
    DOMHelper.toggleElement(this.element, show, { duration: 200 });
  }

  /**
   * Resets the progress bar
   */
  reset() {
    this.setValue(0);
    this.status = '';
    if (this.statusElement) {
      this.statusElement.textContent = '';
    }
  }

  /**
   * Gets the progress bar element
   * @returns {HTMLElement} Progress bar element
   */
  getElement() {
    return this.element;
  }

  /**
   * Destroys the progress bar
   */
  destroy() {
    if (this.element && this.element.parentNode) {
      this.element.parentNode.removeChild(this.element);
    }
  }
}

/**
 * StatusMessage component for displaying status and error messages
 */
export class StatusMessage {
  /**
   * Creates a new status message instance
   * @param {Object} options - Status message configuration
   */
  constructor(options = {}) {
    this.id = options.id || `status-${Date.now()}`;
    this.type = options.type || 'info'; // info, success, warning, error
    this.message = options.message || '';
    this.dismissible = options.dismissible !== false;
    this.autoHide = options.autoHide || 0; // 0 = no auto hide
    this.className = options.className || 'status-message';
    
    this.element = this.createElement();
    this.autoHideTimer = null;
    
    if (this.autoHide > 0) {
      this.startAutoHide();
    }
  }

  /**
   * Creates the status message DOM element
   * @returns {HTMLElement} Status message element
   */
  createElement() {
    const container = DOMHelper.createElement('div', {
      attributes: { id: this.id },
      className: `${this.className} status-${this.type}`,
      styles: {
        padding: '12px 16px',
        margin: '10px 0',
        borderRadius: '4px',
        border: '1px solid',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        fontSize: '0.875em',
        ...this.getTypeStyles()
      }
    });

    // Message content
    const content = DOMHelper.createElement('div', {
      className: 'status-content',
      innerHTML: this.message
    });
    container.appendChild(content);

    // Dismiss button
    if (this.dismissible) {
      const dismissBtn = DOMHelper.createElement('button', {
        attributes: { type: 'button' },
        className: 'status-dismiss',
        styles: {
          background: 'none',
          border: 'none',
          fontSize: '16px',
          cursor: 'pointer',
          padding: '0 4px',
          opacity: '0.7'
        },
        textContent: '×'
      });

      this.dismissCleanup = DOMHelper.addEventListener(dismissBtn, 'click', () => {
        this.hide();
      });

      container.appendChild(dismissBtn);
    }

    return container;
  }

  /**
   * Gets styling based on message type
   * @returns {Object} Type-specific styles
   */
  getTypeStyles() {
    const typeStyles = {
      info: {
        backgroundColor: '#d1ecf1',
        borderColor: '#bee5eb',
        color: '#0c5460'
      },
      success: {
        backgroundColor: '#d4edda',
        borderColor: '#c3e6cb',
        color: '#155724'
      },
      warning: {
        backgroundColor: '#fff3cd',
        borderColor: '#ffeaa7',
        color: '#856404'
      },
      error: {
        backgroundColor: '#f8d7da',
        borderColor: '#f5c6cb',
        color: '#721c24'
      }
    };

    return typeStyles[this.type] || typeStyles.info;
  }

  /**
   * Updates the message content and type
   * @param {string} message - New message content
   * @param {string} type - New message type
   */
  update(message, type = this.type) {
    this.message = message;
    this.type = type;
    
    // Update content
    const contentElement = this.element.querySelector('.status-content');
    if (contentElement) {
      DOMHelper.updateContent(contentElement, message);
    }

    // Update styling
    Object.entries(this.getTypeStyles()).forEach(([property, value]) => {
      this.element.style[property] = value;
    });

    // Update class
    this.element.className = `${this.className} status-${this.type}`;

    // Restart auto-hide timer
    if (this.autoHide > 0) {
      this.startAutoHide();
    }
  }

  /**
   * Shows the status message
   */
  show() {
    DOMHelper.toggleElement(this.element, true, { duration: 200 });
  }

  /**
   * Hides the status message
   */
  hide() {
    DOMHelper.toggleElement(this.element, false, { duration: 200 });
    this.stopAutoHide();
  }

  /**
   * Starts auto-hide timer
   */
  startAutoHide() {
    this.stopAutoHide();
    this.autoHideTimer = setTimeout(() => {
      this.hide();
    }, this.autoHide);
  }

  /**
   * Stops auto-hide timer
   */
  stopAutoHide() {
    if (this.autoHideTimer) {
      clearTimeout(this.autoHideTimer);
      this.autoHideTimer = null;
    }
  }

  /**
   * Gets the status message element
   * @returns {HTMLElement} Status message element
   */
  getElement() {
    return this.element;
  }

  /**
   * Destroys the status message
   */
  destroy() {
    this.stopAutoHide();
    if (this.dismissCleanup) {
      this.dismissCleanup();
    }
    if (this.element && this.element.parentNode) {
      this.element.parentNode.removeChild(this.element);
    }
  }
}

/**
 * FileInput component with drag-and-drop support
 */
export class FileInput {
  /**
   * Creates a new file input instance
   * @param {Object} options - File input configuration
   */
  constructor(options = {}) {
    this.id = options.id || `file-input-${Date.now()}`;
    this.placeholder = options.placeholder || UI.PLACEHOLDERS.FILE_PATH;
    this.accept = options.accept || '.txz,.tar.xz';
    this.multiple = options.multiple || false;
    this.required = options.required || false;
    this.onFileSelect = options.onFileSelect || null;
    this.enableDragDrop = options.enableDragDrop !== false;
    
    this.element = this.createElement();
    this.setupEventListeners();
  }

  /**
   * Creates the file input DOM element
   * @returns {HTMLElement} File input element
   */
  createElement() {
    const container = DOMHelper.createElement('div', {
      attributes: { id: this.id },
      className: 'file-input-container',
      styles: {
        display: 'flex',
        gap: '10px',
        alignItems: 'center'
      }
    });

    // Text input for file path
    this.inputElement = DOMHelper.createElement('input', {
      attributes: {
        type: 'text',
        placeholder: this.placeholder,
        required: this.required
      },
      className: 'file-path-input',
      styles: {
        flex: '1',
        padding: '10px',
        fontSize: '14px',
        borderRadius: '4px',
        border: '1px solid #ccc',
        backgroundColor: COLORS.BACKGROUND.WHITE
      }
    });

    // Browse button
    this.browseButton = new Button({
      text: UI.BUTTONS.BROWSE,
      className: 'browse-button',
      onClick: () => this.openFileDialog()
    });

    container.appendChild(this.inputElement);
    container.appendChild(this.browseButton.getElement());

    // Hidden file input
    this.fileInputElement = DOMHelper.createElement('input', {
      attributes: {
        type: 'file',
        accept: this.accept,
        multiple: this.multiple,
        style: 'display: none;'
      }
    });

    container.appendChild(this.fileInputElement);

    return container;
  }

  /**
   * Sets up event listeners
   */
  setupEventListeners() {
    // File input change
    this.fileInputChangeCleanup = DOMHelper.addEventListener(
      this.fileInputElement, 
      'change', 
      (event) => this.handleFileSelect(event.target.files)
    );

    // Text input change
    this.inputChangeCleanup = DOMHelper.addEventListener(
      this.inputElement,
      'input',
      (event) => this.handleInputChange(event.target.value)
    );

    // Drag and drop
    if (this.enableDragDrop) {
      this.setupDragDrop();
    }
  }

  /**
   * Sets up drag-and-drop functionality
   */
  setupDragDrop() {
    const dragOverHandler = (event) => {
      event.preventDefault();
      event.stopPropagation();
      this.element.classList.add('drag-over');
    };

    const dragLeaveHandler = (event) => {
      event.preventDefault();
      event.stopPropagation();
      this.element.classList.remove('drag-over');
    };

    const dropHandler = (event) => {
      event.preventDefault();
      event.stopPropagation();
      this.element.classList.remove('drag-over');
      
      const files = event.dataTransfer.files;
      this.handleFileSelect(files);
    };

    this.dragOverCleanup = DOMHelper.addEventListener(this.element, 'dragover', dragOverHandler);
    this.dragLeaveCleanup = DOMHelper.addEventListener(this.element, 'dragleave', dragLeaveHandler);
    this.dropCleanup = DOMHelper.addEventListener(this.element, 'drop', dropHandler);
  }

  /**
   * Opens the file dialog
   */
  openFileDialog() {
    this.fileInputElement.click();
  }

  /**
   * Handles file selection
   * @param {FileList} files - Selected files
   */
  handleFileSelect(files) {
    if (files.length > 0) {
      const file = files[0]; // Take first file for single file mode
      const filePath = file.name; // Browser can't get full path for security
      
      this.inputElement.value = filePath;
      
      if (this.onFileSelect) {
        this.onFileSelect({
          file: file,
          name: file.name,
          size: file.size,
          type: file.type,
          lastModified: file.lastModified
        });
      }
    }
  }

  /**
   * Handles text input change
   * @param {string} value - Input value
   */
  handleInputChange(value) {
    if (this.onFileSelect && value.trim()) {
      this.onFileSelect({
        path: value.trim(),
        name: value.split('/').pop(),
        size: null,
        type: null,
        lastModified: null
      });
    }
  }

  /**
   * Gets the current file path
   * @returns {string} Current file path
   */
  getValue() {
    return this.inputElement.value;
  }

  /**
   * Sets the file path
   * @param {string} path - File path to set
   */
  setValue(path) {
    this.inputElement.value = path || '';
  }

  /**
   * Clears the input
   */
  clear() {
    this.inputElement.value = '';
    this.fileInputElement.value = '';
  }

  /**
   * Enables or disables the input
   * @param {boolean} disabled - Disabled state
   */
  setDisabled(disabled) {
    this.inputElement.disabled = disabled;
    this.browseButton.setDisabled(disabled);
  }

  /**
   * Gets the file input element
   * @returns {HTMLElement} File input element
   */
  getElement() {
    return this.element;
  }

  /**
   * Destroys the file input
   */
  destroy() {
    if (this.fileInputChangeCleanup) this.fileInputChangeCleanup();
    if (this.inputChangeCleanup) this.inputChangeCleanup();
    if (this.dragOverCleanup) this.dragOverCleanup();
    if (this.dragLeaveCleanup) this.dragLeaveCleanup();
    if (this.dropCleanup) this.dropCleanup();
    
    this.browseButton.destroy();
    
    if (this.element && this.element.parentNode) {
      this.element.parentNode.removeChild(this.element);
    }
  }
}

/**
 * Export all UI components for easy import
 */
export default {
  Button,
  ProgressBar,
  StatusMessage,
  FileInput
};
