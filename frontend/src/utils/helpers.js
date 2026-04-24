/**
 * Utility Helper Functions for SSSD Inspector Frontend
 * Provides common utility functions and formatting helpers
 * @version 2.0.0
 */

import { COLORS, UI, CSS_CLASSES } from '../config/constants.js';

/**
 * FormatHelper class for data formatting and display utilities
 */
export class FormatHelper {
  /**
   * Formats boolean values as styled HTML spans
   * @param {boolean} value - Boolean value to format
   * @param {Object} options - Formatting options
   * @returns {string} Formatted HTML string
   */
  static formatBoolean(value, options = {}) {
    const { 
      trueText = 'Yes', 
      falseText = 'No',
      trueClass = CSS_CLASSES.STATES.SUCCESS,
      falseClass = CSS_CLASSES.STATES.FAIL 
    } = options;

    const cssClass = value ? trueClass : falseClass;
    const text = value ? trueText : falseText;
    
    return `<span class="${cssClass}">${text}</span>`;
  }

  /**
   * Formats arrays as HTML list items
   * @param {Array} array - Array to format
   * @param {Object} options - Formatting options
   * @returns {string} Formatted HTML string
   */
  static formatArray(array, options = {}) {
    const { 
      emptyMessage = 'None',
      itemClass = '',
      listClass = '',
      separator = '<br>'
    } = options;

    if (!array || !Array.isArray(array) || array.length === 0) {
      return emptyMessage;
    }

    const itemClassAttr = itemClass ? ` class="${itemClass}"` : '';
    const listClassAttr = listClass ? ` class="${listClass}"` : '';

    if (separator === '<br>') {
      return array.map(item => `<span${itemClassAttr}>${item}</span>`).join(separator);
    }

    const items = array.map(item => `<li${itemClassAttr}>${item}</li>`).join('');
    return `<ul${listClassAttr}>${items}</ul>`;
  }

  /**
   * Formats timestamps for display
   * @param {string} timestamp - ISO timestamp string
   * @param {Object} options - Formatting options
   * @returns {string} Formatted timestamp
   */
  static formatTimestamp(timestamp, options = {}) {
    const { 
      includeTime = true,
      dateFormat = 'medium',
      locale = 'en-US'
    } = options;

    if (!timestamp) return 'Unknown';

    try {
      const date = new Date(timestamp);
      
      if (includeTime) {
        return date.toLocaleString(locale, {
          year: 'numeric',
          month: dateFormat,
          day: 'numeric',
          hour: '2-digit',
          minute: '2-digit'
        });
      } else {
        return date.toLocaleDateString(locale, {
          year: 'numeric',
          month: dateFormat,
          day: 'numeric'
        });
      }
    } catch (error) {
      console.warn('Invalid timestamp format:', timestamp);
      return timestamp;
    }
  }

  /**
   * Formats file sizes in human-readable format
   * @param {number} bytes - Size in bytes
   * @param {Object} options - Formatting options
   * @returns {string} Formatted file size
   */
  static formatFileSize(bytes, options = {}) {
    const { decimals = 2, locale = 'en-US' } = options;

    if (bytes === 0) return '0 Bytes';

    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];

    const i = Math.floor(Math.log(bytes) / Math.log(k));

    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
  }

  /**
   * Truncates text with ellipsis
   * @param {string} text - Text to truncate
   * @param {number} maxLength - Maximum length
   * @param {Object} options - Truncation options
   * @returns {string} Truncated text
   */
  static truncateText(text, maxLength, options = {}) {
    const { suffix = '...', wordBoundary = true } = options;

    if (!text || text.length <= maxLength) {
      return text || '';
    }

    if (wordBoundary) {
      const truncated = text.substring(0, maxLength - suffix.length);
      const lastSpaceIndex = truncated.lastIndexOf(' ');
      
      if (lastSpaceIndex > 0) {
        return truncated.substring(0, lastSpaceIndex) + suffix;
      }
    }

    return text.substring(0, maxLength - suffix.length) + suffix;
  }

  /**
   * Escapes HTML special characters
   * @param {string} text - Text to escape
   * @returns {string} Escaped text
   */
  static escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text || '';
    return div.innerHTML;
  }

  /**
   * Highlights search terms in text
   * @param {string} text - Text to search in
   * @param {string} searchTerm - Term to highlight
   * @param {Object} options - Highlighting options
   * @returns {string} Text with highlighted terms
   */
  static highlightSearch(text, searchTerm, options = {}) {
    const { 
      highlightClass = 'highlight',
      caseSensitive = false 
    } = options;

    if (!searchTerm || !text) {
      return text || '';
    }

    const flags = caseSensitive ? 'g' : 'gi';
    const regex = new RegExp(`(${searchTerm})`, flags);
    
    return text.replace(regex, `<span class="${highlightClass}">$1</span>`);
  }
}

/**
 * DOMHelper class for DOM manipulation utilities
 */
export class DOMHelper {
  /**
   * Creates an HTML element with attributes and content
   * @param {string} tagName - HTML tag name
   * @param {Object} options - Element options
   * @returns {HTMLElement} Created element
   */
  static createElement(tagName, options = {}) {
    const element = document.createElement(tagName);
    
    // Set attributes
    if (options.attributes) {
      Object.entries(options.attributes).forEach(([key, value]) => {
        element.setAttribute(key, value);
      });
    }

    // Set CSS classes
    if (options.className) {
      element.className = options.className;
    }

    // Set CSS styles
    if (options.styles) {
      Object.entries(options.styles).forEach(([property, value]) => {
        element.style[property] = value;
      });
    }

    // Set text content
    if (options.textContent) {
      element.textContent = options.textContent;
    }

    // Set HTML content
    if (options.innerHTML) {
      element.innerHTML = options.innerHTML;
    }

    // Append child elements
    if (options.children) {
      options.children.forEach(child => {
        element.appendChild(child);
      });
    }

    return element;
  }

  /**
   * Finds an element by selector with error handling
   * @param {string} selector - CSS selector
   * @param {HTMLElement} parent - Parent element to search in (optional)
   * @returns {HTMLElement|null} Found element or null
   */
  static findElement(selector, parent = document) {
    try {
      return parent.querySelector(selector);
    } catch (error) {
      console.warn('Invalid selector:', selector);
      return null;
    }
  }

  /**
   * Finds multiple elements by selector
   * @param {string} selector - CSS selector
   * @param {HTMLElement} parent - Parent element to search in (optional)
   * @returns {NodeList} Found elements
   */
  static findElements(selector, parent = document) {
    try {
      return parent.querySelectorAll(selector);
    } catch (error) {
      console.warn('Invalid selector:', selector);
      return [];
    }
  }

  /**
   * Adds event listeners with cleanup support
   * @param {HTMLElement} element - Target element
   * @param {string} event - Event name
   * @param {Function} handler - Event handler
   * @param {Object} options - Event options
   * @returns {Function} Cleanup function
   */
  static addEventListener(element, event, handler, options = {}) {
    element.addEventListener(event, handler, options);
    
    return () => {
      element.removeEventListener(event, handler, options);
    };
  }

  /**
   * Shows or hides an element with animation
   * @param {HTMLElement} element - Element to toggle
   * @param {boolean} show - Whether to show or hide
   * @param {Object} options - Animation options
   */
  static toggleElement(element, show, options = {}) {
    const { 
      display = 'block',
      duration = 300,
      opacity = true 
    } = options;

    if (!element) return;

    if (show) {
      element.style.display = display;
      if (opacity) {
        element.style.opacity = '0';
        element.style.transition = `opacity ${duration}ms ease-in-out`;
        
        // Force reflow
        element.offsetHeight;
        
        element.style.opacity = '1';
      }
    } else {
      if (opacity) {
        element.style.opacity = '0';
        setTimeout(() => {
          element.style.display = 'none';
        }, duration);
      } else {
        element.style.display = 'none';
      }
    }
  }

  /**
   * Updates element content safely
   * @param {HTMLElement} element - Target element
   * @param {string} content - New content
   * @param {boolean} escape - Whether to escape HTML
   */
  static updateContent(element, content, escape = false) {
    if (!element) return;

    if (escape) {
      element.textContent = content || '';
    } else {
      element.innerHTML = content || '';
    }
  }
}

/**
 * StorageHelper class for browser storage utilities
 */
export class StorageHelper {
  /**
   * Saves data to localStorage with error handling
   * @param {string} key - Storage key
   * @param {any} value - Value to store
   * @returns {boolean} Success status
   */
  static saveToLocalStorage(key, value) {
    try {
      const serializedValue = JSON.stringify(value);
      localStorage.setItem(key, serializedValue);
      return true;
    } catch (error) {
      console.warn('Failed to save to localStorage:', error);
      return false;
    }
  }

  /**
   * Retrieves data from localStorage with error handling
   * @param {string} key - Storage key
   * @param {any} defaultValue - Default value if key doesn't exist
   * @returns {any} Retrieved value or default
   */
  static getFromLocalStorage(key, defaultValue = null) {
    try {
      const serializedValue = localStorage.getItem(key);
      if (serializedValue === null) {
        return defaultValue;
      }
      return JSON.parse(serializedValue);
    } catch (error) {
      console.warn('Failed to retrieve from localStorage:', error);
      return defaultValue;
    }
  }

  /**
   * Removes data from localStorage
   * @param {string} key - Storage key
   * @returns {boolean} Success status
   */
  static removeFromLocalStorage(key) {
    try {
      localStorage.removeItem(key);
      return true;
    } catch (error) {
      console.warn('Failed to remove from localStorage:', error);
      return false;
    }
  }

  /**
   * Clears all localStorage data for the app
   * @param {string} appPrefix - Prefix for app-specific keys
   */
  static clearAppStorage(appPrefix = 'sssd-inspector-') {
    try {
      const keysToRemove = [];
      for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i);
        if (key && key.startsWith(appPrefix)) {
          keysToRemove.push(key);
        }
      }
      
      keysToRemove.forEach(key => localStorage.removeItem(key));
      return true;
    } catch (error) {
      console.warn('Failed to clear app storage:', error);
      return false;
    }
  }
}

/**
 * URLHelper class for URL manipulation utilities
 */
export class URLHelper {
  /**
   * Gets URL parameters as an object
   * @param {string} url - URL to parse (optional, defaults to current URL)
   * @returns {Object} URL parameters
   */
  static getUrlParameters(url = window.location.href) {
    const urlObj = new URL(url);
    const params = {};
    
    urlObj.searchParams.forEach((value, key) => {
      params[key] = value;
    });
    
    return params;
  }

  /**
   * Updates URL parameters
   * @param {Object} params - Parameters to update
   * @param {boolean} replaceHistory - Whether to replace history state
   */
  static updateUrlParameters(params, replaceHistory = false) {
    const url = new URL(window.location.href);
    
    Object.entries(params).forEach(([key, value]) => {
      if (value === null || value === undefined) {
        url.searchParams.delete(key);
      } else {
        url.searchParams.set(key, value);
      }
    });
    
    const method = replaceHistory ? 'replaceState' : 'pushState';
    window.history[method]({}, '', url.toString());
  }

  /**
   * Validates if a URL is properly formatted
   * @param {string} url - URL to validate
   * @returns {boolean} Validation result
   */
  static isValidUrl(url) {
    try {
      new URL(url);
      return true;
    } catch (error) {
      return false;
    }
  }
}

/**
 * Export all helper utilities for easy import
 */
export default {
  FormatHelper,
  DOMHelper,
  StorageHelper,
  URLHelper
};
