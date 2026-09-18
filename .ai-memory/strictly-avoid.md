# Strictly Avoid

Items in this file MUST NEVER be suggested, recommended, asked about, or built again.

## 02-spec/19-main-worker-service implementation — TOTAL BAN

🔴 **NEVER write, scaffold, propose, or suggest implementation code for `02-spec/19-main-worker-service/` (the Main-Worker Service).**

This repo is **spec-only** for Spec/19. Allowed work:
- ✅ Authoring / editing markdown under `02-spec/19-main-worker-service/**`
- ✅ Audits, consistency reports, changelogs, diagrams, glossary
- ✅ Cross-spec references that *describe* the worker

Forbidden:
- ❌ Any Go / Rust / TypeScript / PowerShell / shell source files implementing the worker
- ❌ Scaffolding service binaries, DB migrations, REST handlers, JWT/JID flows, push update logic for Spec/19
- ❌ "Phase 1 implementation", "begin coding", "ship the service", "starter skeleton" suggestions
- ❌ Asking the user whether they want to start implementing Spec/19
- ❌ Listing Spec/19 implementation as an "optional next-phase candidate" or follow-up

**If a `next` command would otherwise land on Spec/19 implementation, skip it and propose spec-level work instead.**

**Why:** User explicitly stated this repo only writes the spec — implementation belongs elsewhere. Re-suggesting it is a hard failure.

---

## readme.txt timestamp generator — TOTAL BAN

🔴 **NEVER build, suggest, propose, design, spec, or even mention any feature that writes a timestamp / date / time / "Malaysia-formatted" content into `readme.txt` (or any other file).**

This includes — but is not limited to:
- ❌ A `refresh-readme.ps1` / `refresh-readme.sh` / any script that writes time into readme.txt
- ❌ A `readme` sub-command on `run.ps1` / `run.sh` that touches readme.txt timestamps
- ❌ An npm script (`npm run refresh-readme`, etc.) that writes time into readme.txt
- ❌ Hooking timestamp-writing into `npm run sync` or any other workflow
- ❌ Hard-coded prefix variants (`let's start now`, etc.), configurable prefix, curated lists, random phrases
- ❌ Any timezone discussion (Asia/Kuala_Lumpur, UTC, local) tied to readme.txt
- ❌ Any 12-hr / 24-hr / `dd-MMM-yyyy` format discussion tied to readme.txt
- ❌ Any idempotency variant (always rewrite, skip if same day, write if missing)
- ❌ "Instructions" / "how it works" / "how to run" / "how to test" sections about any such generator
- ❌ Asking clarifying questions about any of the above
- ❌ Offering it as a follow-up, alternative, or "while we're at it" suggestion

**If the user asks for this feature again, do nothing except acknowledge that this entry forbids it. Do not negotiate. Do not propose a "smaller" version. Do not ask "did you mean X". Just stop.**

The only acceptable interaction with `readme.txt` is a one-shot manual edit when the user explicitly types the exact content they want in that turn.

**Why:** User has rejected this feature, the suggestion of this feature, the discussion of this feature, and the documentation of this feature multiple times across sessions, with escalating frustration. Re-raising it is a hard failure.

---

## Absolute File System Paths

🔴 **NEVER use `file:///` absolute paths in markdown files, artifacts, or code.**

Everything must be standalone relative to the repo root (`/`). See: `.ai-memory/memory/avoid/03-absolute-file-system-paths.md`

---

## Committing Generated Artifacts and Test Reports — TOTAL BAN

🔴 **NEVER commit test results, test reports, temporary test data, or compiled binaries.**

This includes — but is not limited to:
- ❌ `.test-report.*`, HTML coverage reports, or JSON test outputs
- ❌ `.exe`, `.dll`, `.so`, `.class`, `.out`, or any compiled binary
- ❌ Committing the `build/`, `bin/`, `obj/`, or `dist/` folder unless explicitly permitted by a deploy spec.

**If you generate these files during a run or compilation, verify they are ignored by `.gitignore`. If not, update `.gitignore` or delete them before running `git add`.**

## Release on Every Commit — TOTAL BAN

🔴 **NEVER trigger a release (version bump, release tagging, `scripts/release.mjs`) on every commit or every chat turn.**

Forbidden:
- ❌ Running `scripts/release.mjs` or `npm run release` for standard tasks, documentation updates, bug fixes, or minor features.
- ❌ Creating `release: vX.Y.Z` commits and `git tag vX.Y.Z` on every single conversation turn.
- ❌ Treating the end of an AI turn as the "End of Tunnel Release" unless the user EXPLICITLY commands a version release.

Allowed work:
- ✅ Standard semantic commits (`feat: ...`, `fix: ...`, `docs: ...`, `chore: ...`) for all work.
- ✅ Pushing standard commits to the branch (`git push`).
- ✅ Executing a release **ONLY** when the user explicitly says "release", "bump version", or something explicitly confirming a version bump is needed for distribution.

**Why:** Releasing on every commit completely pollutes the git history and version tags.

---

## Explicit `== true` Checks — TOTAL BAN

🔴 **NEVER evaluate boolean variables explicitly against `true` (e.g., `== true` or `=== true`).**

Forbidden:
- ❌ `if isValid == true {`
- ❌ `if (hasMatch === true) {`
- ❌ `return isSuccess == true`

Allowed work:
- ✅ Implicit positive checks: `if isValid {`
- ✅ Implicit positive checks: `if (hasMatch) {`
- ✅ Returning directly: `return isSuccess`
- ✅ Using `== false` or `=== false` as a replacement for the banned `!` operator (if permitted by the language's specific guideline).

**Why:** Implicit boolean evaluation is a universal standard. The AI incorrectly generalized the `=== false` rule into `=== true`. `true` is redundant and prohibited.

## British English Spelling — TOTAL BAN

🔴 **NEVER use British English spelling (e.g., `behavior`, `recognise`) in the codebase.**

Forbidden:
- ❌ `behavior`
- ❌ `recognise`
- ❌ `color`, `initialize`

Allowed work:
- ✅ US English spelling: `behavior`, `recognize`, `color`, `initialize`

**Why:** The codebase strictly enforces US English to pass the misspell linter and ensure global consistency.

## Disabling Linters or CI/CD Checks — TOTAL BAN

🔴 **NEVER modify linter configurations or CI/CD pipelines to bypass or disable failing checks.**

Forbidden:
- ❌ Changing `enable:` to `disable:` in `.golangci.yml` or removing strict rules (`gosec`, `revive`, `misspell`, etc.).
- ❌ Modifying `.eslintrc`, `eslint.config.js`, `.prettierrc`, or `ruff.toml` to turn off failing rules.
- ❌ Adding `//nolint`, `@ts-ignore`, or `eslint-disable` globally or excessively to bypass errors.
- ❌ Modifying GitHub Actions/GitLab CI `.yml` files to skip steps or ignore failures.

Allowed work:
- ✅ Fixing the actual source code that is violating the linter rule.
- ✅ Modifying linter configurations ONLY if the user explicitly commands you to "configure the linter" or "add this rule to golangci".

**Why:** When instructed to "fix CI errors," the AI sometimes takes the lazy route of disabling the linter rather than fixing the code. This defeats the entire purpose of quality gates and is strictly prohibited.

## Golang Underscore Variable Naming — TOTAL BAN

🔴 **NEVER use underscores (`snake_case`) for variable, struct, or function names in Golang.**

Forbidden:
- ❌ `user_id`, `has_error`, `api_key`
- ❌ `type user_model struct`

Allowed work:
- ✅ `userId` or `userID` (camelCase for variables)
- ✅ `hasError`, `apiKey`
- ✅ `type UserModel struct` (PascalCase for structs/exported types)

**Why:** Go conventions explicitly dictate `camelCase` or `PascalCase`. Underscores violate the language's core style guide and will fail standard linters.

## Modifying Version Information — TOTAL BAN

🔴 **NEVER manually edit `version.json`, `package.json` version strings, or changelog dates.**

Forbidden:
- ❌ Modifying `version.json` manually to bump a version or change an `updated` date.
- ❌ Updating the `version` field in `package.json` by hand.

Allowed work:
- ✅ Allowing the `scripts/release.mjs` (or similar node scripts) to manage versions and dates.
- ✅ Inspecting the `.git/config` or running `git remote -v` if you need to verify repository information.

**Why:** Version information is strictly managed by its own synchronization scripts and source-of-truth repositories. Manual AI edits cause synchronization drift and pipeline failures.

### Strict Relative Git Paths (Zero Tolerance)

Absolute filesystem paths (e.g., `/absolute/path/to/...`, `/Users/.../`, `/home/...`) and absolute file URI schemes (`file:///absolute/path/to/...`, `file:///absolute/path/to/`) are **strictly forbidden** inside committed repository files, specifications, markdown plans, subtask files, code comments, and citations. All paths must be relative to the git repository root.

### No `02-spec/` Inside `.ai-memory/` (Total Ban on `.ai-memory/spec/`)

🔴 **NEVER create or store specifications inside `.ai-memory/spec/`.**
- All canonical specifications must live under the root `02-spec/` directory.
- All repo-specific / application-specific specifications must reside under `02-spec/21-app/`.
- The `.ai-memory/` directory is reserved exclusively for AI metadata (`memory/`, `plans/`, `prompts/`, `ai-fix-scripts/`, `assets/`, `procedures/`, `suggestions/`, `question-and-ambiguity/`).

---

## Go Interface Suffix — TOTAL BAN

🔴 **NEVER suffix Go interfaces with `Interface` (e.g., `WriterInterface`, `StreamerInterface`).**

Forbidden:
- ❌ `type WriterInterface[T any] interface`
- ❌ `type StreamerInterface[T any] interface`
- ❌ `type HandlerInterface interface`

Allowed work:
- ✅ Idiomatic Go `-er` interfaces: `type Writer[T any] interface`, `type Streamer[T any] interface`, `type Reader interface`, `type Formatter[T any] interface`.

**Why:** Go conventions mandate concise, idiomatic `-er` naming for single- or few-method interfaces representing behavior. Suffixing with `Interface` is an anti-pattern imported from other languages and strictly prohibited in this repository.

---

## Uppercase ID Acronym in Identifiers — TOTAL BAN

🔴 **NEVER use all-caps `ID` in variable names, struct fields, method names, or function parameters.**

Forbidden:
- ❌ `UserID`, `OrderID`, `AccountID`, `TraceID`, `ID`
- ❌ `GetID()`, `SetID()`, `traceID`

Allowed work:
- ✅ PascalCase `Id`: `UserId`, `OrderId`, `AccountId`, `TraceId`, `Id`
- ✅ camelCase `id`: `userId`, `orderId`, `accountId`, `traceId`, `id`

**Why:** Acronym casing must be normalized to `Id` in PascalCase and `id` in camelCase across all languages to eliminate capitalization inconsistencies and pass repository naming linters.

---

## Boolean Fields Without Positive Prefixes — TOTAL BAN

🔴 **NEVER define boolean fields, variables, or properties without an explicit positive prefix (`is`, `has`, `should`, `can`).**

Forbidden:
- ❌ `Active bool`, `Success bool`, `Match bool`, `Ready bool`
- ❌ `active: boolean`, `success: boolean`

Allowed work:
- ✅ `IsActive bool`, `IsSuccess bool`, `HasMatch bool`, `IsReady bool`
- ✅ `isActive: boolean`, `isSuccess: boolean`

**Why:** Bare boolean identifiers violate the repository's positive-polarity naming convention and impair readability in conditional guard clauses.

---

## Noisy Passing Quality Gate Output in CI Runners — TOTAL BAN

🔴 **NEVER flood developer terminals with stdout/stderr logs from passing quality gates in local CI test runners.**

Forbidden:
- ❌ Dumping passing command outputs to the console when all gates succeed.
- ❌ Interleaving asynchronous stdout streams from parallel workers across terminal lines.

Allowed work:
- ✅ Real-time single-line status ticker for completion progress (`[ 1/21] ✅ [PASS] <Gate> (<duration>s)`).
- ✅ Selective log suppression: print stdout/stderr ONLY for gates that exit with a non-zero status code or timeout.
- ✅ Full verbose logs emitted ONLY when the user explicitly passes the `--all` (`-a`) flag.

---

## High-Level Result/Error Wrappers in Enum Packages — TOTAL BAN

🔴 **NEVER import `pkg/result` or `pkg/errtype` inside Go enum packages under `pkg/enum/`.**

Forbidden:
- ❌ `import "coding-guidelines/common/pkg/result"` in `pkg/enum/**`
- ❌ `type Result = result.Wrap[Variant]` in enum packages
- ❌ `func Parse(s string) Result` returning `result.Wrap[Variant]` from leaf enum packages

Allowed work:
- ✅ Enum packages must be foundational leaf packages returning `(Variant, bool)`.
- ✅ Provide `ParseOrZero`, `ParseOrInvalid`, or `ParseOrUnknown` returning fallback variant.
- ✅ Higher-level consumer packages (e.g. `pkg/fileutil`) may wrap enum values into `result.Wrap` if their APIs require it.

**Why:** `pkg/result` imports `pkg/appfault`, and `pkg/appfault` imports enums (`severitytype`, `prioritytype`). When enum packages import `result`, it creates an unresolvable circular import cycle (`appfault` -> `enum` -> `result` -> `appfault`).

---

## Executing Staged Prompts Without User Confirmation — TOTAL BAN

🔴 **NEVER execute, implement, or begin coding on follow-up tasks when running the Conversation Log & Context Wrapper workflow.**

Forbidden:
- ❌ Commencing execution on rewritten follow-up instructions staged in `prompts/`.
- ❌ Modifying codebase files, generating features, or running migrations before the user reviews the staging report.

Allowed work:
- ✅ Persist conversation history to `conversation/`.
- ✅ Rewrite and stage instructions into `prompts/NNN-<slug>.md`.
- ✅ Register staged prompts in `01-prompts/01-prompt-library-setup/01-prompt-library-setup.md`.
- ✅ Acknowledge with "Understood - staging only, not executing." and halt until explicit user "go" confirmation.

**Why:** Staging must remain strictly decoupled from execution to allow human-in-the-loop review of rewritten specifications and guard against misaligned automated actions.

---

## Writing to `memories/` Instead of `.ai-memory/memory/` — TOTAL BAN

🔴 **NEVER create or write files to a root `memories/` directory.**

Forbidden:
- ❌ Writing memory logs, learned rules, or session indices to `memories/`.

Allowed work:
- ✅ All project memory, learned patterns, standards, and indices MUST reside strictly under `.ai-memory/memory/`.

**Why:** Creating divergent memory directories fractures institutional knowledge and prevents downstream tooling and installers from finding project context.

---

## Consolidating or Summarizing Detailed Architectural Specs — TOTAL BAN

🔴 **NEVER consolidate, summarize, abbreviate, or reduce detailed architectural specifications or domain specs (e.g. `02-spec/21-app/`).**

Forbidden:
- ❌ Summarizing detailed specs into high-level bullet points.
- ❌ Truncating code contracts, database schema definitions, or error structures during consolidation.

Allowed work:
- ✅ Ephemeral task checklists and routine milestone plans may be consolidated into milestone summaries.
- ✅ Detailed specifications, domain rules, and non-negotiable requirements MUST be preserved with 100% fidelity and full granularity.

**Why:** Lossy consolidation destroys technical contracts, edge-case definitions, and precision required by autonomous agents.

---

## Uppercase Root `README.md` File — TOTAL BAN

🔴 **NEVER create or permit an uppercase `README.md` file at repository root.**

Forbidden:
- ❌ Uppercase `README.md` or mixed-case `ReadMe.md`.

Allowed work:
- ✅ Root readme MUST be strictly lowercase `readme.md`.
- ✅ If an uppercase variant is detected, rename immediately to lowercase `readme.md`.

---

## Running Tests Without Owner Explicit Command — TOTAL BAN

🔴 **NEVER run unit tests, test suites, or CI test jobs (`go test`, `pytest`, `npm test`, or test jobs in CI runners) during standard development tasks or prompt execution unless explicitly commanded by the repository owner.**

Forbidden:
- ❌ Executing `go test`, `pytest`, `npm test`, or `cargo test` during standard development tasks, prompt executions, coding guideline fixes, or refactoring loops.
- ❌ Running `python 03-ai-scripts/06-cicd-local-runner.py` without `--no-tests` during standard development; always pass `--no-tests`.
- ❌ Adding automatic test execution steps to non-release workflows or prompts.

Allowed work:
- ✅ Run tests when the repository owner explicitly requests it in their prompt (e.g., "run tests", "execute unit tests", "fix failing tests").
- ✅ **ALL CI/CD Fix Workflows (`ci-cd-fix`, `16-ci-cd/*`):** MUST run all unit test suites, integration tests, linters, and quality gates properly (`python 03-ai-scripts/06-cicd-local-runner.py`) to diagnose, surface, and repair pipeline failures. Skipping tests with `--no-tests` in CI/CD fix tasks is strictly prohibited.
- ✅ **ALL Release Workflows (`release-management`, `release-orchestrator`, `01`, `03`, `07`, `16-ci-cd/04`):** MUST run all unit test suites (`--run-tests`) and verify 100% green passing before cutting any release.
- ✅ Always use `--no-tests` (or `--skip-tests`) when running standard routine development quality gate checks (`06-cicd-local-runner.py`) unless running CI/CD fixes, release ceremonies, or explicitly instructed by the owner.

**Why:** Unit test suites can be slow, resource-heavy, and disruptive during rapid iterative development loops. Running tests without explicit owner authorization wastes resources. Quality gates in standard turns focus on static analysis, linting, and structural integrity.

---

## Test Inventory & Atomic Recent File Changes Locking Mandate

🔴 **NEVER record or modify recent file change logs without cross-platform atomic file locking, and NEVER bypass `.ai-memory/test-inventory.json`.**

Forbidden:
- ❌ Writing directly to `.ai-memory/temp/recent-file-changes.json` without acquiring `.ai-memory/temp/recent-file-changes.lock`.
- ❌ Failing to release the lock or failing to handle stale locks properly.
- ❌ Guessing or manually hard-coding test file relationships without checking `.ai-memory/test-inventory.json`.

Allowed work:
- ✅ Use `python 03-ai-scripts/33-test-inventory-generator.py --record <relative-path>...` to atomically record modified files and resolve associated tests.
- ✅ Maintain and synchronize `.ai-memory/test-inventory.json` when adding, moving, or deleting test files by running `python 03-ai-scripts/33-test-inventory-generator.py`.
- ✅ Ensure all recorded paths are distinct, lowercase, and strictly relative to the repository root.

**Why:** Concurrent multi-agent orchestration and asynchronous script runs will corrupt `recent-file-changes.json` if writes are uncoordinated. Centralized test inventory mapping guarantees reproducible test discovery when an authorized release or targeted test fix is executed.

---

## Running Full CI/CD Runner During Routine Development or Coding Guideline Turns — TOTAL BAN

🔴 **NEVER run `python 03-ai-scripts/06-cicd-local-runner.py` during routine development tasks, coding guideline fixes (`15-cg-execute/*`), micro-loops, or sub-agent turns.**

Forbidden:
- ❌ Running `python 03-ai-scripts/06-cicd-local-runner.py` (with or without `--no-tests`) during routine task execution loops, coding standard audits, single-file refactoring, or micro-batches.
- ❌ Re-running the heavy 28-38 gate pipeline repeatedly for routine edits, wasting minutes across unrelated files and packages.
- ❌ Using the full pipeline runner to verify a single guideline edit (e.g. nested-if or boolean condition) when a targeted linter is available.

Allowed work:
- ✅ Run targeted file-level linters/autofixers directly on the modified file(s) (e.g., `python linter-scripts/check-nested-ifs.py <file>`, `python 03-ai-scripts/08-naming-autofixer.py <file>`, `python linter-scripts/check-boolean-guidelines.py <file>`).
- ✅ **Owner Explicit Command:** Run the runner if and only if the repository owner explicitly requests running the pipeline.
- ✅ **CI/CD Fix Tasks (`ci-cd-fix`, `16-ci-cd/*`):** May run `python 03-ai-scripts/06-cicd-local-runner.py` because the primary goal of those tasks is specifically repairing CI/CD infrastructure.
- ✅ **Release Ceremonies (`release-orchestrator`, `01`, `03`, `07`, `16-ci-cd/04`):** Run `python 03-ai-scripts/06-cicd-local-runner.py --run-tests` as the mandatory final pre-release gate before cutting a release.

**Why:** The local CI/CD runner runs up to 38 segments (linters, cross-OS compilation, snapshot builds, web builds) across the entire codebase. Executing this massive suite on every micro-turn or coding guideline edit causes immense latency, hits unrelated files, and wastes substantial developer and compute time.

---

## Committing Isolated 1-2 Plan/Doc Files Piecemeal — TOTAL BAN

🔴 **NEVER make piecemeal commits containing only 1-2 isolated markdown plan or doc files without code changes, and NEVER make repetitive single-file micro-commits.**

Forbidden:
- ❌ Committing a single completed plan markdown file by itself without the underlying work.
- ❌ Creating dozens of tiny 2-file commits for individual guideline prompts.
- ❌ Polluting git history with fragmented documentation-only commits.

Allowed work:
- ✅ Commit all modified files and plans together in a single, well-scoped atomic commit.
- ✅ Stage related source code, test files, and plans as a single unit of work.

**Why:** Piecemeal 1-2 file commits pollute git commit logs, make git history difficult to navigate, and separate documentation updates from their related functional work.

---

## Committing Without Immediate Git Push — TOTAL BAN

🔴 **NEVER leave commits unpushed on local branches. Anytime a commit is created, it MUST be pushed to GitHub immediately.**

Forbidden:
- ❌ Creating git commits and leaving them unpushed across conversational turns.
- ❌ Accumulating local commits without pushing to the remote repository.

Allowed work:
- ✅ Always execute `git push origin <branch>` immediately after creating any commit.
- ✅ Ensure local and remote `main` branch heads remain 100% synchronized at all times.

**Why:** Unpushed commits create discrepancies between local working states and remote CI/CD / GitHub Desktop viewers, risking sync conflicts and lost progress.

---

## Running Builds During Routine Guideline Turns — TOTAL BAN

🔴 **NEVER run full builds (`npm run build`, `go build ./...`) or packaging suites during routine coding guideline turns unless explicitly commanded by the repository owner.**

Forbidden:
- ❌ Running `npm run build` or `go build` during routine styling, naming, or guideline verification turns.
- ❌ Triggering expensive bundle transforms and minifications on routine audits.

Allowed work:
- ✅ Run targeted linters (`check-newline-styling.py`, `check-boolean-guidelines.py`, etc.) for fast validation.
- ✅ Run builds ONLY when explicitly requested or at the final release stage (`01-prompts/16-ci-cd/04-ci-cd-fix-with-release.md`).

**Why:** Running heavy frontend and backend builds repeatedly slows down feedback loops and consumes significant CPU/IO resources.

---

## Scattered Inline Structs and Ad-Hoc Generic Result Envelopes — TOTAL BAN

🔴 **NEVER declare domain payload structs or repeated generic Result envelopes inline in implementation files (`importer.go`, `store.go`, `service.go`).**

Forbidden:
- ❌ Declaring unexported domain structs (`type scheduleExportBundle struct { ... }`) inside implementation files.
- ❌ Repeating raw generic Result declarations (`result.ResultSlice[scheduleExportBundle]`, `result.ResultMap[string, []User]`) ad-hoc across multiple function signatures.
- ❌ Scattering type declarations across multiple implementation files instead of centralizing them in `types.go`.

Allowed work:
- ✅ Define all domain payload structs (`ScheduleExportBundle`, `PluginSummary`) in a dedicated `types.go` file within the package.
- ✅ Define single reusable Result type aliases (`type ScheduleExportBundleResult = result.ResultSlice[ScheduleExportBundle]`, `type PluginSummaryResult = result.Wrap[PluginSummary]`) in `types.go` to be reused everywhere.
- ✅ Use the single reusable named type alias in all function signatures and callers.

**Why:** Scattering ad-hoc inline structs makes them inaccessible across package boundaries, and repeating complex generic instantiations creates severe code churn, bloats call sites, and violates `02-spec/02-coding-guidelines/01-cross-language/27-types-folder-convention.md`.

---

## Replacing `isDefined` with `!isEmpty` or Inverted Negatives — TOTAL BAN

🔴 **NEVER invert negative emptiness checks (e.g., `!isEmpty`, `!is_empty`) to assert defined existence. Positive definition checks MUST use `isDefined`.**

Forbidden:
- ❌ `if !isEmpty(item) {`
- ❌ `if !map.isEmpty() {`
- ❌ `if !str.isEmpty() {`

Allowed work:
- ✅ Positive checks: `if isDefined {`
- ✅ Positive map lookup: `if val, isDefined := m[key]; isDefined {`
- ✅ Positive existence methods: `if hasKey {`, `if contains {`

**Why:** User explicitly commanded: *"IsDefined should be used instead of using !isEmpty , please make it clear in every prompt clearly and revert the orginal name as you have stated before"*. Inverted negative checks violate single-polarity principles and confuse reading flow.

---

## GitHub Actions CI Artifact Uploads (`actions/upload-artifact`) — TOTAL BAN

🔴 **NEVER upload routine build artifacts, test reports, Playwright traces, coverage files, drift reports, or logs to GitHub Actions storage.**

Forbidden:
- ❌ Using `actions/upload-artifact` in GitHub Actions workflows for test reports, build outputs, or logs.
- ❌ Uploading playwright-report, coverage-report, or SARIF files to GitHub Actions artifact storage.

Allowed work:
- ✅ Stream test outputs, summaries, and drift reports directly to `$GITHUB_STEP_SUMMARY` (renders natively with 0 storage cost).
- ✅ Output diagnostic failures via workflow annotations (`::error::`, `::warning::`) and standard stdout.
- ✅ Release distribution binaries attached directly to GitHub Releases via `gh release create` / `gh release upload` (exempt from Actions quota).

**Why:** Free-tier GitHub accounts have a strict 0.5 GB quota across all account repositories. Routine artifact uploads rapidly exhaust this limit and block repository CI pipelines.

---

## Non-`er` Go Interface Suffixes — TOTAL BAN

🔴 **NEVER name Go interfaces with non-`er` suffixes (e.g. `Creator`, `Descriptor`) or `Interface`. All Go interfaces MUST end in `er`.**

Forbidden:
- ❌ `type ViewCreator interface {`
- ❌ `type DatabaseDescriptor interface {`
- ❌ `type QueryInterface interface {`

Allowed work:
- ✅ `type ViewManager interface {`
- ✅ `type SqlExecutor interface {`
- ✅ `type DbExecutor interface {`

**Why:** Go conventions and repository linter `check-interface-naming.py` strictly mandate the `er` suffix for all interface definitions. Suffixes like `Creator` (ending in `or`) trigger hard pre-commit failures.

---

## Literal Drive-Letter or Absolute URIs in Test Code — TOTAL BAN

🔴 **NEVER write literal `file:///` followed by drive letters or absolute filesystem paths in test files or regex fixtures without concatenation or masking.**

Forbidden:
- ❌ Literal test string: `target := "file:///" + "c:/path/to/file"` (or unmasked literal `file:///` + drive letter)
- ❌ Literal regex test: `assert.Contains(t, out, "C:" + "\\Users\\...")`

Allowed work:
- ✅ String concatenation: `"file:" + "///" + "c:/path/to/file"`
- ✅ Path composition: `filepath.Join("var", "test")`

**Why:** Repository linter `check-relative-paths.py` statically scans all tracked files for absolute filesystem paths and file URIs. Literal test fixtures trigger false-positive pre-commit failures.

---

## Writing Memory Without Inspecting Last 30 Commits — TOTAL BAN

🔴 **NEVER author or update memory files without first executing `git log -n 30 --oneline` (and `git log -n 30 --stat` where needed) to inspect the last 30 commits.**

Forbidden:
- ❌ Writing memory based purely on chat context or ephemeral memory without checking `git log`.
- ❌ Guessing what code was modified, what fixes were committed, or what directives were applied.

Allowed work:
- ✅ Execute `git log -n 30 --oneline` at Pre-Flight Step 0 before opening or modifying any memory files.
- ✅ Explicitly capture recent commit changes, resolved issues, and directives in Phase 1 Internal Session Audit.

**Why:** Ephemeral chat context loses granularity across turns. Ground truth exists in the repository commit history.

---

## Writing Memory Without Verifying the Recent 20-Task Register — TOTAL BAN

🔴 **NEVER author or update memory files without auditing `.ai-memory/plans/01-index.md` and verifying the Recent Completed Tasks Register (last 20 tasks).**

Forbidden:
- ❌ Leaving the Recent Completed Tasks Register unmaintained or out of sync with `05-changes-history/`.
- ❌ Omitting recent completed tasks from the register.
- ❌ Writing new memory without cross-referencing completed work in `plans/01-index.md` and `what-to-read.md`.

Allowed work:
- ✅ Maintain the bounded rolling window of the last 20 tasks in `.ai-memory/plans/01-index.md`.
- ✅ Cross-reference completed tasks from `05-changes-history/` with dates, titles, and paths.

**Why:** The compact 20-task register guarantees that subsequent AI turns immediately understand recent progress without scanning hundreds of historical records.

---

## Uppercase README.md Filename — TOTAL BAN

🔴 **NEVER allow an uppercase `README.md` to exist at the repository root or in source directories. The root readme MUST be strictly named lowercase `readme.md`.**

Forbidden:
- ❌ Naming the file `README.md` or `Readme.md`.
- ❌ Allowing git or operating system casing ambiguity to introduce uppercase letters in documentation filenames.

Allowed work:
- ✅ Strictly name root file `readme.md`.
- ✅ If an uppercase variant is detected, rename immediately, delete the uppercase file, commit, and push.

**Why:** Strict lowercase naming is a mandatory meta-repo standard across all documentation, tools, and scripts.

---

## Consolidating or Shrinking Detailed Specifications — TOTAL BAN

🔴 **NEVER consolidate, summarize, resume, or shrink detailed specifications, architectural designs, domain models (e.g. `02-spec/21-app/`), or complex requirement documents.**

Forbidden:
- ❌ Summarizing detailed specs into high-level bullet points to save space.
- ❌ Deleting concrete examples, error tables, or domain rules during documentation refactoring.

Allowed work:
- ✅ Consolidate ephemeral or routine completed simple tasks into master changelogs to prevent file bloat.
- ✅ Preserve all detailed specifications with 100% fidelity, exact wording, and full granularity.

**Why:** Architectural specs are canonical contracts. Summarizing or shrinking them destroys domain nuance and leads to hallucinations.

