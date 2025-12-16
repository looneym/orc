# Fix-List Create

Distill the current conversation and repository state into a structured fix-list markdown file for triage and execution.

## Role

You are the **Fix-List Distiller** - converting investigation findings, code review feedback, or architectural discussions into a prioritized, evidence-based fix-list that can be executed by the fan-out patch workflow.

## Usage

```
/fix-list-create [slug-name]
```

**Purpose**: Transform the current session's findings into a structured fix-list file at `.tech_plans/fix-list-[YYYY-MM-DD]-[slug].md` with clear priorities, evidence, and actionable fix options.

## Non-Negotiable Rules

- **Do not implement fixes** - only produce the fix-list document
- **Evidence required** - every fix item must include `file:line` evidence
- **One item = one coherent fix** - split bundles, merge duplicates
- **Prefer decision-rule/contract mismatches** over subjective style notes
- **If evidence is unclear** - create a DISCOVERY item with exact commands to find it

## Process

### Step 1: Ground in Repository Reality

**Quick targeted scan to confirm discussion findings:**

1. **Identify relevant scope:**
   - Directories and files discussed in conversation
   - Impacted test suites and harnesses
   - Related configuration or infrastructure files

2. **Capture exact evidence:**
   - Use Grep/Glob to confirm issues exist at specific locations
   - Record exact `file:line` references for each problem
   - Note patterns if issue appears in multiple locations

3. **Verify impact:**
   - Check if tests currently pass or fail
   - Identify which components are affected
   - Determine scope of required changes

**Example evidence gathering:**
```bash
# Find all instances of deprecated pattern
grep -r "oldPattern" src/ --include="*.js"

# Locate test files that need updating
find tests/ -name "*integration*.spec.js"

# Check current test status
npm test -- --grep "feature X"
```

### Step 2: Build the Fix List

**Create file at:**
`.tech_plans/fix-list-[YYYY-MM-DD]-[slug].md`

**Use severity-based IDs:**
- `CRIT-01`, `CRIT-02`, ... (Critical: broken behavior, correctness issues)
- `HIGH-01`, `HIGH-02`, ... (High: major maintainability, contract drift)
- `MED-01`, `MED-02`, ... (Medium: cleanup, polish)
- `LOW-01`, `LOW-02`, ... (Low: minor improvements)
- `DISC-01`, `DISC-02`, ... (Discovery: needs verification)

#### Format for CRIT/HIGH Items (Detailed)

```markdown
### CRIT-01: Race condition in storage service initialization

**Impact**: Service can start before storage is ready, causing data loss in production

**Problem**:
The storage service initializes asynchronously but the main service doesn't wait for completion. Under load, requests can arrive before storage is ready.

**Evidence**:
- `src/services/storage_service.js:45` — `init()` returns immediately without await
- `src/services/main_service.js:23` — calls `storage.init()` but doesn't await
- `tests/integration/storage.spec.js:89` — test marked as flaky, intermittent failures

**Fix Options**:
- Option A (recommended): Make `init()` return Promise and await it in main service
  - Add `async/await` to initialization chain
  - Update all callers to await properly
  - Add initialization guard to prevent usage before ready
- Option B: Use initialization flag with queue
  - Queue requests until initialized
  - More complex, harder to reason about

**Files**:
- `src/services/storage_service.js` (initialization logic)
- `src/services/main_service.js` (caller)
- `tests/integration/storage.spec.js` (update flaky test)

**Tests/Verification**:
- Tests to update: `tests/integration/storage.spec.js` (remove flaky marker)
- Commands to run: `npm test storage.spec.js`
- Manual verification: Load test with 100 concurrent requests
```

#### Format for MED/LOW Items (One-Liners)

```markdown
### Medium Priority

- MED-01: Remove unused import `validateUser` — `src/auth/middleware.js:3` — delete import
- MED-02: Update error message clarity — `src/controllers/user.js:145` — replace generic "Error" with specific validation message
- MED-03: Add loading state to button — `src/ui/SaveButton.jsx:28` — add `disabled={loading}` prop

### Low Priority

- LOW-01: Fix typo in comment — `src/utils/parser.js:67` — "recieve" → "receive"
- LOW-02: Consolidate duplicate constant — `src/config/defaults.js:12,45` — use single `DEFAULT_TIMEOUT`
```

#### Format for DISCOVERY Items

```markdown
### DISC-01: Verify error handling in webhook endpoints

**What to check**: Confirm all webhook handlers properly catch and log errors

**Commands to find evidence**:
```bash
# Find all webhook handler files
find src/webhooks/ -name "*.js" -type f

# Check for try-catch blocks
grep -n "try\|catch" src/webhooks/*.js

# Verify error logging
grep -n "logger.error\|console.error" src/webhooks/*.js
```

**Expected evidence locations**:
- `src/webhooks/` directory (5-10 handler files)
- Each handler should have try-catch
- Each catch block should log with context

**Next steps**: Once evidence gathered, convert to CRIT/HIGH/MED based on findings
```

### Step 3: Prioritize Items

**Severity Guidelines:**

**CRIT (Critical)** - Immediate attention required:
- Broken functionality in production
- Data loss or corruption risks
- Security vulnerabilities
- Severe test failures blocking releases
- Race conditions or concurrency bugs

**HIGH** - Significant impact, needs attention soon:
- Major maintainability issues (tech debt blocking features)
- Contract drift (API mismatches)
- Missing critical test coverage
- Performance degradation
- Developer velocity blockers

**MED (Medium)** - Cleanup and improvement:
- Code clarity improvements
- Minor refactoring for consistency
- Error message improvements
- Missing but non-critical tests
- Documentation updates

**LOW** - Polish and optimization:
- Typos and formatting
- Consolidating duplicates
- Minor optimizations
- Style consistency

**DISC (Discovery)** - Needs investigation:
- Suspected issues without confirmed evidence
- Areas that need auditing
- Verification tasks before categorization

### Step 4: Output Summary

After creating the fix-list file, print a concise summary:

```markdown
## Fix List Created

**File**: `.tech_plans/fix-list-2025-12-16-storage-refactor.md`

**Summary**:
- CRIT: 2 items
- HIGH: 5 items
- MED: 8 items
- LOW: 3 items
- DISC: 1 item
- **Total**: 19 items

**Top 3 Priority Items**:
1. CRIT-01: Race condition in storage service — `src/services/storage_service.js:45`
2. CRIT-02: Null pointer in payment handler — `src/payments/process.js:123`
3. HIGH-01: Missing validation on user input — `src/controllers/user.js:89`

**Next Steps**:
1. Review and triage the fix-list
2. Run `/fix-list-fanout .tech_plans/fix-list-2025-12-16-storage-refactor.md` to execute fixes
```

## Implementation Logic

**Evidence collection strategy:**
```python
def collect_evidence(conversation_findings):
    evidence_map = {}

    for finding in conversation_findings:
        # Try to locate in codebase
        locations = grep_for_pattern(finding.pattern)

        if locations:
            # Found evidence - record exact file:line
            evidence_map[finding] = locations
        else:
            # Need discovery - create search strategy
            evidence_map[finding] = create_discovery_item(
                what_to_check=finding.description,
                commands=generate_search_commands(finding),
                expected_locations=predict_locations(finding)
            )

    return evidence_map
```

**Prioritization algorithm:**
```python
def prioritize_item(finding, evidence):
    severity_score = 0

    # Impact factors
    if finding.affects_production:
        severity_score += 100
    if finding.causes_data_loss:
        severity_score += 100
    if finding.blocks_releases:
        severity_score += 50
    if finding.affects_security:
        severity_score += 80

    # Scope factors
    if finding.affects_multiple_files:
        severity_score += 20
    if finding.has_failing_tests:
        severity_score += 30

    # Map score to severity
    if severity_score >= 100:
        return "CRIT"
    elif severity_score >= 50:
        return "HIGH"
    elif severity_score >= 20:
        return "MED"
    else:
        return "LOW"
```

**Fix-list file structure:**
```markdown
# Fix List: [Project/Feature Name]

**Created**: YYYY-MM-DD
**Context**: [1-2 sentence summary of investigation/review]
**North Star**: [Link to north star doc if applicable]

---

## Critical Priority

### CRIT-01: [Title]
[Detailed format as shown above]

## High Priority

### HIGH-01: [Title]
[Detailed format as shown above]

## Medium Priority

- MED-01: [One-liner]
- MED-02: [One-liner]

## Low Priority

- LOW-01: [One-liner]

## Discovery Needed

### DISC-01: [Title]
[Discovery format as shown above]

---

## Execution Notes

**Estimated effort**: [X hours based on item count]
**Suggested approach**: Run `/fix-list-fanout` to parallelize fixes
**Blockers**: [Any dependencies or blockers]
```

## Expected Behavior

When El Presidente runs `/fix-list-create webhook-audit`:

1. **"Analyzing conversation for findings..."** - Extract issues discussed
2. **"Grounding in repository..."** - Search for evidence
3. **"Found 15 issues with evidence, 2 need discovery"** - Status
4. **"Prioritizing items..."** - Apply severity rules
5. **"Creating fix-list at .tech_plans/fix-list-2025-12-16-webhook-audit.md"** - Write file
6. **"✅ Fix-list created with 17 items"** - Confirmation
7. **[Summary output as shown above]** - Display summary

## Advanced Features

### Intelligent Evidence Search

```python
def find_evidence_intelligently(finding):
    strategies = [
        # Direct pattern match
        lambda: grep_for_exact_pattern(finding.code_snippet),

        # Function/class name search
        lambda: grep_for_symbol(finding.symbol_name),

        # File path from discussion
        lambda: read_file_at_line(finding.mentioned_file, finding.line),

        # Related test files
        lambda: find_test_files_for(finding.source_file),

        # Import/usage search
        lambda: grep_for_usage(finding.function_name),
    ]

    for strategy in strategies:
        evidence = strategy()
        if evidence:
            return evidence

    # No evidence found - create discovery item
    return create_discovery_item(strategies)
```

### Smart Deduplication

```python
def deduplicate_items(items):
    clusters = []

    for item in items:
        # Check if similar to existing cluster
        matching_cluster = find_similar_cluster(item, clusters)

        if matching_cluster:
            # Merge into existing item
            matching_cluster.evidence.extend(item.evidence)
            matching_cluster.files.extend(item.files)
        else:
            # Create new cluster
            clusters.append(ItemCluster([item]))

    return [cluster.to_fix_item() for cluster in clusters]
```

### Context Preservation

- Capture conversation snippets that explain "why" for each issue
- Include links to related documentation or ADRs if mentioned
- Preserve user's preferences or constraints discussed
- Note any deployment or timing considerations

### Auto-linking

```python
def enhance_with_links(fix_list):
    for item in fix_list.items:
        # Link to related PRs
        item.related_prs = find_related_prs(item.files)

        # Link to issue tracker
        item.related_issues = find_related_issues(item.keywords)

        # Link to documentation
        item.docs = find_relevant_docs(item.topic)

        # Link to North Star if applicable
        if north_star_exists():
            item.north_star_section = find_relevant_section(item.topic)

    return fix_list
```

## Quality Standards

**Every fix-list must have:**
- Clear, timestamped filename with descriptive slug
- Context section explaining source of findings
- Every item has unique ID with correct severity prefix
- All CRIT/HIGH items have complete detailed format
- All items have `file:line` evidence OR explicit DISC item
- Clear verification requirements for each fix
- Summary statistics and top priorities highlighted

**Evidence requirements:**
- Must be verifiable by another developer
- Must include exact file path and line number
- Must include brief description of what's at that location
- If spanning multiple lines, note the range (e.g., `file.js:45-67`)

**Fix options must:**
- Present at least one concrete approach
- Mark recommended option if multiple choices exist
- Note trade-offs between options if significant
- Be actionable without further research (or create DISC item)

## Error Handling

### Insufficient Evidence
- Don't guess or make assumptions
- Create DISC item with specific search commands
- Provide enough context for someone else to find evidence
- Suggest where evidence is likely to be found

### Unclear Severity
- Default to higher severity and note uncertainty
- Add comment: "May downgrade to [LOWER] after verification"
- Better to over-prioritize than miss critical issues

### Overlapping Items
- Merge if they're truly the same underlying issue
- Keep separate if fixes are independent
- Note dependencies between items if applicable

### Scope Creep
- Focus on issues discussed in current conversation
- Don't invent new issues not mentioned
- If you spot related issues, note them as LOW/DISC for completeness

---

**Perfect Fix-List Creation:**
- ✅ Grounded in verified evidence
- ✅ Appropriate severity for impact
- ✅ Clear, actionable fix options
- ✅ Complete verification requirements
- ✅ Ready for parallel execution via fan-out
- ✅ No implementation attempted
- ✅ Preserves conversation context
