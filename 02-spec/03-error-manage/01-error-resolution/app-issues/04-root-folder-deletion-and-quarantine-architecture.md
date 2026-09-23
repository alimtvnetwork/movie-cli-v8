# Root Cause Analysis: Scan Root Directory Deletion Prevention & Two-Stage Quarantine Architecture

**Document ID:** `04-root-folder-deletion-and-quarantine-architecture.md`  
**Classification:** 🔴 Critical Serious Issue / Code Red Post-Mortem  
**Component:** File Management / Web UI / REST Staged Deletions / Quarantine Engine  
**Version Added:** v2.343.0  
**AI Confidence:** High  
**Ambiguity:** None  

---

## 1. Incident Description & Symptom

When a user triggered a movie removal from the Web UI for a video file residing directly inside a scan root folder (e.g., `# Movies/Agent.Jita.mkv`), the deletion handler removed the entire root scan folder (`# Movies`) containing hundreds of media items rather than solely the specific movie file.

Furthermore, on network-attached mounts and virtualization shares (such as VMware Host-Guest Shared Folders on drive `Z:`), Windows does not provide a Recycle Bin facility. Direct OS unlinking on network shares resulted in permanent, unrecoverable file loss.

---

## 2. Root Cause Analysis (4-Part RCA)

### 2.1 Direct Cause
In `cmd/movie_rest_staged.go` and `templates/report.html`, the Web UI exposed a folder-level deletion action (`delete_folder: true`). When executing folder deletion for a movie whose path was `Z:\DownloadRelated\DownloadCompletedVm\# Movies\Agent.Jita.mkv`, the backend calculated the target path via `filepath.Dir(media.CurrentFilePath)`. This resolved to `Z:\DownloadRelated\DownloadCompletedVm\# Movies` — the entire library root directory.

### 2.2 Lack of Scan Root & Ancestor Validation
The removal handlers previously lacked verification against the active SQLite database's `ScanFolder` table. There was no check confirming whether a target deletion directory was:
1. A filesystem root (`/`, `C:\`, `Z:\`).
2. An operating system system directory (`Windows`, `Program Files`, user home directory).
3. A registered library scan root where scanning starts.
4. An ancestor directory of a library scan root.
5. A shared directory containing multiple active movies.

### 2.3 Bypass of Safe Staging & Direct Deletion
The frontend offered an immediate deletion path (`?immediate=true`) that bypassed batch staging and called `trashbin.MoveToTrash(filePath)` directly. Because network shares do not support the Windows Recycle Bin, this acted as an immediate irreversible deletion.

### 2.4 Lack of User Intent Confirmation on High-Volume Actions
Large deletions (dozens of gigabytes or dozens of files) did not enforce high-stakes intent verification in the terminal. A single UI button click was sufficient to initiate catastrophic deletion.

---

## 3. Implemented Architectural Solution

### 3.1 Strict Safeguard Engine (`cmd/safeguard_removal.go`)
- **`validateRemovalTarget(targetPath, database)`**:
  - Rejects empty paths, filesystem roots, and system directories.
  - Queries `ScanFolder` table in `movie.db`. Rejects any registered scan root or ancestor directory.
  - Checks active movies in directory; blocks whole-directory removal if directory contains more than 1 active movie.
- **`resolveMediaRemovalTarget(media, database)`**:
  - Inspects file structure.
  - If parent directory is a registered scan root, targets ONLY the individual movie file and its sidecars (`.nfo`, `.srt`, `.vtt`, `.jpg`, `.png`).
  - Automatically identifies dedicated single-movie folders versus shared folders.

### 3.2 Two-Stage Temporary Deletion Quarantine (`temp-remove`)
- Movies approved for removal are never deleted directly.
- Files and sidecars are moved atomically (`os.Rename`) on the same filesystem/share into:
  `<scanRoot>/temp-remove/<timestamp>_<mediaId>_<sanitizedTitle>/`
- The scanner (`cmd/movie_scan_collect.go`) strictly ignores `temp-remove` during directory scans.

### 3.3 Database Audit Logging (`db/migrate_v8.go` & `db/task.go`)
- Migration v8 introduces the `Task` table to track every stage: `pending`, `quarantined`, `purged`, `undone`, `failed`.
- Full undo capability via `movie undo` (`cmd/movie_undo_exec.go`), which restores quarantined files back to their original path and restores the database media record.

### 3.4 Terminal Exit Verification & High-Volume Contingency Rule
- When the Web UI server shuts down (`cmd/movie_ui.go`), `promptPendingQuarantinePurge` inspects pending tasks in `temp-remove`.
- **Standard Removal (< 15 items and < 30 GB)**: Displays item list and accepts `CONFIRM` or random title.
- **High-Volume Removal (> 15 items OR > 30 GB)**:
  - Triggers the High-Volume Safeguard.
  - `CONFIRM` is explicitly prohibited.
  - The terminal prints a random movie title from the queued list (e.g. `>>> Inception <<<`).
  - The user MUST type this exact movie title in the terminal to authorize permanent purge.
  - Pressing Enter or entering any other input safely preserves all files in `temp-remove`.

---

## 4. Strictly Avoid Rules for Future Development

1. 🔴 **NEVER delete or wipe a scan root directory.** Under no circumstance may any function remove a directory matching `ScanFolder` or an ancestor directory.
2. 🔴 **NEVER execute direct hard/unlinked deletion of movies.** All removals MUST stage to `<scanRoot>/temp-remove/` first.
3. 🔴 **NEVER bypass terminal confirmation for permanent purging.** Only explicit terminal verification can purge items from `temp-remove`.
4. 🔴 **High-Volume Contingency Mandate:** Large removals (> 15 items or > 30 GB) MUST enforce exact random movie title verification in the terminal.
