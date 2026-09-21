---
name: movie-cli-rest-api-and-web-ui
description: "Local REST API endpoints, routing, middleware, embedded React/Vite dashboard, and interactive HTML reports in movie-cli-v8."
---

# Movie CLI REST API and Web UI Skill

## Overview

`movie-cli-v8` embeds an HTTP server and web dashboard that exposes library data over REST endpoints, powers the React 18 + Vite UI, and enables interactive filtering, tagging, and playback from modern web browsers.

## REST Server Architecture (`cmd/movie_rest*.go`)

### Starting the Server
```sh
movie rest [flags]
movie ui              # Alias: starts server and auto-opens default browser
movie scan --rest     # Runs a scan, boots REST server, and opens report
```

### Configuration Flags
- `--port <N>`, `-p`: Configures HTTP listen port (default `8086`).
- `--open`: Automatically launches default browser at `http://localhost:<port>`.

### Route Mappings & Handlers
| Method | Endpoint | Handler File | Description |
|---|---|---|---|
| `GET` | `/api/media` | `movie_rest_handlers.go` | Returns paginated media collection |
| `GET` | `/api/media/{id}` | `movie_rest_handlers.go` | Returns single media record by ID |
| `PATCH` | `/api/media/{id}` | `movie_rest_handlers.go` | Updates media fields (title, year, etc.) |
| `DELETE`| `/api/media/{id}` | `movie_rest_handlers.go` | Soft-deletes media item from library |
| `GET` | `/api/media/{id}/details`| `movie_rest_details.go` | Full modal payload: media, genres, cast, tags, similar |
| `GET` | `/api/media/{id}/similar`| `movie_rest_details.go` | TMDb recommendations based on item |
| `PATCH` | `/api/media/{id}/watched`| `movie_rest_handlers.go` | Toggles watched status in watchlist |
| `GET` | `/api/tags` | `movie_rest_handlers.go` | Lists all tags with media counts |
| `POST` | `/api/tags` | `movie_rest_handlers.go` | Associates a tag with a media item |
| `DELETE`| `/api/tags` | `movie_rest_handlers.go` | Dissociates a tag from a media item |
| `GET` | `/api/stats` | `movie_rest_handlers.go` | Returns library counts, storage, top genres |
| `GET` | `/api/dashboard/filters` | `movie_rest_dashboard.go`| Filter facets: genres, years, types, tag lists |
| `GET` | `/api/dashboard/list` | `movie_rest_dashboard.go`| Filterable, sortable, paginated card listing |
| `GET` | `/api/thumbnails/{id}` | `movie_rest_thumb.go` | Serves locally cached poster/backdrop images |
| `GET` | `/api/export` | `movie_rest_export.go` | Exports media collection as JSON or CSV |

## Web Dashboard & Frontend Architecture (`src/`)

- **Tech Stack:** React 18, Vite, TypeScript, Tailwind CSS, Radix UI primitives, Lucide React icons.
- **Entry Points:**
  - `index.html`: Web root referencing `src/main.tsx`.
  - `src/App.tsx`: Top-level router and layout providers.
  - `src/pages/Index.tsx`: Main dashboard view rendering media grid, search input, genre filters, and sort options.
- **Components:**
  - `src/components/MediaCard.tsx`: Card rendering poster art, title, release year, rating badge, and action popovers.
  - `src/components/MediaDetailModal.tsx`: Comprehensive modal displaying backdrop, overview, trailer video embed, genres, cast, and tag manager.
  - `src/components/FilterBar.tsx`: Faceted search controls (genre pills, movie vs TV toggle, watched filter).

## Standalone HTML Report (`cmd/movie_scan_html.go`, `templates/`)

- Generated automatically at `.movie-output/report.html` during `movie scan`.
- Self-contained single-page report embedding static CSS and JavaScript so users can view scan results offline without running a server.
- When the local REST server is active, `report.html` connects dynamically to `http://localhost:8086` to enable in-browser playback, tagging, and status updates.
