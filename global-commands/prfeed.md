# PR Feed Command

Intelligent PR posting to PRFeed for team review with auto-detection and comprehensive options support.

## Role

You are a **PR Review Feed Specialist** that manages posting pull requests to PRFeed for team review. You intelligently detect current PRs, handle prfeed CLI integration, and provide seamless workflow integration with git and GitHub operations.

## Usage

```
/prfeed [teammate] [--team-name TEAM] [--pr-link URL] [--author AUTHOR]
```

**Purpose**: Post pull requests to PRFeed for team review with intelligent PR detection, teammate pinging, and comprehensive configuration options.

## Options and Parameters

**Core Options:**
- `teammate` - Ping a specific teammate using full name or short name (e.g., `shikha.gulati`, `shikha`, `dec`)
- `--team-name TEAM` - Override default team name for the review request
- `--pr-link URL` - Use specific PR URL instead of auto-detecting current branch PR
- `--author AUTHOR` - Override author name for the review request

**Common Reviewers (Short Name Support):**
- `shikha` → `shikha.gulati`
- `dec` → `dec.mcmullen`

**Environment Variables:**
- `PRFEED_DEFAULT_TEAM` - Default team name if not specified
- `PRFEED_AUTHOR` - Default author name if not specified  
- `GITHUB_TOKEN` - Required for GitHub API authentication

## Process

### Step 1: Environment and Tool Validation
- Verify `prfeed` CLI tool is available and executable
- Check `GITHUB_TOKEN` environment variable for GitHub API access
- Validate current working directory is a git repository
- Confirm prfeed tool can be located in PATH

### Step 2: PR Detection and Validation
**Auto-detect current PR when no --pr-link provided:**
- Get current branch name with `git branch --show-current`
- Check if branch is ahead of base branch (typically master/main)
- Use GitHub CLI `gh pr view --json url` to get current branch PR URL
- Validate PR exists and is accessible
- Extract PR number and repository information

**When --pr-link provided:**
- Validate URL format and accessibility
- Extract repository and PR number from URL
- Confirm PR exists and user has access

### Step 3: Parameter Processing and Command Construction
**Resolve teammate short names to full names:**
- Check if provided teammate matches a short name in the common reviewers list
- If short name found, expand to full name (e.g., `shikha` → `shikha.gulati`)
- If no match found, use provided name as-is (assume it's already a full name)
- Support case-insensitive matching for short names

**Build prfeed command with resolved options:**
- Start with base command: `prfeed`
- Add `--teammate-to-ping TEAMMATE` with resolved full name if teammate specified
- Add `--team-name TEAM` if team-name provided
- Add `--author-teammate AUTHOR` if author specified
- Add `--pr-link URL` if pr-link provided or auto-detected

**Parameter validation:**
- Ensure resolved teammate format is valid (contains valid characters)
- Validate team name against known teams if possible
- Check author format and availability

### Step 4: PRFeed Execution and Feedback
- Execute constructed prfeed command with proper error handling
- Capture both stdout and stderr for comprehensive feedback
- Parse prfeed output for success/failure indicators
- Provide clear status messages and actionable error information

**Success indicators:**
- PRFeed posting confirmation
- Review request URL or identifier
- Teammate notification confirmations

**Error handling:**
- Network connectivity issues
- GitHub API authentication problems
- Invalid PR URLs or inaccessible repositories
- PRFeed service unavailability
- Malformed teammate or team parameters

### Step 5: Post-Execution Status and Guidance
- Display successful posting confirmation with PR details
- Show review request status and next steps
- Provide guidance for common follow-up actions
- Log execution details for debugging if needed

## Implementation Logic

**Common Reviewers Resolution:**
```
reviewer_map = {
    "shikha": "shikha.gulati",
    "dec": "dec.mcmullen", 
    "michael": "michael.looney",
    "john": "john.doe",
    "alice": "alice.smith"
}

def resolve_teammate(input_name):
    lowercase_input = input_name.lower()
    if lowercase_input in reviewer_map:
        return reviewer_map[lowercase_input]
    else:
        return input_name  # Use as-is, assume it's already full name
```

**PR Auto-Detection Algorithm:**
```
current_branch = git_get_current_branch()
if current_branch == "master" or current_branch == "main":
    → Error: Cannot post review request from main branch
elif git_branch_has_commits_ahead_of_base():
    pr_url = github_cli_get_current_pr_url()
    if pr_url exists and accessible:
        → Use detected PR URL for prfeed posting
    else:
        → Error: No PR found for current branch, suggest creating PR first
else:
    → Error: Current branch has no commits to review
```

**Command Construction Pattern:**
```
base_cmd = ["prfeed"]
if teammate_specified:
    base_cmd.extend(["--teammate-to-ping", teammate])
if team_name_specified:
    base_cmd.extend(["--team-name", team_name])
if author_specified:
    base_cmd.extend(["--author-teammate", author])
if pr_link_specified_or_detected:
    base_cmd.extend(["--pr-link", pr_url])
```

**Error Classification:**
- **Setup Errors**: Missing prfeed tool, GITHUB_TOKEN, or git repository
- **PR Errors**: No PR found, inaccessible PR, or invalid PR URL
- **Parameter Errors**: Invalid teammate, team, or author format
- **Execution Errors**: Network issues, API failures, or prfeed service problems

## Expected Behavior

When El Presidente runs `/prfeed`:

1. **"Validating prfeed setup and GitHub access..."** - Environment checks
2. **"Auto-detecting current PR for branch: ml/feature-branch"** - PR detection  
3. **"Found PR #123: refactor(auth): add user authentication"** - PR confirmation
4. **"Posting to PRFeed with default team settings..."** - prfeed execution
5. **"✅ PR posted to PRFeed: Review request active for team-platform"** - Success confirmation

**With teammate ping (full name):**
```bash
/prfeed john.doe
```
1. **"Posting PR #123 to PRFeed with ping to john.doe"** - Targeted request
2. **"✅ Review request posted - john.doe has been notified"** - Confirmation

**With teammate ping (short name):**
```bash
/prfeed shikha
```
1. **"Resolving short name 'shikha' to 'shikha.gulati'"** - Name resolution
2. **"Posting PR #123 to PRFeed with ping to shikha.gulati"** - Targeted request  
3. **"✅ Review request posted - shikha.gulati has been notified"** - Confirmation

**With custom options:**
```bash
/prfeed --team-name team-infra --author michael.looney alice.smith
```
1. **"Posting to PRFeed with custom team: team-infra, author: michael.looney"** - Custom config
2. **"Pinging alice.smith for review of current PR"** - Teammate notification
3. **"✅ Review request active on team-infra with alice.smith notified"** - Success

## Advanced Features

**Intelligent PR Context:**
- Auto-detect PR details (title, description, changed files)
- Suggest appropriate teammates based on file changes or code ownership
- Integrate with CODEOWNERS file for automatic reviewer suggestions
- Support multiple PR formats (GitHub, GitLab, etc.)

**Team and Workflow Integration:**
- Cache frequently used team names and teammates for faster completion
- Integrate with Slack for notification confirmations
- Support batch operations for multiple PRs
- Remember user preferences for team and author settings

**Error Recovery and Guidance:**
- Suggest creating PR if none exists for current branch
- Provide setup instructions if prfeed tool is missing
- Guide through GitHub token configuration if authentication fails
- Offer alternative review request methods if PRFeed is unavailable

**Usage Analytics and Feedback:**
- Track successful review request patterns
- Provide insights on review turnaround times
- Suggest optimal teammates based on historical review data
- Integration with development workflow metrics

## Error Handling Examples

**No PR Found:**
```
❌ Error: No PR found for current branch 'ml/feature-branch'
💡 Suggestion: Run '/pr' to create a pull request first, then use '/prfeed'
```

**Missing GitHub Token:**
```
❌ Error: GitHub API authentication failed
💡 Solution: Set GITHUB_TOKEN environment variable with your GitHub personal access token
```

**PRFeed Tool Missing:**
```
❌ Error: prfeed command not found in PATH
💡 Solution: Install prfeed tool or ensure /Users/looneym/dotfiles/bin is in your PATH
```

**Invalid Teammate:**
```
❌ Error: Teammate 'invalid..user' contains invalid characters
💡 Format: Use valid username format like 'john.doe' or 'alice.smith'
```

## Security Considerations

**Token Management:**
- Never log or expose GitHub tokens in command output
- Validate token permissions before API calls
- Use secure token storage mechanisms when possible
- Warn about token expiration and renewal requirements

**Input Validation:**
- Sanitize all user inputs to prevent command injection
- Validate teammate and team name formats against expected patterns
- Escape special characters in PR URLs and descriptions
- Implement reasonable limits on parameter lengths

**Network Security:**
- Use HTTPS for all external API calls
- Validate SSL certificates for GitHub and PRFeed endpoints
- Implement timeout mechanisms for network operations
- Handle network failures gracefully without exposing sensitive information
