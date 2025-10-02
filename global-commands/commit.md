# Commit Command

Automatic checkpoint commits with conventional commit messages and intelligent staging.

## Role

You are an **Automatic Commit Specialist** that handles complete commit workflows using conventional commit standards. You stage modified files, analyze changes, classify them into conventional commit types (feat, fix, refactor, docs, test, chore, revert), and execute commits automatically.

## Usage

```
/commit [--amend] [--type <type>]
```

**Purpose**: One-command checkpoint commits with conventional commit messages - stage changes, classify modifications, generate proper commit message, and commit automatically.

## Conventional Commit Types

**Classification Rules:**
- **feat**: New features, capabilities, or user-facing functionality
- **fix**: Bug fixes, error corrections, or issue resolutions
- **refactor**: Code restructuring without changing functionality
- **docs**: Documentation changes, README updates, comments
- **test**: Adding or modifying tests, test utilities
- **chore**: Maintenance tasks, dependency updates, build changes, tooling
- **revert**: Reverting previous commits

## Process

### Step 1: Change Detection and Staging
- Run `git status` to identify all modified files
- **Automatic Staging**: `git add .` to stage all changes
- Show what's being staged for transparency
- Handle new files, modifications, deletions, and renames

### Step 2: Conventional Commit Classification
**Analyze changes to determine commit type:**

- **feat**: New functions, classes, components, features, or capabilities
- **fix**: Error handling, bug corrections, broken functionality repairs
- **refactor**: Code reorganization, performance improvements, cleanup without new features
- **docs**: README, comments, documentation files, code documentation
- **test**: Test files (.test., .spec.), testing utilities, test configurations
- **chore**: package.json, build configs, tooling, dependencies, formatting

**Auto-detect primary type** based on:
- File types modified (source vs test vs docs vs config)
- Nature of changes (new code vs fixes vs restructuring)
- Scope of impact (user-facing vs internal vs infrastructure)

### Step 3: Conventional Message Generation

**Format**: `type(scope): description`

**Message Structure:**
- **Type**: Auto-classified from analysis above
- **Scope**: Optional - component/module affected (e.g., `auth`, `api`, `ui`)
- **Description**: Concise imperative summary (50 chars max)

**Examples:**
- `feat(auth): add password reset functionality`
- `fix(api): handle null response in user endpoint`
- `refactor(utils): extract common validation logic`
- `docs: update API documentation for new endpoints`
- `test(auth): add unit tests for login validation`
- `chore: update dependencies to latest versions`

### Step 4: Automatic Commit Execution
- Execute: `git commit -m "conventional_message"`
- Display commit type classification reasoning
- Show final commit hash and message
- Confirm successful conventional commit

### Step 5: Multi-Type Change Handling
**When changes span multiple types:**
- Prioritize most significant change type
- Use primary type with comprehensive description
- Example: New feature with docs → `feat: add user management with documentation`
- Large mixed changes: Suggest breaking into focused commits

## Implementation Logic

**Type Detection Algorithm:**
```
if (new_functionality_added): → feat
elif (bugs_fixed or errors_corrected): → fix  
elif (code_restructured_no_new_features): → refactor
elif (only_documentation_changed): → docs
elif (only_tests_modified): → test
elif (build_deps_tooling_maintenance): → chore
elif (reverting_previous_commit): → revert
else: → chore (default fallback)
```

**Scope Detection:**
- Analyze modified file paths for common patterns
- Extract module/component names from imports or directory structure
- Use most frequently affected component as scope
- Omit scope if changes are too broad or unclear

**Description Generation:**
- Imperative mood: "add", "fix", "update", "remove"
- Concise but descriptive
- Focus on what, not how
- Include key affected functionality

## Expected Behavior

When El Presidente runs `/commit`:

1. **"Staging all changes..."** - Shows files being added
2. **"Analyzing changes... classified as: feat"** - Shows type reasoning
3. **"Generated message: feat(api): add user authentication endpoints"**
4. **"Committed abc1234: feat(api): add user authentication endpoints"**
5. **Ready for next development cycle**

**Perfect Conventional Commits:**
- Automatic type classification
- Consistent message format
- Clean commit history
- No manual formatting needed
- Follows conventional commit specification exactly