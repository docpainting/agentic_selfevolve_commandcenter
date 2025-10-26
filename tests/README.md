# Browser Automation Tests

This directory contains test scripts to verify browser automation functionality.

## test_browser.go

Tests the ChromeDP browser automation with numbered overlays.

### What it tests:

1. **Navigation** - Navigate to GitHub
2. **Screenshot** - Capture page screenshot
3. **Element Detection** - Find all interactive elements (links, buttons, inputs)
4. **Search** - Test search functionality
5. **Numbered Overlays** - Inject cyan numbered boxes over elements (Rango-style)

### Run the test:

```bash
cd tests
go run test_browser.go
```

### Expected output:

```
🌐 Testing Browser Automation with ChromeDP
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📍 Test 1: Navigate to GitHub
✅ Page title: GitHub: Let's build from here · GitHub

📸 Test 2: Capture screenshot
✅ Screenshot saved: /home/ubuntu/github_screenshot.png (123456 bytes)

🔍 Test 3: Find interactive elements
✅ Found 150 interactive elements

📋 First 10 elements:
  [0] <A> Skip to content
  [1] <BUTTON> Toggle navigation
  [2] <INPUT> Search GitHub
  [3] <A> Pull requests
  [4] <A> Issues
  ...

🔎 Test 4: Test search
✅ Search completed

🎯 Test 5: Simulate numbered overlays
✅ Created 45 numbered overlays
✅ Overlay screenshot saved: /home/ubuntu/github_with_overlays.png (145678 bytes)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🎉 All browser tests passed!
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📸 Screenshots saved:
  1. /home/ubuntu/github_screenshot.png
  2. /home/ubuntu/github_with_overlays.png

✨ Browser automation is working perfectly!
```

### What the overlays look like:

The script injects cyan (#15A7FF) numbered boxes over every interactive element:

```
┌─────────────────────────────────────┐
│  GitHub                             │
│  ┌──┐                               │
│  │ 0│ Skip to content               │
│  └──┘                               │
│  ┌──┐                               │
│  │ 1│ [Search]                      │
│  └──┘                               │
│  ┌──┐  ┌──┐  ┌──┐                  │
│  │ 2│  │ 3│  │ 4│                  │
│  └──┘  └──┘  └──┘                  │
│  Pull   Issues  Marketplace        │
└─────────────────────────────────────┘
```

This is exactly how the agent will see and interact with web pages!

### Troubleshooting:

**Error: "chrome not found"**
```bash
# Install Chrome/Chromium
sudo apt-get install chromium-browser
```

**Error: "context deadline exceeded"**
- Increase timeout in script
- Check internet connection
- Try different URL

**No overlays visible**
- Check screenshot file
- Verify JavaScript execution
- Ensure elements are in viewport
