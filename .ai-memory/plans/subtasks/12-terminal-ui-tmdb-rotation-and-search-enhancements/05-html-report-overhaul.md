# Subtask 05: Modernized Standalone HTML Report Overhaul

> **Parent Plan:** `.ai-memory/plans/pending/12-terminal-ui-tmdb-rotation-and-search-enhancements.md`  
> **Status:** Complete  
> **Target Files:**
> - `templates/report.html`
> - `cmd/movie_scan_html.go`

---

## 1. Problem Statement

1. The HTML report template (`templates/report.html`) contains an unstyled `.rest-banner` element:
   ```html
   <div class="rest-banner" id="restBanner">
     <span class="dot"></span>
     Run <code>movie rest</code> or <code>movie ui</code> to enable interactive features (port <strong>{{.Port}}</strong>)
   </div>
   ```
   Because `.rest-banner` has zero CSS styling, it renders awkwardly as raw plain text with no visual polish or copy-to-clipboard functionality.
2. The user specifically instructed:
   "In lot of cases, the HTML that you create, that needs to be changed. I asked several times. You didn't do it. And at the end, you do not show the movie UI command. That is also terrible. You need to suggest to the user so that it's a friendly one. Okay?"
3. The report needs a modern glassmorphism banner, clear interactive instructions, polished responsive cards, and an easy one-click copy command for `movie ui`.

---

## 2. Proposed Changes

### A. Polished Hero Banner & UI Prompt (`templates/report.html`)
- Add comprehensive styling for `.rest-banner` / `.hero-guidance-banner`:
  - Glassmorphic gradient background with subtle border and shadow.
  - Clear, prominent callout: "🚀 Launch Full Interactive Web Dashboard: `movie ui`" with a 1-click "📋 Copy Command" button.
  - Server status indicator dot (pulsing green if API is detected, amber if static report mode).
  - Quick instruction pills explaining what `movie ui` does (live search, playback, tagging, edit, trash bin).

### B. Responsive Glassmorphism Cards & Badges (`templates/report.html`)
- Ensure all media cards display smooth hover transitions, clean poster fallback icons, genre pills, rating stars, and file size badges.
- For items without TMDb metadata, display a subtle "Local File" badge so users immediately know which items need manual matching or were scanned offline.

### C. Search & Filter Usability
- Ensure the quick search input and genre/type dropdowns filter cards instantly with responsive counts.

---

## 3. Verification Criteria
- `report.html` renders with a modern, beautiful dark-theme glassmorphism hero banner promoting `movie ui`.
- One-click copy for `movie ui` works seamlessly in browser.
- All media items (with and without posters) render attractively without clipping or awkward line wraps.
