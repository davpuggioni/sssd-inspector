/**
 * Tests for UI Components
 * @version 2.0.0
 */

// Mock DOM environment
document.body.innerHTML = '<div id="app"></div>';

// Import components after DOM setup
const { Button, ProgressBar, StatusMessage, FileInput } = require('../src/components/UIComponents.js');

describe('Button Component', () => {
    test('creates button with default properties', () => {
        const btn = new Button({ text: 'Test' });
        const el = btn.getElement();
        expect(el.tagName).toBe('BUTTON');
        expect(el.textContent).toBe('Test');
        expect(el.disabled).toBe(false);
    });

    test('creates button with custom properties', () => {
        const btn = new Button({
            text: 'Analyze',
            disabled: true,
            className: 'btn btn-primary'
        });
        const el = btn.getElement();
        expect(el.textContent).toBe('Analyze');
        expect(el.disabled).toBe(true);
        expect(el.className).toBe('btn btn-primary');
    });

    test('updates button text', () => {
        const btn = new Button({ text: 'Start' });
        btn.setText('Stop');
        expect(btn.getElement().textContent).toBe('Stop');
    });

    test('toggles disabled state', () => {
        const btn = new Button({ text: 'Test' });
        btn.setDisabled(true);
        expect(btn.getElement().disabled).toBe(true);
        btn.setDisabled(false);
        expect(btn.getElement().disabled).toBe(false);
    });

    test('click handler fires', () => {
        let clicked = false;
        const btn = new Button({ text: 'Test', onClick: () => { clicked = true; } });
        btn.getElement().click();
        expect(clicked).toBe(true);
    });

    test('destroy cleans up', () => {
        const container = document.createElement('div');
        const btn = new Button({ text: 'Test' });
        container.appendChild(btn.getElement());
        btn.destroy();
        expect(container.children.length).toBe(0);
    });
});

describe('ProgressBar Component', () => {
    test('creates progress bar with default values', () => {
        const pb = new ProgressBar();
        const el = pb.getElement();
        expect(el).toBeDefined();
        expect(el.style.display).toBe('none'); // 0% = hidden
    });

    test('updates progress value', () => {
        const pb = new ProgressBar({ showPercentage: true });
        pb.setValue(50, 'Processing...');
        expect(pb.value).toBe(50);
        expect(pb.fillElement.style.width).toBe('50%');
    });

    test('clamps value to valid range', () => {
        const pb = new ProgressBar();
        pb.setValue(-10);
        expect(pb.value).toBe(0);
        pb.setValue(150);
        expect(pb.value).toBe(100);
    });

    test('shows when value > 0', () => {
        const pb = new ProgressBar();
        pb.setValue(0);
        expect(pb.element.style.display).toBe('none');
        pb.setValue(50);
        expect(pb.element.style.display).toBe('block');
    });

    test('updates status text', () => {
        const pb = new ProgressBar({ showStatus: true });
        pb.setValue(25, 'Scanning...');
        expect(pb.statusElement.textContent).toBe('Scanning...');
    });

    test('resets to initial state', () => {
        const pb = new ProgressBar({ showPercentage: true, showStatus: true });
        pb.setValue(75, 'Almost done');
        pb.reset();
        expect(pb.value).toBe(0);
        expect(pb.fillElement.style.width).toBe('0%');
        expect(pb.status).toBe('');
    });

    test('toggle shows/hides', () => {
        const pb = new ProgressBar();
        pb.toggle(false);
        expect(pb.element.style.display).toBe('none');
        pb.toggle(true);
        expect(pb.element.style.display).toBe('block');
    });

    test('destroy cleans up', () => {
        const container = document.createElement('div');
        const pb = new ProgressBar();
        container.appendChild(pb.getElement());
        pb.destroy();
        expect(container.children.length).toBe(0);
    });
});

describe('StatusMessage Component', () => {
    test('creates status message with defaults', () => {
        const sm = new StatusMessage({ message: 'Test message' });
        const el = sm.getElement();
        expect(el).toBeDefined();
        expect(el.textContent).toContain('Test message');
    });

    test('creates different types', () => {
        const types = ['info', 'success', 'warning', 'error'];
        types.forEach(type => {
            const sm = new StatusMessage({ message: 'Test', type });
            expect(sm.element.className).toContain(`status-${type}`);
        });
    });

    test('update changes message and type', () => {
        const sm = new StatusMessage({ message: 'Old', type: 'info' });
        sm.update('New message', 'success');
        expect(sm.message).toBe('New message');
        expect(sm.type).toBe('success');
    });

    test('show and hide work', () => {
        const sm = new StatusMessage({ message: 'Test' });
        sm.hide();
        expect(sm.element.style.display).toBe('none');
    });

    test('auto-hide timer works', (done) => {
        const sm = new StatusMessage({ message: 'Test', autoHide: 100 });
        expect(sm.autoHideTimer).toBeDefined();
        setTimeout(() => {
            expect(sm.element.style.display).toBe('none');
            done();
        }, 200);
    });

    test('destroy cleans up timers', () => {
        const sm = new StatusMessage({ message: 'Test', autoHide: 5000 });
        sm.destroy();
        expect(sm.autoHideTimer).toBeNull();
    });
});

describe('FileInput Component', () => {
    test('creates file input with defaults', () => {
        const fi = new FileInput();
        const el = fi.getElement();
        expect(el).toBeDefined();
        expect(fi.inputElement.placeholder).toBeDefined();
    });

    test('getValue and setValue work', () => {
        const fi = new FileInput();
        fi.setValue('/path/to/file.txz');
        expect(fi.getValue()).toBe('/path/to/file.txz');
    });

    test('clear resets input', () => {
        const fi = new FileInput();
        fi.setValue('/path/to/file.txz');
        fi.clear();
        expect(fi.getValue()).toBe('');
    });

    test('setDisabled toggles state', () => {
        const fi = new FileInput();
        fi.setDisabled(true);
        expect(fi.inputElement.disabled).toBe(true);
        expect(fi.browseButton.disabled).toBe(true);
        fi.setDisabled(false);
        expect(fi.inputElement.disabled).toBe(false);
        expect(fi.browseButton.disabled).toBe(false);
    });

    test('destroy cleans up', () => {
        const container = document.createElement('div');
        const fi = new FileInput();
        container.appendChild(fi.getElement());
        fi.destroy();
        expect(container.children.length).toBe(0);
    });
});