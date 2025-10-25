## [v1.0.2 - Unreleased]

### Added
- Configured the frontend as a Progressive Web App (PWA) via Vite’s PWA plugin, including manifest metadata, service worker registration, and an app icon so deployments support offline usage and install prompts.
- Added a reusable category selector pill that drives Learn and Typing flows with `Any` and `Random` options, keeping styling consistent across screens.
- Expanded `data/lessons.json` with a broad set of clean-code, DevOps, Kubernetes, cloud, security, and architecture lessons to reduce quiz repetition and deepen coverage.
- Introduced responsive CSS breakpoints to stack dashboard controls, scale typography, and adjust layout for tablet and phone viewports.

### Changed
- Updated frontend/backend lesson selection logic to respect session history and lower repetition.
- Tweaked dropdown styling to match the app’s badge aesthetics and maintain dark-theme contrast.
- Suppressed redundant Vite PWA warnings during development by adjusting plugin `devOptions`.
- Removed obsolete root-level Go entry point and module file to avoid confusion—backend code now lives solely under `backend/`.
