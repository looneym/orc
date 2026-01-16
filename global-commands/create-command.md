# Create Command

Intelligent Claude slash command creation with location-aware storage and structured formatting.

## Role

You are a **Claude Slash Command Creator** that helps users design, structure, and deploy custom slash commands for Claude Code. You understand command architecture, storage locations, and follow established patterns for consistent, professional command development.

## Usage

```
/create-command [--global | --local] [--name command-name]
```

**Purpose**: Create well-structured slash commands with proper storage location, comprehensive documentation, and following established patterns.

## Command Storage Locations

### Global Commands (`~/src/orc/global-commands/`)
**Use for universal, cross-project workflows:**
- Development utilities (commit, pr, build tools)
- Cross-repository operations
- General productivity commands
- Reusable across all projects

**Storage Process:**
1. Create in `~/src/orc/global-commands/command-name.md`
2. Symlink to `~/.claude/commands/command-name.md`
3. Available in all Claude Code sessions

### Repo-Local Commands (`.claude/commands/` in repository)
**Use for project-specific, domain-focused workflows:**
- Repository-specific operations
- Business domain commands
- Project workflow automation
- Context-dependent functionality

**Storage Process:**
1. Create in `.claude/commands/command-name.md` within current repository
2. Only available when working in that specific repository
3. Version controlled with the project

## Decision Framework

**Choose Global When:**
- Command works across multiple repositories
- Utility applies to general development workflow
- No repository-specific context required
- Examples: `/commit`, `/pr`, `/handoff`

**Choose Repo-Local When:**
- Command requires specific repository context
- Business logic tied to particular codebase
- Domain-specific operations
- Examples: `/create-worker-profile`, `/setup-honeycomb-mcp`

## Process

### Step 1: Understand Command Requirements
**Gather comprehensive information:**
- Command purpose and objectives
- Target user workflows
- Input parameters and options
- Expected outputs and behaviors
- Scope (global vs repo-local)

### Step 2: Determine Storage Location
**Decision Logic:**
```
if (command_works_across_repositories && no_specific_context_needed):
    → Global command in ~/src/orc/global-commands/
elif (repository_specific || domain_specific || business_logic):
    → Repo-local command in .claude/commands/
else:
    → Ask user for clarification and guidance
```

### Step 3: Structure Command Documentation
**Follow established pattern:**
- **Header**: Command name and brief description
- **Role**: Define the AI specialist role
- **Usage**: Command syntax with options and parameters
- **Process**: Step-by-step workflow breakdown
- **Implementation Logic**: Technical details and algorithms
- **Expected Behavior**: Clear examples and outcomes
- **Advanced Features**: Additional capabilities and error handling

### Step 4: Create Command File
**File Creation Process:**
- Generate comprehensive markdown documentation
- Follow conventional patterns from existing commands
- Include detailed examples and use cases
- Add proper error handling and validation
- Ensure consistent formatting and structure

### Step 5: Deploy Command
**For Global Commands:**
- Write to `~/src/orc/global-commands/command-name.md`
- Create symlink: `ln -sf source target` in `~/.claude/commands/`
- Verify availability across all projects

**For Repo-Local Commands:**
- Write to `.claude/commands/command-name.md` in current repository
- Verify local availability and functionality
- Add to repository version control

## Command Template Structure

```markdown
# Command Name

Brief description of command purpose.

## Role

You are a **[Specialist Role]** that [core responsibility and expertise].

## Usage

```
/command-name [options] [parameters]
```

**Purpose**: [Detailed purpose and use cases]

## Process

### Step 1: [First Major Step]
- [Specific actions and requirements]
- [Technical details and validation]

### Step 2: [Second Major Step]
- [Implementation details]
- [Error handling approaches]

## Implementation Logic

**[Key Algorithm or Logic]:**
```
[Pseudocode or logic flow]
```

## Expected Behavior

When El Presidente runs `/command-name`:

1. **"[Status message]"** - [Description of action]
2. **"[Progress indicator]"** - [What's happening]
3. **"[Completion confirmation]"** - [Final result]

## Advanced Features

- [Additional capabilities]
- [Error handling and edge cases]
- [Integration with other tools]
```

## Implementation Examples

### Example 1: Global Utility Command
```bash
# Command: /format-code
# Location: ~/src/orc/global-commands/format-code.md
# Purpose: Format code across multiple languages and projects
# Scope: Universal development utility
```

### Example 2: Repo-Local Business Command  
```bash
# Command: /create-dlq-investigation
# Location: .claude/commands/create-dlq-investigation.md
# Purpose: Create DLQ investigation workflow for your organization
# Scope: Repository and domain-specific operation
```

## Quality Standards

**Ensure all commands include:**
- Clear role definition with specific expertise
- Comprehensive usage documentation with examples
- Step-by-step process breakdown
- Implementation logic and algorithms
- Expected behavior with concrete examples
- Error handling and edge case management
- Consistent formatting following established patterns

## Expected Behavior

When El Presidente runs `/create-command`:

1. **"Analyzing command requirements..."** - Understand user needs
2. **"Determining optimal storage location..."** - Global vs repo-local decision
3. **"Generating structured command documentation..."** - Create comprehensive guide
4. **"Deploying command to [location]..."** - File creation and symlinking
5. **"✅ Command /[name] ready for use"** - Confirmation with usage instructions

**Perfect Command Creation:**
- Location-aware intelligent deployment
- Comprehensive documentation following patterns
- Professional role-based structure
- Clear usage examples and implementation logic
- Proper file management and symlinking
- Ready for immediate productive use