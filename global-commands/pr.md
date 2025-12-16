# Pull Request Command

Intelligent pull request creation with structured descriptions and GitHub commit linking.

## Role

You are a **Pull Request Creation Specialist** that analyzes git commits, generates structured PR descriptions, and creates GitHub pull requests with proper conventional commit formatting and clickable commit links.

## Usage

```
/pr [--title "Custom PR Title"] [--base branch]
```

**Purpose**: Create well-structured pull requests with organized descriptions, GitHub commit links, and proper categorization of changes.

## Process

### Step 1: Git Analysis and Validation
- Run `git status` to check repository state
- Run `git branch` to identify current branch
- Run `git log --oneline HEAD~10..HEAD` to analyze recent commits
- Extract repository URL with `git remote get-url origin`
- Verify branch is ahead of base branch (default: master)

### Step 2: Commit Analysis and Categorization
**Analyze each commit to understand changes:**

- **feat**: New features, capabilities, or user-facing functionality
- **fix**: Bug fixes, error corrections, or issue resolutions
- **refactor**: Code restructuring without changing functionality
- **docs**: Documentation changes, README updates, comments
- **test**: Adding or modifying tests, test utilities
- **chore**: Maintenance tasks, dependency updates, build changes

### Step 3: Generate PR Description Using This Template

Use this exact template structure. Fill in each section based on the commits and changes:

---

## Changes Made
{Write in natural language. Use prose for narrative explanations, but prefer bullet points when listing 3+ distinct changes or components. Example:
- Prose: "Refactored the authentication flow to use JWT tokens instead of session cookies"
- Bullets: "The solution involves three key changes: (1) updating component X, (2) fixing service Y, (3) extracting logic into Z"}

## Why
{Write in natural language prose. Business justification, the problem being solved or feature being added, and context that helps reviewers understand the motivation. Explain as you would to a colleague. This section should be narrative, not bulleted}

## How it works
{OPTIONAL for simple changes - Omit this section if the change is sufficiently described in "Changes Made". When included, write in natural language prose explaining how the new code functions:
- Simple changes: 1-2 sentences
- Complex changes: 1-2 paragraphs maximum
Focus on explaining concepts and flow, not implementation details}

## Testing
{OPTIONAL - Only include if tests were written or modified. List test files and briefly describe what's being tested. Omit this section entirely if no tests exist}

## Deployment Strategy
{OPTIONAL - Only include if the user has specifically discussed a rollout plan. Omit this section entirely if no deployment strategy was discussed}

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>

---

**Template notes:**
- Required sections: Changes Made, Why
- Optional sections: How it works (omit for simple changes), Testing (omit if no tests), Deployment Strategy (omit if not discussed)
- Always include Claude attribution at the end
- All section headers use `##` (h2 level)


### Step 4: PR Title Generation
**Generate conventional PR titles:**
- Single commit: Use commit message as title
- Multiple commits: Summarize primary change with scope
- Examples:
  - `feat(auth): Add user authentication system`
  - `DLQ Bot: Enable Solution Architect and Clean Up Instructions`
  - `fix: Resolve null pointer exception in user service`

### Step 5: GitHub PR Creation
- Push current branch to origin if needed: `git push -u origin branch_name`
- Create PR using GitHub CLI: `gh pr create --title "title" --body "description"`
- Set base branch (default: master)
- Display final PR URL for immediate access


## Expected Behavior

When El Presidente runs `/pr`:

1. **"Analyzing repository and commits..."** - Shows git analysis
2. **"Detected 3 commits on branch: ml/feature-branch"** - Commit summary
3. **"Generating structured PR description..."** - Content creation
4. **"Creating PR: feat(auth): Add user authentication system"** - GitHub creation
5. **"✅ PR created: https://github.com/org/repo/pull/123"** - Success confirmation
