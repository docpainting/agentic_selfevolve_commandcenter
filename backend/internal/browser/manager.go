package browser

import (
	"context"
	"fmt"
	"sync"
	"time"

	"agent-workspace/backend/internal/memory"
	"agent-workspace/backend/pkg/models"

	"github.com/chromedp/chromedp"
)

// Manager manages browser automation
type Manager struct {
	ctx          context.Context
	cancel       context.CancelFunc
	allocCtx     context.Context
	allocCancel  context.CancelFunc
	shortTermMem *memory.ShortTermMemory
	currentURL   string
	elements     []models.BrowserElement
	mu           sync.RWMutex
	initialized  bool
}

// NewManager creates a new browser manager
func NewManager(shortTermMem *memory.ShortTermMemory) *Manager {
	return &Manager{
		shortTermMem: shortTermMem,
		elements:     make([]models.BrowserElement, 0),
	}
}

// Initialize initializes the browser context
func (m *Manager) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return nil
	}

	// Create allocator context
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	m.allocCtx = allocCtx
	m.allocCancel = allocCancel

	// Create browser context
	ctx, cancel := chromedp.NewContext(allocCtx)
	m.ctx = ctx
	m.cancel = cancel

	m.initialized = true
	return nil
}

// Navigate navigates to a URL
func (m *Manager) Navigate(url string) error {
	if err := m.ensureInitialized(); err != nil {
		return err
	}

	m.mu.Lock()
	m.currentURL = url
	m.mu.Unlock()

	ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
	)
}

// Click clicks an element by ID
func (m *Manager) Click(elementID int) error {
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

	ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
	defer cancel()

	// Click at element coordinates
	return chromedp.Run(ctx,
		chromedp.MouseClickXY(element.X+element.Width/2, element.Y+element.Height/2),
	)
}

// Type types text into an element
func (m *Manager) Type(elementID int, text string) error {
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

	ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
	defer cancel()

	// Click element first, then type
	return chromedp.Run(ctx,
		chromedp.MouseClickXY(element.X+element.Width/2, element.Y+element.Height/2),
		chromedp.Sleep(100*time.Millisecond),
		chromedp.SendKeys("body", text),
	)
}

// GetCurrentURL returns the current URL
func (m *Manager) GetCurrentURL() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentURL
}

// GetElements returns detected elements
func (m *Manager) GetElements() []models.BrowserElement {
	m.mu.RLock()
	defer m.mu.RUnlock()

	elements := make([]models.BrowserElement, len(m.elements))
	copy(elements, m.elements)
	return elements
}

// SetElements sets the detected elements
func (m *Manager) SetElements(elements []models.BrowserElement) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.elements = elements
}

// Scroll scrolls the page
func (m *Manager) Scroll(x, y int) error {
	if err := m.ensureInitialized(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.Evaluate(fmt.Sprintf("window.scrollBy(%d, %d)", x, y), nil),
	)
}

// ExecuteScript executes JavaScript
func (m *Manager) ExecuteScript(script string) (interface{}, error) {
	if err := m.ensureInitialized(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
	defer cancel()

	var result interface{}
	err := chromedp.Run(ctx,
		chromedp.Evaluate(script, &result),
	)

	return result, err
}

// GetPageTitle returns the page title
func (m *Manager) GetPageTitle() (string, error) {
	if err := m.ensureInitialized(); err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	defer cancel()

	var title string
	err := chromedp.Run(ctx,
		chromedp.Title(&title),
	)

	return title, err
}

// GetPageHTML returns the page HTML
func (m *Manager) GetPageHTML() (string, error) {
	if err := m.ensureInitialized(); err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
	defer cancel()

	var html string
	err := chromedp.Run(ctx,
		chromedp.OuterHTML("html", &html),
	)

	return html, err
}

// WaitForElement waits for an element to appear
func (m *Manager) WaitForElement(selector string, timeout time.Duration) error {
	if err := m.ensureInitialized(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(m.ctx, timeout)
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.WaitVisible(selector),
	)
}

// Cleanup closes the browser
func (m *Manager) Cleanup() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return nil
	}

	if m.cancel != nil {
		m.cancel()
	}

	if m.allocCancel != nil {
		m.allocCancel()
	}

	m.initialized = false
	return nil
}

// ensureInitialized ensures the browser is initialized
func (m *Manager) ensureInitialized() error {
	m.mu.RLock()
	initialized := m.initialized
	m.mu.RUnlock()

	if !initialized {
		return m.Initialize()
	}

	return nil
}

// GetContext returns the browser context (for advanced operations)
func (m *Manager) GetContext() context.Context {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ctx
}

