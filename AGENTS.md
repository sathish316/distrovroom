# Repository instructions

## Pull request reviews

- Read [REVIEW_AGENT.md](REVIEW_AGENT.md) and apply its guidelines when reviewing a PR.
- Reflect on the general principles and patterns in the user's review comments, update `REVIEW_AGENT.md` with concise, reusable guidelines, and include those updates in the same PR.
- Review the PR diff and relevant surrounding code for actionable bugs and regressions. Focus on code review; do not run tests unless requested.
- Post findings on the PR with priority, file/line references, impact, and a suggested fix. Prefer inline comments; if a pending review blocks posting, use a conversation comment without changing the user's pending review.
- Another agent will address the review comments you post. Wait for the user to confirm all review comments are addressed, then check the changes and merge the PR.
- Before merging, check the latest PR changes and reported CI status.
