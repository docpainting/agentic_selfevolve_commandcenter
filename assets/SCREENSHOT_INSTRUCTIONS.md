# Screenshots for Repository

## Main Preview Screenshot

To add the main preview screenshot to the README:

### Option 1: Take Your Own Screenshot

1. Start the dev server:
   ```bash
   cd frontend
   npm run dev
   ```

2. Open http://localhost:3000 in your browser

3. Take a full-page screenshot (recommended tools):
   - **Chrome DevTools**: Cmd/Ctrl + Shift + P → "Capture full size screenshot"
   - **Firefox**: Right-click → "Take Screenshot" → "Save full page"
   - **macOS**: Cmd + Shift + 4 (then select area)
   - **Windows**: Win + Shift + S

4. Save as `assets/preview.png` or `assets/preview.jpg`

5. Add to README.md after the title:
   ```markdown
   <div align="center">
   
   ![Agentic Command Center Preview](assets/preview.png)
   
   </div>
   ```

### Option 2: Use Figma/Design Tool

Create a mockup showing:
- **Left Panel**: File tree with folders expanded
- **Center Panel**: Chat interface with sample conversation
- **Right Panel**: OpenEvolve with progress indicators
- **Bottom Panel**: Terminal tab showing commands
- **Midnight glassmorphism** theme (dark blue background, glass panels, cyan accents)

### Recommended Screenshot Composition

**Full Interface View** (1920x1080 or higher):
- Show all 4 panels (left, center, right, bottom)
- Include connection status indicators at top
- Show some sample content in each panel
- Capture the midnight glassmorphism aesthetic

**Key Features to Highlight**:
1. Beautiful glassmorphism UI
2. File tree on left
3. Chat with agent responses in center
4. OpenEvolve panel on right
5. Terminal/Browser tabs at bottom
6. Numbered browser overlays (if showing browser tab)
7. Connection status indicators

### Additional Screenshots (Optional)

Create these for a complete showcase:

1. **`assets/chat-interface.png`** - Close-up of chat panel
2. **`assets/browser-automation.png`** - Browser panel with numbered overlays
3. **`assets/terminal-panel.png`** - Terminal with AI/User commands
4. **`assets/openevolve-panel.png`** - OpenEvolve watchdog alerts
5. **`assets/mcp-tools.png`** - MCP integration panel
6. **`assets/file-tree.png`** - File tree with VS Code-style layout

### Image Specifications

- **Format**: PNG (for transparency) or JPG (for smaller size)
- **Resolution**: Minimum 1920x1080, prefer 2560x1440 or 4K
- **Compression**: Use TinyPNG or similar to reduce file size
- **Max file size**: Keep under 2MB for GitHub

### Adding to README

Once you have screenshots, update README.md:

```markdown
## 🎨 Preview

<div align="center">

![Main Interface](assets/preview.png)

*The complete IDE/GUI command center with midnight glassmorphism design*

</div>

### Feature Highlights

<table>
  <tr>
    <td width="50%">
      <img src="assets/chat-interface.png" alt="Chat Interface"/>
      <p align="center"><b>AI Chat Interface</b></p>
    </td>
    <td width="50%">
      <img src="assets/browser-automation.png" alt="Browser Automation"/>
      <p align="center"><b>Browser Automation</b></p>
    </td>
  </tr>
  <tr>
    <td width="50%">
      <img src="assets/terminal-panel.png" alt="Terminal"/>
      <p align="center"><b>Integrated Terminal</b></p>
    </td>
    <td width="50%">
      <img src="assets/openevolve-panel.png" alt="OpenEvolve"/>
      <p align="center"><b>Self-Evolution Tracking</b></p>
    </td>
  </tr>
</table>
```

### Temporary Solution

Until you add real screenshots, you can add this to README:

```markdown
## 🎨 Preview

> **Note**: Screenshots coming soon! For now, start the dev server to see the beautiful midnight glassmorphism UI in action:
> 
> ```bash
> cd frontend && npm run dev
> ```
> 
> Then open http://localhost:3000

**Key Visual Features:**
- 🌙 Midnight blue gradient background
- ✨ Glass panels with backdrop blur
- 💎 Cyan (#15A7FF) accents for AI elements
- 🎨 Smooth animations and lens flare effects
- 📱 Responsive 3-panel layout
- 🖥️ Integrated browser and terminal
```

---

## Quick Screenshot Checklist

- [ ] Start dev server (`npm run dev`)
- [ ] Open in browser (http://localhost:3000)
- [ ] Take full-page screenshot
- [ ] Save as `assets/preview.png`
- [ ] Optimize image size (< 2MB)
- [ ] Add to README.md
- [ ] Commit and push to GitHub
- [ ] Verify image displays on GitHub

---

**Pro Tip**: Take screenshots in dark mode browser for best contrast with the midnight theme!

