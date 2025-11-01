## [v0.0.2 - 2025-11-01]

### Added
- Introduced Pro Mode coding challenges with Monaco editor tooling, hint flow, and `/api/prochallenge` endpoints that execute hidden Go test suites for XP and coin rewards.
- Seeded `data/pro_challenges.json` with advanced and medium-difficulty scenarios (including new clean-code exercises) plus private protest suites to validate submissions.
- Created a reusable `ModeCard` component and new dashboard Coding panel so Learn, Coding, and Typing each present focused stats and actions.

### Changed
- Expanded Pro Mode filters with topic and difficulty selectors and surfaced current selections alongside coin/XP totals.
- Updated dashboard layout to align cards in a responsive grid and moved category selection inside the Learn Mode card.

## [v0.0.1 - Unreleased]

### Added
- Configured the frontend as a Progressive Web App (PWA) via Vite’s PWA plugin, including manifest metadata, service worker registration, and an app icon so deployments support offline usage and install prompts.
- Added a reusable category selector pill that drives Learn and Typing flows with `Any` and `Random` options, keeping styling consistent across screens.
- Expanded `data/lessons.json` with a broad set of clean-code, DevOps, Kubernetes, cloud, security, and architecture lessons to reduce quiz repetition and deepen coverage.
- Introduced responsive CSS breakpoints to stack dashboard controls, scale typography, and adjust layout for tablet and phone viewports.
- Surfaced typing-mode streak/best on the dashboard and persisted those metrics alongside quiz streaks.

### Changed
- Updated frontend/backend lesson selection logic to respect session history and lower repetition.
- Tweaked dropdown styling to match the app’s badge aesthetics and maintain dark-theme contrast.
- Suppressed redundant Vite PWA warnings during development by adjusting plugin `devOptions`.
- Removed obsolete root-level Go entry point and module file to avoid confusion—backend code now lives solely under `backend/`.
- Reworked the dashboard into dedicated Learn and Typing panels for clearer stats and responsiveness improvements.
