# Review guidelines

- **Check the architecture first.** Define environment-specific Linux commands in `catalog.yml`; users select category/items in their private config. Group selected items by category in the private config. Run install commands from the catalog for one item or one category, in config order, instead of implementing tool-specific workflows in Go.
- **Reuse existing solutions.** Prefer established libraries for algorithms such as fuzzy search and Linux commands for setup tasks. Share common logic across commands and remove helpers made unnecessary by the chosen approach.
- **Make names consistent and CLI-friendly.** Use lowercase, hyphenated catalog names without spaces. Share case-insensitive canonicalization between category and item lookups. Prefer concise names such as `Show` over `ShowItem` when context is clear.
- Explain non-obvious code in brief code comments.
- **Keep platform scope narrow.** Target Linux setup; avoid platform-specific branches unless requested.
- **Keep catalog commands simple.** Use a concise Linux command when it covers the task. Render placeholders from supplied Cobra flags generically and fail when a required value is missing.
- **Keep command help self-explanatory.** Document the essential config path and CLI usage concisely; avoid duplicating verbose command guides.
