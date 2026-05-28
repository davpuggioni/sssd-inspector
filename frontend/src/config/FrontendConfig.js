/**
 * Frontend Configuration Manager for SSSD Inspector
 * Dynamic configuration system with validation and defaults
 * @version 2.0.0
 */

import { DEBUG, UI, COLORS, ZOOM, PROGRESS } from './constants.js';

/**
 * FrontendConfig class manages all frontend configuration
 * Provides validation, defaults, and dynamic updates
 */
export class FrontendConfig {
  /**
   * Creates a new FrontendConfig instance
   * @param {Object} options - Configuration options
   */
  constructor(options = {}) {
    this.config = this.mergeWithDefaults(options);
    this.validators = this.createValidators();
    this.observers = new Set();
    this.validate();
  }

  /**
   * Default configuration values
   * @returns {Object} Default configuration object
   */
  static getDefaults() {
    return {
      // Application settings
      app: {
        name: UI.APP_TITLE,
        version: UI.APP_VERSION,
        debug: DEBUG.ENABLED,
        logLevel: DEBUG.CURRENT_LOG_LEVEL,
      },

      // UI settings
      ui: {
        theme: 'light',
        language: 'en',
        fontSize: 'medium',
        animations: true,
        tooltips: true,
        autoSave: true,
      },

      // Analysis settings
      analysis: {
        autoStart: false,
        showProgress: true,
        progressUpdateInterval: PROGRESS.UPDATE_INTERVAL,
        timeout: 300000, // 5 minutes in milliseconds
        maxFileSize: 100 * 1024 * 1024, // 100MB
        supportedFormats: ['.txz', '.tar.xz'],
      },

      // Report settings
      report: {
        defaultFormat: 'html',
        showTimestamps: true,
        showEmptySections: false,
        expandDetails: false,
        zoomLevel: ZOOM.DEFAULT,
        printOptimized: true,
      },

      // Export settings
      export: {
        defaultLocation: 'downloads',
        includeMetadata: true,
        anonymizeByDefault: false,
        compression: false,
      },

      // Performance settings
      performance: {
        debounceDelay: 300,
        maxConcurrentRequests: 3,
        cacheResults: true,
        cacheSize: 50,
        lazyLoading: true,
      },

      // Accessibility settings
      accessibility: {
        highContrast: false,
        reducedMotion: false,
        screenReader: false,
        keyboardNavigation: true,
        focusVisible: true,
      },

      // Notification settings
      notifications: {
        enabled: true,
        duration: 5000,
        position: 'top-right',
        showProgress: true,
        sound: false,
      },

      // Storage settings
      storage: {
        prefix: 'sssd-inspector-',
        encrypt: false,
        compression: false,
        maxStorageSize: 10 * 1024 * 1024, // 10MB
      },
    };
  }

  /**
   * Merges user options with defaults
   * @param {Object} options - User configuration options
   * @returns {Object} Merged configuration
   */
  mergeWithDefaults(options) {
    const defaults = FrontendConfig.getDefaults();
    return this.deepMerge(defaults, options);
  }

  /**
   * Deep merge two objects
   * @param {Object} target - Target object
   * @param {Object} source - Source object
   * @returns {Object} Merged object
   */
  deepMerge(target, source) {
    const result = { ...target };
    
    for (const key in source) {
      if (source.hasOwnProperty(key)) {
        if (this.isObject(source[key]) && this.isObject(target[key])) {
          result[key] = this.deepMerge(target[key], source[key]);
        } else {
          result[key] = source[key];
        }
      }
    }
    
    return result;
  }

  /**
   * Checks if a value is an object
   * @param {any} value - Value to check
   * @returns {boolean} Whether the value is an object
   */
  isObject(value) {
    return value !== null && typeof value === 'object' && !Array.isArray(value);
  }

  /**
   * Creates configuration validators
   * @returns {Object} Validator functions
   */
  createValidators() {
    return {
      app: {
        name: (value) => typeof value === 'string' && value.length > 0,
        version: (value) => typeof value === 'string' && /^\d+\.\d+\.\d+$/.test(value),
        debug: (value) => typeof value === 'boolean',
        logLevel: (value) => [0, 1, 2, 3].includes(value),
      },
      ui: {
        theme: (value) => ['light', 'dark', 'auto'].includes(value),
        language: (value) => typeof value === 'string' && value.length === 2,
        fontSize: (value) => ['small', 'medium', 'large', 'extra-large'].includes(value),
        animations: (value) => typeof value === 'boolean',
        tooltips: (value) => typeof value === 'boolean',
        autoSave: (value) => typeof value === 'boolean',
      },
      analysis: {
        autoStart: (value) => typeof value === 'boolean',
        showProgress: (value) => typeof value === 'boolean',
        progressUpdateInterval: (value) => typeof value === 'number' && value > 0,
        timeout: (value) => typeof value === 'number' && value > 0,
        maxFileSize: (value) => typeof value === 'number' && value > 0,
        supportedFormats: (value) => Array.isArray(value) && value.every(f => typeof f === 'string'),
      },
      report: {
        defaultFormat: (value) => ['html', 'text', 'pdf'].includes(value),
        showTimestamps: (value) => typeof value === 'boolean',
        showEmptySections: (value) => typeof value === 'boolean',
        expandDetails: (value) => typeof value === 'boolean',
        zoomLevel: (value) => typeof value === 'number' && value >= ZOOM.MIN && value <= ZOOM.MAX,
        printOptimized: (value) => typeof value === 'boolean',
      },
      export: {
        defaultLocation: (value) => ['downloads', 'desktop', 'documents', 'custom'].includes(value),
        includeMetadata: (value) => typeof value === 'boolean',
        anonymizeByDefault: (value) => typeof value === 'boolean',
        compression: (value) => typeof value === 'boolean',
      },
      performance: {
        debounceDelay: (value) => typeof value === 'number' && value >= 0,
        maxConcurrentRequests: (value) => typeof value === 'number' && value > 0,
        cacheResults: (value) => typeof value === 'boolean',
        cacheSize: (value) => typeof value === 'number' && value > 0,
        lazyLoading: (value) => typeof value === 'boolean',
      },
      accessibility: {
        highContrast: (value) => typeof value === 'boolean',
        reducedMotion: (value) => typeof value === 'boolean',
        screenReader: (value) => typeof value === 'boolean',
        keyboardNavigation: (value) => typeof value === 'boolean',
        focusVisible: (value) => typeof value === 'boolean',
      },
      notifications: {
        enabled: (value) => typeof value === 'boolean',
        duration: (value) => typeof value === 'number' && value > 0,
        position: (value) => ['top-left', 'top-right', 'bottom-left', 'bottom-right'].includes(value),
        showProgress: (value) => typeof value === 'boolean',
        sound: (value) => typeof value === 'boolean',
      },
      storage: {
        prefix: (value) => typeof value === 'string' && value.length > 0,
        encrypt: (value) => typeof value === 'boolean',
        compression: (value) => typeof value === 'boolean',
        maxStorageSize: (value) => typeof value === 'number' && value > 0,
      },
    };
  }

  /**
   * Validates the current configuration
   * @throws {Error} If validation fails
   */
  validate() {
    const errors = this.validateConfig(this.config, '', this.validators);
    
    if (errors.length > 0) {
      throw new Error(`Configuration validation failed:\n${errors.join('\n')}`);
    }
  }

  /**
   * Validates a configuration object
   * @param {Object} config - Configuration to validate
   * @param {string} path - Current path in the configuration
   * @param {Object} validators - Validator functions
   * @returns {Array<string>} Array of error messages
   */
  validateConfig(config, path, validators) {
    const errors = [];
    
    for (const key in config) {
      if (config.hasOwnProperty(key)) {
        const currentPath = path ? `${path}.${key}` : key;
        const value = config[key];
        
        if (validators[key]) {
          // Validate primitive values
          if (typeof value !== 'object' || Array.isArray(value)) {
            if (!validators[key](value)) {
              errors.push(`Invalid value for ${currentPath}: ${JSON.stringify(value)}`);
            }
          }
        } else if (this.isObject(value) && validators[key]) {
          // Validate nested objects
          const nestedErrors = this.validateConfig(value, currentPath, validators[key]);
          errors.push(...nestedErrors);
        }
      }
    }
    
    return errors;
  }

  /**
   * Gets a configuration value by path
   * @param {string} path - Dot-separated path to the value
   * @param {any} defaultValue - Default value if path doesn't exist
   * @returns {any} Configuration value
   */
  get(path, defaultValue = undefined) {
    return this.getNestedValue(this.config, path, defaultValue);
  }

  /**
   * Sets a configuration value by path
   * @param {string} path - Dot-separated path to the value
   * @param {any} value - Value to set
   * @param {boolean} validate - Whether to validate the new value
   */
  set(path, value, validate = true) {
    const oldValue = this.get(path);
    this.setNestedValue(this.config, path, value);
    
    if (validate) {
      try {
        this.validate();
      } catch (error) {
        // Revert on validation error
        this.setNestedValue(this.config, path, oldValue);
        throw error;
      }
    }
    
    this.notifyObservers(path, value, oldValue);
  }

  /**
   * Gets a nested value from an object
   * @param {Object} obj - Object to get value from
   * @param {string} path - Dot-separated path
   * @param {any} defaultValue - Default value
   * @returns {any} Nested value
   */
  getNestedValue(obj, path, defaultValue) {
    const keys = path.split('.');
    let current = obj;
    
    for (const key of keys) {
      if (current === null || current === undefined || !current.hasOwnProperty(key)) {
        return defaultValue;
      }
      current = current[key];
    }
    
    return current;
  }

  /**
   * Sets a nested value in an object
   * @param {Object} obj - Object to set value in
   * @param {string} path - Dot-separated path
   * @param {any} value - Value to set
   */
  setNestedValue(obj, path, value) {
    const keys = path.split('.');
    let current = obj;
    
    for (let i = 0; i < keys.length - 1; i++) {
      const key = keys[i];
      
      if (!current.hasOwnProperty(key) || !this.isObject(current[key])) {
        current[key] = {};
      }
      
      current = current[key];
    }
    
    current[keys[keys.length - 1]] = value;
  }

  /**
   * Updates multiple configuration values
   * @param {Object} updates - Object with updates
   * @param {boolean} validate - Whether to validate the updates
   */
  update(updates, validate = true) {
    const oldConfig = this.deepMerge({}, this.config);
    
    try {
      this.config = this.deepMerge(this.config, updates);
      
      if (validate) {
        this.validate();
      }
      
      // Notify observers of all changes
      this.notifyObservers('*', this.config, oldConfig);
    } catch (error) {
      // Revert on validation error
      this.config = oldConfig;
      throw error;
    }
  }

  /**
   * Resets configuration to defaults
   * @param {string} path - Optional path to reset (resets all if not provided)
   */
  reset(path = null) {
    if (path) {
      const defaultValue = this.getNestedValue(FrontendConfig.getDefaults(), path);
      this.set(path, defaultValue);
    } else {
      const oldConfig = this.config;
      this.config = FrontendConfig.getDefaults();
      this.notifyObservers('*', this.config, oldConfig);
    }
  }

  /**
   * Adds an observer for configuration changes
   * @param {Function} observer - Observer function
   * @returns {Function} Function to remove the observer
   */
  addObserver(observer) {
    this.observers.add(observer);
    
    return () => {
      this.observers.delete(observer);
    };
  }

  /**
   * Notifies all observers of a configuration change
   * @param {string} path - Path that changed
   * @param {any} newValue - New value
   * @param {any} oldValue - Old value
   */
  notifyObservers(path, newValue, oldValue) {
    for (const observer of this.observers) {
      try {
        observer(path, newValue, oldValue);
      } catch (error) {
        console.warn('Configuration observer error:', error);
      }
    }
  }

  /**
   * Exports configuration to JSON
   * @returns {string} JSON string of configuration
   */
  export() {
    return JSON.stringify(this.config, null, 2);
  }

  /**
   * Imports configuration from JSON
   * @param {string} json - JSON string to import
   * @param {boolean} validate - Whether to validate the imported config
   */
  import(json, validate = true) {
    try {
      const config = JSON.parse(json);
      this.update(config, validate);
    } catch (error) {
      throw new Error(`Failed to import configuration: ${error.message}`);
    }
  }

  /**
   * Saves configuration to localStorage
   * @param {string} key - Storage key (optional)
   * @returns {boolean} Success status
   */
  save(key = 'frontend-config') {
    try {
      const storageKey = this.get('storage.prefix', 'sssd-inspector-') + key;
      const json = this.export();
      localStorage.setItem(storageKey, json);
      return true;
    } catch (error) {
      console.warn('Failed to save configuration:', error);
      return false;
    }
  }

  /**
   * Loads configuration from localStorage
   * @param {string} key - Storage key (optional)
   * @returns {boolean} Success status
   */
  load(key = 'frontend-config') {
    try {
      const storageKey = this.get('storage.prefix', 'sssd-inspector-') + key;
      const json = localStorage.getItem(storageKey);
      
      if (json) {
        this.import(json);
        return true;
      }
      
      return false;
    } catch (error) {
      console.warn('Failed to load configuration:', error);
      return false;
    }
  }

  /**
   * Gets the complete configuration object
   * @returns {Object} Configuration object
   */
  getAll() {
    return this.deepMerge({}, this.config);
  }

  /**
   * Creates a copy of the configuration
   * @returns {FrontendConfig} New configuration instance
   */
  clone() {
    return new FrontendConfig(this.getAll());
  }

  /**
   * Applies system preferences (dark mode, reduced motion, etc.)
   */
  applySystemPreferences() {
    // Check for dark mode preference
    if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
      this.set('ui.theme', 'dark');
    }

    // Check for reduced motion preference
    if (window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      this.set('ui.animations', false);
      this.set('accessibility.reducedMotion', true);
    }

    // Check for high contrast preference
    if (window.matchMedia && window.matchMedia('(prefers-contrast: high)').matches) {
      this.set('accessibility.highContrast', true);
    }
  }

  /**
   * Destroys the configuration instance
   */
  destroy() {
    this.observers.clear();
  }
}

/**
 * Global configuration instance
 */
let globalConfig = null;

/**
 * Gets the global configuration instance
 * @param {Object} options - Configuration options
 * @returns {FrontendConfig} Configuration instance
 */
export function getConfig(options = {}) {
  if (!globalConfig) {
    globalConfig = new FrontendConfig(options);
    globalConfig.applySystemPreferences();
  }
  return globalConfig;
}

/**
 * Sets the global configuration instance
 * @param {FrontendConfig} config - Configuration instance
 */
export function setConfig(config) {
  if (globalConfig) {
    globalConfig.destroy();
  }
  globalConfig = config;
}

/**
 * Resets the global configuration
 */
export function resetConfig() {
  if (globalConfig) {
    globalConfig.destroy();
    globalConfig = null;
  }
}

export default FrontendConfig;
