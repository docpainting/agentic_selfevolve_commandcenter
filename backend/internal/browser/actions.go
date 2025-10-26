package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

// Action represents a browser action
type Action struct {
	Type       string
	Parameters map[string]interface{}
	Timestamp  time.Time
	Result     interface{}
	Error      error
}

// ActionHistory tracks browser actions
type ActionHistory struct {
	actions []Action
}

// NewActionHistory creates a new action history
func NewActionHistory() *ActionHistory {
	return &ActionHistory{
		actions: make([]Action, 0),
	}
}

// ClickBySelector clicks an element by CSS selector
func (m *Manager) ClickBySelector(selector string) error {
	if err := m.ensureInitialized(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.WaitVisible(selector),
		chromedp.Click(selector),
	)
}

// TypeBySelector types text into an element by CSS selector
func (m *Manager) TypeBySelector(selector, text string) error {
	if err := m.ensureInitialized(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.WaitVisible(selector),
		chromedp.SendKeys(selector, text),
	)
}

// ClickAndType clicks an element and types text
func (m *Manager) ClickAndType(elementID int, text string) error {
	if err := m.Click(elementID); err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)
	return m.Type(elementID, text)
}

// SelectOption selects an option from a dropdown
func (m *Manager) SelectOption(selector, value string) error {
	if err := m.ensureInitialized(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.WaitVisible(selector),
		chromedp.SetValue(selector, value),
	)
}

// CheckCheckbox checks a checkbox
func (m *Manager) CheckCheckbox(selector string) error {
	if err := m.ensureInitialized(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.WaitVisible(selector),
		chromedp.SetAttributeValue(selector, "checked", "true"),
	)
}

// UncheckCheckbox unchecks a checkbox
func (m *Manager) UncheckCheckbox(selector string) error {
	if err := m.ensureInitialized(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.WaitVisible(selector),
		chromedp.RemoveAttribute(selector, "checked"),
	)
}

// GetText gets text from an element
func (m *Manager) GetText(selector string) (string, error) {
	if err := m.ensureInitialized(); err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
	defer cancel()

	var text string
	err := chromedp.Run(ctx,
		chromedp.WaitVisible(selector),
		chromedp.Text(selector, &text),
	)

	return text, err
}

// GetAttribute gets an attribute from an element
func (m *Manager) GetAttribute(selector, attribute string) (string, error) {
	if err := m.ensureInitialized(); err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
	defer cancel()

	var value string
	err := chromedp.Run(ctx,
		chromedp.WaitVisible(selector),
		chromedp.AttributeValue(selector, attribute, &value, nil),
	)

	return value, err
}

// Hover hovers over an element
func (m *Manager) Hover(elementID int) error {
	if err := m.ensureInitialized(); err != nil {
		return err
	}

	m.mu.RLock()
	if elementID < 0 || elementID >= len(m.elements) {
		m.mu.RUnlock()
		return fmt.Errorf("invalid element ID: %d", elementID)
	}

	element := m.elements[elementID]
	m.mu.RUnlock()

	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.MouseClickXY(element.X+element.Width/2, element.Y+element.Height/2, chromedp.ButtonNone),
	)
}

// DoubleClick double-clicks an element
func (m *Manager) DoubleClick(elementID int) error {
	if err := m.Click(elementID); err != nil {
		return err
	}

	time.Sleep(50 * time.Millisecond)
	return m.Click(elementID)
}

// RightClick right-clicks an element
func (m *Manager) RightClick(elementID int) error {
	if err := m.ensureInitialized(); err != nil {
		return err
	}

	m.mu.RLock()
	if elementID < 0 || elementID >= len(m.elements) {
		m.mu.RUnlock()
		return fmt.Errorf("invalid element ID: %d", elementID)
	}

	element := m.elements[elementID]
	m.mu.RUnlock()

	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.MouseClickXY(element.X+element.Width/2, element.Y+element.Height/2, chromedp.ButtonRight),
	)
}

// ScrollToElement scrolls to an element
func (m *Manager) ScrollToElement(elementID int) error {
	if err := m.ensureInitialized(); err != nil {
		return err
	}

	m.mu.RLock()
	if elementID < 0 || elementID >= len(m.elements) {
		m.mu.RUnlock()
		return fmt.Errorf("invalid element ID: %d", elementID)
	}

	element := m.elements[elementID]
	m.mu.RUnlock()

	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	defer cancel()

	script := fmt.Sprintf("window.scrollTo(%f, %f)", element.X, element.Y-100)
	return chromedp.Run(ctx,
		chromedp.Evaluate(script, nil),
	)
}

// WaitForNavigation waits for navigation to complete
func (m *Manager) WaitForNavigation(timeout time.Duration) error {
	if err := m.ensureInitialized(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(m.ctx, timeout)
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.WaitReady("body"),
	)
}

// GoBack navigates back
func (m *Manager) GoBack() error {
	if err := m.ensureInitialized(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.NavigateBack(),
	)
}

// GoForward navigates forward
func (m *Manager) GoForward() error {
	if err := m.ensureInitialized(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.NavigateForward(),
	)
}

// Reload reloads the page
func (m *Manager) Reload() error {
	if err := m.ensureInitialized(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.Reload(),
	)
}

// SetViewport sets the viewport size
func (m *Manager) SetViewport(width, height int64) error {
	if err := m.ensureInitialized(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.EmulateViewport(width, height),
	)
}

// GetCookies gets all cookies
func (m *Manager) GetCookies() ([]interface{}, error) {
	if err := m.ensureInitialized(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	defer cancel()

	var cookies []interface{}
	err := chromedp.Run(ctx,
		chromedp.Evaluate("document.cookie", &cookies),
	)

	return cookies, err
}

// ClearCookies clears all cookies
func (m *Manager) ClearCookies() error {
	if err := m.ensureInitialized(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.Evaluate("document.cookie.split(';').forEach(c => document.cookie = c.replace(/^ +/, '').replace(/=.*/, '=;expires=' + new Date().toUTCString() + ';path=/'))", nil),
	)
}

// SubmitForm submits a form
func (m *Manager) SubmitForm(selector string) error {
	if err := m.ensureInitialized(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.WaitVisible(selector),
		chromedp.Submit(selector),
	)
}

// FillForm fills multiple form fields
func (m *Manager) FillForm(fields map[string]string) error {
	if err := m.ensureInitialized(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
	defer cancel()

	tasks := make([]chromedp.Action, 0, len(fields)*2)
	for selector, value := range fields {
		tasks = append(tasks,
			chromedp.WaitVisible(selector),
			chromedp.SendKeys(selector, value),
		)
	}

	return chromedp.Run(ctx, tasks...)
}

// TakeFullPageScreenshot takes a full page screenshot
func (m *Manager) TakeFullPageScreenshot() ([]byte, error) {
	if err := m.ensureInitialized(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
	defer cancel()

	var buf []byte
	err := chromedp.Run(ctx,
		chromedp.FullScreenshot(&buf, 90),
	)

	return buf, err
}

// IsElementVisible checks if an element is visible
func (m *Manager) IsElementVisible(selector string) (bool, error) {
	if err := m.ensureInitialized(); err != nil {
		return false, err
	}

	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	defer cancel()

	var visible bool
	script := fmt.Sprintf(`
		(function() {
			const el = document.querySelector('%s');
			if (!el) return false;
			const rect = el.getBoundingClientRect();
			return rect.width > 0 && rect.height > 0;
		})()
	`, selector)

	err := chromedp.Run(ctx,
		chromedp.Evaluate(script, &visible),
	)

	return visible, err
}

