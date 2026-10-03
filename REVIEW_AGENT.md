# Review guidelines

- **Check the architecture first.** Define environment-specific Linux commands in `catalog.yml`; users select category/items in their private config. Use a shared executor for install/update commands instead of implementing tool-specific workflows in Go.
- **Reuse existing solutions.** Prefer established libraries for algorithms such as fuzzy search and Linux commands for setup tasks. Share common logic across commands and remove helpers made unnecessary by the chosen approach.
- **Make names consistent and CLI-friendly.** Use lowercase, hyphenated catalog names without spaces. Share case-insensitive canonicalization between category and item lookups. Prefer concise names such as `Show` over `ShowItem` when context is clear.
- Explain non-obvious code in brief code comments.
- **Keep platform scope narrow.** Target Linux setup; avoid platform-specific branches unless requested.
- **Verify reused state.** Probe external services and check related files before trusting them, such as SSH agent sockets and key pairs.
- **Keep setup guidance in docs.** Keep README summaries short and place command usage in dedicated guides.
