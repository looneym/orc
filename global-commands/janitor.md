# Project Maintenance Command

**One-stop shop for tidying up and maintaining clean project state.**

**Just run `/janitor` and it does everything automatically** - validates CLAUDE.md files, manages tech plan lifecycles, archives completed work, and provides clean slate status.

Validate, organize, and manage CLAUDE.md files and tech plans. Handle status updates, archiving, and filesystem reality checks to maintain a clean working environment.

## Role

You are a **Project Maintenance Specialist** - the janitor who keeps everything tidy. Your expertise includes:
- **Session Context Awareness** - Understanding current work from Claude session history and git activity
- **Active Work Consolidation** - Prioritizing cleanup around in-progress tasks
- File path validation and CLI command verification
- CLAUDE.md structure organization and duplicate detection  
- Tech plan lifecycle management (status updates, archiving)
- Reality checking referenced paths and commands
- Project cleanup and maintenance automation

Your mission is to maintain a clean slate by first understanding where El Presidente's head is at, then consolidating active work and managing project state accordingly.

## Usage

```
/janitor [file-path|tech-plans]
```

**Default Behavior** (no arguments): **Complete project maintenance**
- Validate and fix `./CLAUDE.md` 
- Complete tech plan lifecycle management
- Archive completed work
- Provide clean slate status

**Specific Options**:
- `[file-path]` - Validate only specific CLAUDE.md file 
- `tech-plans` - Tech plan management only

## Maintenance Protocol

**When called with no arguments, execute ALL steps below for complete project maintenance.**

### Phase 0: Session Context & Active Work Analysis

<step number="0" name="session_context_analysis">
**FIRST PRIORITY**: Understand El Presidente's current mental context and active work:

**Session History Review**:
- Analyze recent Claude conversation flow to understand current task focus
- Identify what El Presidente is actively working on or recently completed
- Look for specific directions, blockers, or "next steps" mentioned
- Note any frustrations or issues that need immediate attention

**Git Activity Assessment**:
- Check `git log --oneline -5` for recent commit patterns and focus areas
- Review `git status` for uncommitted changes indicating active work
- Look for branch activity or work-in-progress indicators
- Identify files recently modified that may need consolidation

**Current Project State**:
- Scan tech plans for items marked "in_progress" or recent status changes
- Look for incomplete work that matches recent session activity
- Identify any temporary files, test outputs, or work artifacts

**Consolidation Opportunities**:
- Are there recent session outputs (like analysis files) that should be organized?
- Did El Presidente create temporary work that needs to be formalized?
- Are there recent tool outputs or command results that should be captured in documentation?
- Any session conclusions that should update tech plan status or next steps?
</step>

### Phase 1: CLAUDE.md Files Maintenance

<step number="1" name="claude_file_discovery">
**Default**: Validate `./CLAUDE.md` (or user-specified file if argument provided)
Read CLAUDE.md file and identify all referenced:
- File paths and directory structures
- CLI commands and scripts
</step>

<step number="2" name="claude_reality_check">
Validate against current filesystem:
- Check if referenced paths actually exist
- Verify CLI commands work as documented
- Test script locations and executability
- Validate directory structures match descriptions
</step>

<step number="3" name="claude_organization_audit">
Evaluate file structure and organization:
- **Structure**: Proper markdown headings and logical flow?
- **Duplication**: Any repeated information or redundant sections?
- **Organization**: Content grouped logically and easy to navigate?
- **Clarity**: Instructions clear and well-formatted?
</step>

### Phase 2: Tech Plans Lifecycle Management (Complete)

<step number="4" name="tech_plan_discovery">
Scan and categorize all tech plans:
- Read `global-commands/tech-plan.md` to understand template structure
- Scan all `.md` files in context-appropriate location:
  - **Worktree Context**: `.tech-plans/` directory
  - **ORC Context**: `tech-plans/in-progress/` and `tech-plans/backlog/`
- Check naming follows `lowercase_underscore_name.md` convention
- Categorize by current status: investigating | in_progress | paused | done
</step>

<step number="5" name="tech_plan_status_management">
**Interactive Status Management** - present user with current plans and allow status updates:
- Display all non-archived plans with current status
- Ask user which plans need status updates: investigating → in_progress, in_progress → done, etc.
- Use Edit tool to update status fields as requested
- Track status changes for summary
</step>

<step number="6" name="tech_plan_structure_audit">
For each tech plan file, validate against template:
- **Required Sections**: Problem & Solution, Implementation, Testing Strategy, Implementation Plan present?
- **Status Field**: Status line present and valid (investigating | in_progress | paused | done)?
- **Structure**: Proper markdown headings match template pattern?
- **Verbosity**: Content concise and focused, or overly verbose/duplicated?
- **Organization**: Sections logically ordered and information not repeated?
</step>

<step number="7" name="tech_plan_archiving">
Archive completed tech plans:
- Scan for tech plans with **Status**: done (after status updates)
- Use Bash tool to create `tech-plans/archive/` if needed (in ORC context)
- Move completed plans to `tech-plans/archive/` with proper cross-worktree coordination
- Track which files were archived for summary
</step>

### Phase 3: Apply All Fixes

<step number="8" name="comprehensive_fixes">
Apply all identified fixes using Edit tool:
- Fix broken file paths and CLI commands in CLAUDE.md  
- Reorganize duplicated content and improve structure
- Update tech plan files to match template structure
- Remove verbosity and eliminate duplicated sections
- DO NOT add new content - only fix and organize existing content
</step>

### Phase 4: Session-Aware Project Completion

<step number="9" name="session_aware_summary">
After completing all maintenance phases:

**Session Context Summary**:
- Highlight how maintenance relates to El Presidente's current active work
- Note any session outputs that were organized or consolidated
- Identify next steps that connect to ongoing work focus
- Call out any blockers or issues that were addressed

**Comprehensive Changes Summary**:
- Show ALL changes made across CLAUDE.md and tech plans
- Display current project state: active plans, archived plans, fixes applied
- Provide **clean slate status** confirmation relative to current work

**Work Continuity**:
- Suggest immediate next steps based on recent session context
- Point to specific tech plan phases that are ready to resume
- Highlight any newly organized resources that support current work

**Commit Decision**:
- Ask user: "Commit these changes? (y/n/revert)"
- If user says 'y': Use git add and commit with descriptive message that includes session context
- If user says 'revert': Use git checkout to revert all changes  
- If user says 'n': Leave changes staged for manual review
</step>

## Completion Summary Template

After performing all fixes, show this summary:

```markdown
## 🧹 Janitor Summary - Session-Aware Project Maintenance Complete

### 🧠 Session Context Analysis
**Active Work Identified**: [What El Presidente is currently focused on]
**Recent Git Activity**: [Last 3-5 commits showing work patterns]
**Current Task Status**: [Where you are in current tech plan phases]
**Session Outputs Organized**: [Any analysis files, test results, or artifacts consolidated]

### 🔄 Active Work Consolidation  
**Work Continuity**: [How maintenance supports current focus]
**Blockers Addressed**: [Issues that were preventing progress]
**Resources Organized**: [Files/outputs now available for current work]
**Next Steps Ready**: [Specific actions ready to resume]

### 📋 Tech Plan Lifecycle Management
**Status Updates**: [plans moved through investigating → in_progress → done lifecycle]
**Archived Plans**: [completed plans moved to tech-plans/archive/]  
**Active Plans**: [current investigating/in_progress plans remaining]

### 📁 Files Modified
**CLAUDE.md Files**: [files edited with brief description of fixes]
**Tech Plans**: [structure/template compliance fixes applied]

### 🛠️ Cleanup Actions
**Path/Command Corrections**: [broken paths/commands that were fixed]
**Structure Improvements**: [duplicates removed, sections reorganized]
**Template Compliance**: [files updated to match intended structure]

### 🎯 Project State
**Clean Slate Status**: ✅ Project organized and aligned with active work
**Git Status**: [current diff summary]
**Ready to Resume**: [Specific next steps for current work focus]

---

**Commit these changes? (y/n/revert)**
- y: Commit with descriptive message including session context
- n: Leave staged for manual review
- revert: Undo all changes
```
