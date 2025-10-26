package browser

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"time"

	"agent-workspace/backend/pkg/models"

	"github.com/chromedp/chromedp"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// CaptureScreenshot captures a screenshot of the current page
func (m *Manager) CaptureScreenshot(taskID string) ([]byte, error) {
	if err := m.ensureInitialized(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
	defer cancel()

	var buf []byte
	err := chromedp.Run(ctx,
		chromedp.CaptureScreenshot(&buf),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to capture screenshot: %w", err)
	}

	return buf, nil
}

// AnalyzeScreenshot analyzes a screenshot and detects elements
func (m *Manager) AnalyzeScreenshot(req models.VisionAnalyzeRequest) (interface{}, error) {
	// Capture screenshot if not provided
	screenshot := req.Screenshot
	if len(screenshot) == 0 {
		var err error
		screenshot, err = m.CaptureScreenshot(req.TaskID)
		if err != nil {
			return nil, err
		}
	}

	// Detect interactive elements
	elements, err := m.detectElements()
	if err != nil {
		return nil, fmt.Errorf("failed to detect elements: %w", err)
	}

	// Store elements
	m.SetElements(elements)

	// Create analysis result
	analysis := map[string]interface{}{
		"task_id":          req.TaskID,
		"goal":             req.Goal,
		"screenshot_size":  len(screenshot),
		"elements_count":   len(elements),
		"elements":         elements,
		"current_url":      m.GetCurrentURL(),
		"timestamp":        time.Now().Format(time.RFC3339),
	}

	return analysis, nil
}

// DrawNumberedOverlays draws numbered boxes on screenshot
func (m *Manager) DrawNumberedOverlays(screenshot []byte, elements []models.BrowserElement) ([]byte, error) {
	// Decode screenshot
	img, err := png.Decode(bytes.NewReader(screenshot))
	if err != nil {
		return nil, fmt.Errorf("failed to decode screenshot: %w", err)
	}

	// Create RGBA image for drawing
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// Draw overlays for each element
	cyan := color.RGBA{21, 167, 255, 200} // #15A7FF with alpha
	white := color.RGBA{255, 255, 255, 255}

	for i, element := range elements {
		// Draw rectangle
		drawRect(rgba, int(element.X), int(element.Y), int(element.Width), int(element.Height), cyan)

		// Draw number
		label := fmt.Sprintf("%d", i)
		drawLabel(rgba, int(element.X)+2, int(element.Y)+2, label, white)
	}

	// Encode back to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, rgba); err != nil {
		return nil, fmt.Errorf("failed to encode screenshot: %w", err)
	}

	return buf.Bytes(), nil
}

// detectElements detects interactive elements on the page
func (m *Manager) detectElements() ([]models.BrowserElement, error) {
	if err := m.ensureInitialized(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
	defer cancel()

	// JavaScript to detect interactive elements
	script := `
	(function() {
		const elements = [];
		const selectors = [
			'a[href]',
			'button',
			'input',
			'textarea',
			'select',
			'[role="button"]',
			'[onclick]',
			'[tabindex]'
		];

		const seen = new Set();
		
		selectors.forEach(selector => {
			document.querySelectorAll(selector).forEach(el => {
				if (seen.has(el)) return;
				seen.add(el);

				const rect = el.getBoundingClientRect();
				if (rect.width === 0 || rect.height === 0) return;
				if (rect.top < 0 || rect.left < 0) return;

				elements.push({
					x: rect.left,
					y: rect.top,
					width: rect.width,
					height: rect.height,
					text: el.innerText?.substring(0, 100) || el.value || el.placeholder || '',
					tag: el.tagName.toLowerCase(),
					clickable: true
				});
			});
		});

		return elements;
	})();
	`

	var result []map[string]interface{}
	err := chromedp.Run(ctx,
		chromedp.Evaluate(script, &result),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to evaluate script: %w", err)
	}

	// Convert to BrowserElement
	elements := make([]models.BrowserElement, 0, len(result))
	for i, elem := range result {
		elements = append(elements, models.BrowserElement{
			ID:        i,
			X:         getFloat(elem, "x"),
			Y:         getFloat(elem, "y"),
			Width:     getFloat(elem, "width"),
			Height:    getFloat(elem, "height"),
			Text:      getString(elem, "text"),
			Tag:       getString(elem, "tag"),
			Clickable: getBool(elem, "clickable"),
		})
	}

	return elements, nil
}

// GetScreenshotWithOverlays captures screenshot with numbered overlays
func (m *Manager) GetScreenshotWithOverlays(taskID string) ([]byte, error) {
	// Capture screenshot
	screenshot, err := m.CaptureScreenshot(taskID)
	if err != nil {
		return nil, err
	}

	// Detect elements
	elements, err := m.detectElements()
	if err != nil {
		return nil, err
	}

	// Store elements
	m.SetElements(elements)

	// Draw overlays
	return m.DrawNumberedOverlays(screenshot, elements)
}

// SaveScreenshot saves a screenshot to file
func (m *Manager) SaveScreenshot(taskID, filepath string) error {
	screenshot, err := m.CaptureScreenshot(taskID)
	if err != nil {
		return err
	}

	// TODO: Save to file system
	// For now, store in short-term memory
	if m.shortTermMem != nil {
		// Store screenshot in memory
		_ = screenshot // Use screenshot
	}

	return nil
}

// Helper functions

func drawRect(img *image.RGBA, x, y, width, height int, col color.Color) {
	// Draw top line
	for i := x; i < x+width; i++ {
		img.Set(i, y, col)
		img.Set(i, y+1, col)
	}

	// Draw bottom line
	for i := x; i < x+width; i++ {
		img.Set(i, y+height, col)
		img.Set(i, y+height-1, col)
	}

	// Draw left line
	for i := y; i < y+height; i++ {
		img.Set(x, i, col)
		img.Set(x+1, i, col)
	}

	// Draw right line
	for i := y; i < y+height; i++ {
		img.Set(x+width, i, col)
		img.Set(x+width-1, i, col)
	}
}

func drawLabel(img *image.RGBA, x, y int, label string, col color.Color) {
	point := fixed.Point26_6{
		X: fixed.Int26_6(x * 64),
		Y: fixed.Int26_6((y + 10) * 64),
	}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}

	d.DrawString(label)
}

func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return 0
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// EncodeScreenshotBase64 encodes screenshot as base64
func EncodeScreenshotBase64(screenshot []byte) string {
	return base64.StdEncoding.EncodeToString(screenshot)
}

// DecodeScreenshotBase64 decodes base64 screenshot
func DecodeScreenshotBase64(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

