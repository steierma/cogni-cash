# 📜 Cogni-Cash Commit & Release Convention

We follow the **[Conventional Commits](https://www.conventionalcommits.org/)** standard and use automated versioning starting from **v1.4.0**.

## 🛠 Commit Message Format

Each commit message consists of a **type**, an optional **scope**, and a **subject**:

```text
<type>(<scope>): <subject>
```

### Types:
- `feat`: A new feature (corresponds to a **MINORE** version bump)
- `fix`: A bug fix (corresponds to a **PATCH** version bump)
- `chore`: Regular maintenance, dependency updates, etc.
- `docs`: Documentation changes
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `perf`: A code change that improves performance
- `test`: Adding missing tests or correcting existing tests

### Examples:
- `feat(ui): add dark mode support`
- `fix(auth): resolve session timeout issue`
- `chore(deps): update axios to v1.7`

---

## 🚀 How to Create a Release

We use `standard-version` to automate version bumping, changelog generation, and git tagging.

### 1. Prerequisites
Ensure you have Node.js installed and dependencies are fetched:
```bash
npm install
```

### 2. Performing the Release
When you are ready to cut a new version (e.g., jumping from v1.3.0 to v1.4.0):

```bash
# To automatically determine the next version based on commits:
npm run release

# To force a specific version (e.g., v1.4.0):
npm run release -- --release-as 1.4.0
```

### 3. What happens next?
1. `standard-version` updates the `version` in `package.json`.
2. It updates `CHANGELOG.md` with all commits since the last release.
3. It creates a local git commit and a tag (e.g., `v1.4.0`).
4. **You must push the tag to trigger the CI/CD pipeline:**

```bash
git push --follow-tags origin main
```

---

## 🏗 CI/CD Behavior
- **Pushing to `main` branch:** Builds and deploys images with the `:latest` tag to your **private internal** registry.
- **Pushing a tag (`v*`):** Builds images with that specific version tag (e.g., `:1.4.0`) to **both** internal and public registries.
