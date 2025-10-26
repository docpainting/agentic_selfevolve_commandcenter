package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	log.Println("🌐 Testing Browser Automation with ChromeDP")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Create Chrome context
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Set timeout
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Test 1: Navigate to GitHub
	log.Println("\n📍 Test 1: Navigate to GitHub")
	var title string
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://github.com"),
		chromedp.WaitReady("body"),
		chromedp.Title(&title),
	)
	if err != nil {
		log.Fatal("❌ Failed to navigate:", err)
	}
	log.Printf("✅ Page title: %s\n", title)

	// Test 2: Take screenshot
	log.Println("\n📸 Test 2: Capture screenshot")
	var screenshot []byte
	err = chromedp.Run(ctx,
		chromedp.CaptureScreenshot(&screenshot),
	)
	if err != nil {
		log.Fatal("❌ Failed to capture screenshot:", err)
	}
	
	screenshotPath := "/home/ubuntu/github_screenshot.png"
	err = os.WriteFile(screenshotPath, screenshot, 0644)
	if err != nil {
		log.Fatal("❌ Failed to save screenshot:", err)
	}
	log.Printf("✅ Screenshot saved: %s (%d bytes)\n", screenshotPath, len(screenshot))

	// Test 3: Find interactive elements
	log.Println("\n🔍 Test 3: Find interactive elements")
	var nodes []*chromedp.Node
	err = chromedp.Run(ctx,
		chromedp.Nodes("a, button, input", &nodes, chromedp.ByQueryAll),
	)
	if err != nil {
		log.Fatal("❌ Failed to find elements:", err)
	}
	log.Printf("✅ Found %d interactive elements\n", len(nodes))

	// Show first 10 elements
	log.Println("\n📋 First 10 elements:")
	for i, node := range nodes {
		if i >= 10 {
			break
		}
		text := node.AttributeValue("aria-label")
		if text == "" {
			text = node.NodeValue
		}
		if text == "" {
			text = node.AttributeValue("placeholder")
		}
		log.Printf("  [%d] <%s> %s\n", i, node.NodeName, text)
	}

	// Test 4: Search functionality
	log.Println("\n🔎 Test 4: Test search")
	var searchResults string
	err = chromedp.Run(ctx,
		chromedp.Navigate("https://github.com/search?q=go-light-rag&type=repositories"),
		chromedp.WaitReady("body"),
		chromedp.Sleep(2*time.Second),
		chromedp.Text("body", &searchResults, chromedp.ByQuery),
	)
	if err != nil {
		log.Fatal("❌ Failed to search:", err)
	}
	log.Println("✅ Search completed")

	// Test 5: Numbered overlay simulation
	log.Println("\n🎯 Test 5: Simulate numbered overlays")
	var overlayScript string
	overlayScript = `
		(function() {
			// Find all interactive elements
			const elements = document.querySelectorAll('a, button, input, textarea, select');
			let count = 0;
			const overlays = [];
			
			elements.forEach((el, index) => {
				const rect = el.getBoundingClientRect();
				if (rect.width > 0 && rect.height > 0 && rect.top >= 0 && rect.top < window.innerHeight) {
					// Create overlay
					const overlay = document.createElement('div');
					overlay.textContent = index;
					overlay.style.position = 'absolute';
					overlay.style.left = rect.left + 'px';
					overlay.style.top = rect.top + 'px';
					overlay.style.width = '24px';
					overlay.style.height = '24px';
					overlay.style.backgroundColor = '#15A7FF';
					overlay.style.color = 'white';
					overlay.style.borderRadius = '4px';
					overlay.style.display = 'flex';
					overlay.style.alignItems = 'center';
					overlay.style.justifyContent = 'center';
					overlay.style.fontSize = '12px';
					overlay.style.fontWeight = 'bold';
					overlay.style.zIndex = '10000';
					overlay.style.boxShadow = '0 0 10px rgba(21, 167, 255, 0.5)';
					overlay.style.border = '2px solid #15A7FF';
					
					document.body.appendChild(overlay);
					overlays.push(overlay);
					count++;
				}
			});
			
			return count;
		})();
	`

	var overlayCount int64
	err = chromedp.Run(ctx,
		chromedp.Navigate("https://github.com"),
		chromedp.WaitReady("body"),
		chromedp.Evaluate(overlayScript, &overlayCount),
	)
	if err != nil {
		log.Fatal("❌ Failed to create overlays:", err)
	}
	log.Printf("✅ Created %d numbered overlays\n", overlayCount)

	// Take screenshot with overlays
	var overlayScreenshot []byte
	err = chromedp.Run(ctx,
		chromedp.Sleep(500*time.Millisecond),
		chromedp.CaptureScreenshot(&overlayScreenshot),
	)
	if err != nil {
		log.Fatal("❌ Failed to capture overlay screenshot:", err)
	}

	overlayPath := "/home/ubuntu/github_with_overlays.png"
	err = os.WriteFile(overlayPath, overlayScreenshot, 0644)
	if err != nil {
		log.Fatal("❌ Failed to save overlay screenshot:", err)
	}
	log.Printf("✅ Overlay screenshot saved: %s (%d bytes)\n", overlayPath, len(overlayScreenshot))

	// Summary
	log.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("🎉 All browser tests passed!")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("\n📸 Screenshots saved:")
	log.Println("  1. " + screenshotPath)
	log.Println("  2. " + overlayPath)
	log.Println("\n✨ Browser automation is working perfectly!")
}

