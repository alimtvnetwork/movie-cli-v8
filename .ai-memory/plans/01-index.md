# Plans Index

Master directory of architectural and execution plans.

## Pending Plans

- [02-slides-system-overhaul.md](pending/02-slides-system-overhaul.md): Full slides deck UI and system overhaul.
- [04-guideline-prompt-and-installer-upgrade.md](pending/04-guideline-prompt-and-installer-upgrade.md): Guideline prompt and installer enhancements.
- [09-update-prompts-and-release.md](pending/09-update-prompts-and-release.md): Update prompts and release lifecycle (deferred under WOR policy).
- [11-code-red-refactor-remediation.md](pending/11-code-red-refactor-remediation.md): Remediate Code Red enum, boolean, and query wrapper violations across the codebase.
## Completed Plans

- [17-invalid-api-key-reprompt-and-release.md](completed/17-invalid-api-key-reprompt-and-release.md): Proactive TMDb API key verification, interactive re-prompt loop, and minor release.
- [16-advanced-installer-and-release-process-parity.md](completed/16-advanced-installer-and-release-process-parity.md): Advanced installer and release process parity following GitMap (dual-mode eval auto-source, versioned repo discovery, self-uninstall command, and static smoke linter).
- [15-gitmap-installer-and-release-parity.md](completed/15-gitmap-installer-and-release-parity.md): Full installer and release process parity with GitMap (deploy-manifest.json, init.ps1/init.sh, run.sh, smoke-installer.py, CI/CD dryrun validation).
- [14-improve-installer-and-release-process.md](completed/14-improve-installer-and-release-process.md): Overhaul binary installers, quick one-liners, uninstaller suite, and release workflow dry-run contract verification following GitMap.

- [12-movie-cli-ui-trash-error-and-reset-overhaul.md](completed/12-movie-cli-ui-trash-error-and-reset-overhaul.md): Movie CLI UI, staged trash-bin deletion, AppFault error management, TMDB image handling, reset command, and colorful terminal help.
- [02-multi-repo-folder-structure-and-gitmap-fixes.md](completed/02-multi-repo-folder-structure-and-gitmap-fixes.md): Multi-repo folder structure migration and gitmap AGY pipeline fixes.
- [01-repository-infrastructure-cicd-and-consolidation.md](completed/01-repository-infrastructure-cicd-and-consolidation.md): Repository hygiene, encoding normalization, lowercase conventions, AI scripts `<details>` documentation, CI/CD quality automation, and plan memory consolidation.
- [02-appfault-result-monad-and-verification-systems.md](completed/02-appfault-result-monad-and-verification-systems.md): Go `*appfault.AppError` standard, `Result[T]` generic containers, dynamic type conversions, deterministic map sorting, and `ReflectSetTo` fast path.
- [03-fileutil-pathinfo-and-enum-architecture.md](completed/03-fileutil-pathinfo-and-enum-architecture.md): Modular file operations, cross-platform temp resolution, .NET-style `PathInfo`/`FolderInfo`/`FileInfo` objects, concurrency locking, and 1:1 modular `BaseEnumer` enums.
- [04-applogger-taxonomy-streaming-and-task-db.md](completed/04-applogger-taxonomy-streaming-and-task-db.md): Structured AppLogger subsystem, split SQLite logging (`logs.db` + `tasks/<task-id>.db`), rotating file sink, generic `LazyOnce`, errcmd streaming, named writers, and typed streamers.
- [12-regex-centralization-and-generic-dbengine.md](completed/12-regex-centralization-and-generic-dbengine.md): Lazy regex centralization, core regex harvesting, generic dbengine wrapper, and python struct code generator.
- [13-os-enum-integration-and-generator-upgrade.md](completed/13-os-enum-integration-and-generator-upgrade.md): OS enum integration, deep OS version detection, and enum generator upgrade.
- [14-cicd-fix-with-release.md](completed/14-cicd-fix-with-release.md): Release-triggered CI/CD fix loop, full test suite verification, and automated minor release publication (v6.41.0).

## Recent Completed Tasks Register (Last 20 Tasks)

Compact registry of the last 20 tasks, transaction logs, and milestone plans. Every new memory write and planning loop references this register to maintain continuous context without guessing:

1. **Task 26:** [Write Prompts 30-Commit Git Audit & Recent 20-Task Register](05-changes-history/26-write-prompts-git-audit-and-recent-tasks-register/01-transaction-log.md) (2026-09-13) — Write prompts upgrade (v2.2.0), mandatory 30-commit git history audit, recent 20-task tracking register, and 19-box verification checklist.
2. **Task 25:** [Regex Centralization, Generic DbEngine & isDefined Standard](05-changes-history/25-regex-harvesting-dbengine-and-isdefined-standard/01-transaction-log.md) (2026-09-13) — 80+ regexes from 03-aukgo/core, redistributable pkg/dbengine, isDefined positive convention, zero-storage Actions mandate.
3. **Task 24:** [Task Retention, errcmd Streaming, Atomic Fileutil & ApiManager](05-changes-history/24-task-retention-errcmd-streaming-atomic-fileutil-lazyonce-apimanager/01-transaction-log.md) (2026-09-09) — Task retention pruning, line-by-line live streaming, atomic file writing, ApiManager.
4. **Task 23:** [SQLite Task DB Auto-Migration and Repair](05-changes-history/23-sqlite-task-db-auto-migration-and-repair/01-transaction-log.md) (2026-09-09) — Dynamic schema audit, column repair, and index creation for SQLite task databases.
5. **Task 22:** [Split SQLite Logging, Rotating File Sink & LazyOnce](05-changes-history/22-split-sqlite-logging-rotating-lazyonce-errcmd/01-transaction-log.md) (2026-09-09) — Split logs (`logs.db` + `tasks/<task-id>.db`), rotating file sink with gzip, and thread-safe LazyOnce.
6. **Task 21:** [Completed Plans Consolidation & Safety Backups](05-changes-history/21-completed-plans-consolidation/01-transaction-log.md) (2026-09-09) — Consolidation of plans into 6 core milestones with git safety branch backup.
7. **Task 20:** [Comprehensive Tests for Enums and BaseEnumer](05-changes-history/20-comprehensive-tests-for-enums-and-baseenumer/01-transaction-log.md) (2026-09-09) — 100% test coverage for enums, boundary methods, and baseenumer helpers.
8. **Task 19:** [Leaf Enums, BaseEnumer Helpers & Cycle Elimination](05-changes-history/19-leaf-enums-baseenumer-helpers-and-cycle-elimination/01-transaction-log.md) (2026-09-09) — Leaf enum architecture, parse helpers, and elimination of import cycles.
9. **Task 18:** [Streamer & Writer Self-Passing Implementation](05-changes-history/18-streamer-and-writer-self-passing-research/01-transaction-log.md) (2026-09-04) — Injected write methods with self-passing receiver context and lock/unlock controls.
10. **Task 17:** [Parallel CI/CD Local Runner & Log Filtering](05-changes-history/17-parallel-cicd-local-runner-and-log-filtering/01-transaction-log.md) (2026-09-05) — Multi-threaded ThreadPoolExecutor worker groups and selective failure-only log filtering.
11. **Task 16:** [Boolean Prefixes, Id Standard & Writer Context](05-changes-history/16-boolean-prefixes-id-standard-and-writer-context/01-transaction-log.md) (2026-09-05) — Strict positive boolean naming, PascalCase/camelCase Id rules, and writer execution context.
12. **Task 15:** [Full Runnable Examples: Logger, Json & Streamer](05-changes-history/15-full-examples-logger-json-streamer/01-transaction-log.md) (2026-09-05) — End-to-end examples and CLI demonstrations for pluggable streamwriter and logger packages.
13. **Task 14:** [JsonResult Pure Bytes & Payload Extension](05-changes-history/14-jsonresult-pure-bytes-and-payload-extension/01-transaction-log.md) (2026-09-04) — Minimalist JsonResult byte slice container and JsonPayloadResult typed extension.
14. **Task 13:** [Non-Generic JsonResult Architecture](05-changes-history/13-non-generic-jsonresult/01-transaction-log.md) (2026-09-04) — Transition to non-generic JsonResult eliminating type parameter pollution across call sites.
15. **Task 12:** [Json Naming & Any-Based JsonSource Ingestion](05-changes-history/12-json-naming-and-any-based-jsonsource/01-transaction-log.md) (2026-09-04) — Standardized Json naming convention and polymorphic JsonSource ingestion.
16. **Task 11:** [JsonResult Multi-Source Creation](05-changes-history/11-jsonresult-multi-source-creation/01-transaction-log.md) (2026-09-04) — Constructors for bytes, strings, streams, payloads, and round-trip casting.
17. **Task 10:** [WrappedBytes Interface & Status Flags](05-changes-history/10-wrapped-bytes-interface-and-json-result/01-transaction-log.md) (2026-09-04) — WrappedBytes contract with Value(), Error(), and status inspection accessors.
18. **Task 09:** [Writer Locker Integration & Avoiding Interfacer](05-changes-history/09-writer-locker-and-avoiding-interfacer/01-transaction-log.md) (2026-09-04) — ReentrantMutex deadlock prevention and sync.Locker implementation on writers.
19. **Task 08:** [Idiomatic -er Go Interface Naming Standard](05-changes-history/08-idiomatic-er-interface-naming/01-transaction-log.md) (2026-09-04) — Banned Interface suffix and enforced mandatory Go -er naming pattern across all packages.
20. **Task 07:** [Bytes[T] Monadic Wrapper & *appfault.AppError Standard](05-changes-history/07-bytes-wrapper-and-apperror-standard/01-transaction-log.md) (2026-09-04) — Replaced raw ([]byte, error) with monadic Bytes[T] and standardized structured AppError returns.

- [01-rename-lovable-to-ai-memory.md](completed/01-rename-lovable-to-ai-memory.md): Renamed .lovable to .ai-memory everywhere.
- [03-quiz-feature-implementation.md](completed/03-quiz-feature-implementation.md): Implemented quiz feature frontend and backend scaffold.
- [05-wp-exam-migration-and-quiz-feature.md](completed/05-wp-exam-migration-and-quiz-feature.md): Migrate guidelines and add quiz feature to WP exam.
- [06-wp-exam-migration-retry.md](completed/06-wp-exam-migration-retry.md): Retry of WP exam migration and quiz feature due to subagent hallucinations.
- [07-movie-cli-migration-and-optimization.md](pending/07-movie-cli-migration-and-optimization.md): Migration and DRY optimization of movie-cli-v8.
