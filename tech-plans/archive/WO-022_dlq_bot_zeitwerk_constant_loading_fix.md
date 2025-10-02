# Work Order #022: DLQ Bot Zeitwerk Constant Loading Fix

**Created**: 2025-08-26  
**Category**: 🐛 Bug Fix  
**Priority**: High  
**Effort**: S  
**IMP Assignment**: Unassigned

## Problem Statement

The DLQ bot rollout to EMS (Event Management System) failed in production due to a Zeitwerk constant loading error. The deployment broke the PST environment and had to be reverted over the weekend.

**Production Error**: 
```
Could not spawn process for application /apps/event-management-system/current: 
The application encountered the following error: expected file 
/apps/event-management-system/releases/.../app/lib/dlqbot/create_issue.rb 
to define constant Dlqbot::CreateIssue, but didn't (Zeitwerk::NameError)
```

**Root Cause**: Classic Rails/Zeitwerk autoloading issue where the file path structure doesn't match Rails constant naming conventions, causing the application to fail to start.

**Context**: This issue stems from **WO-019: CloudBot DLQ Issue Creation Automation** which was completed and deployed on Friday. The DLQ bot GitHub issue creation functionality was working but broke during EMS deployment due to incorrect file/constant structure.

## Acceptance Criteria

### Phase 1: Root Cause Analysis
- [ ] **File Structure Investigation**: Analyze current `lib/dlqbot/create_issue.rb` file structure and constant definitions
- [ ] **Zeitwerk Compliance**: Identify specific mismatch between file path and expected constant name
- [ ] **Rails Conventions**: Document correct file naming and constant definition patterns
- [ ] **Deployment Context**: Understand why this worked in development but failed in production

### Phase 2: Fix Implementation
- [ ] **Correct File Structure**: Ensure file path matches Rails/Zeitwerk constant expectations
- [ ] **Proper Constant Definition**: Define `Dlqbot::CreateIssue` constant correctly in the file
- [ ] **Namespace Consistency**: Verify all related DLQ bot files follow consistent naming patterns
- [ ] **Autoload Verification**: Test that Zeitwerk can properly load the constant

### Phase 3: Quality Assurance & Deployment
- [ ] **Spec Coverage**: Add missing test coverage (as noted by Miles McGuire's feedback)
- [ ] **Local Testing**: Verify fix works in development environment with proper autoloading
- [ ] **Staging Deployment**: Test deployment in staging environment before production
- [ ] **Production Rollout**: Careful production deployment with monitoring

## Technical Context

**Repository**: event-management-system (EMS)

**Failed Component**: DLQ bot GitHub issue creation functionality from WO-019

**Zeitwerk Requirements**:
- File path: `app/lib/dlqbot/create_issue.rb`
- Expected constant: `Dlqbot::CreateIssue` 
- Must follow Rails autoloading conventions

**Error Pattern**: 
```ruby
# Current (broken) - file doesn't define expected constant
# app/lib/dlqbot/create_issue.rb
# Missing or incorrectly named Dlqbot::CreateIssue

# Expected (correct) structure needed
module Dlqbot
  class CreateIssue
    # implementation
  end
end
```

**Environment Impact**: 
- ✅ Development: Working (autoloading more permissive)
- ❌ Production: Failed (strict Zeitwerk autoloading)
- ❌ PST: Broken, had to be reverted

## Resources & References

- **Source Work Order**: WO-019 - CloudBot DLQ Issue Creation Automation (completed, caused this issue)
- **Production Error**: Slack message from Dec McMullen with full error trace
- **Team Feedback**: Miles McGuire noted specs would have caught this issue
- **Related WO-012**: DLQ Bot Foundations (foundational work that this builds on)

## Implementation Notes

**Investigation Areas**:
1. **Current File Structure**: Check exact content of `lib/dlqbot/create_issue.rb`
2. **Constant Definition**: Verify how `Dlqbot::CreateIssue` is currently defined (or missing)
3. **Rails Conventions**: Ensure file path → constant mapping follows Zeitwerk rules
4. **Related Files**: Check other DLQ bot files for consistent naming patterns

**Zeitwerk Rules**:
- `app/lib/dlqbot/create_issue.rb` → `Dlqbot::CreateIssue` constant required
- File must define the constant that Zeitwerk expects from the path
- Module nesting must match directory structure

**Fix Approach**:
```ruby
# Correct structure for app/lib/dlqbot/create_issue.rb
module Dlqbot
  class CreateIssue
    # GitHub issue creation logic from WO-019
    def call(alarm_data)
      # implementation
    end
  end
end
```

**Testing Requirements**:
- Unit tests for `Dlqbot::CreateIssue` class
- Integration tests for constant loading
- Deployment verification tests
- Autoloading behavior validation

**Success Metrics**:
- Application starts successfully in all environments
- Zeitwerk can load `Dlqbot::CreateIssue` constant without errors
- DLQ bot GitHub issue creation functionality restored
- Test coverage added to prevent regression

---

## Work Order Lifecycle

### Status History
- **2025-08-26**: Created → 02-NEXT (urgent production fix needed)

### IMP Notes
**Status**: 📅 **NEXT** - High priority production fix for DLQ bot deployment failure

**Urgency**: High - Production functionality broken and reverted over weekend

**Root Issue**: WO-019 implementation had Zeitwerk constant loading bug that wasn't caught in development due to more permissive autoloading behavior.

**Team Feedback Integration**: 
- Miles McGuire noted specs would have caught this - add comprehensive test coverage
- Dec McMullen reverted over weekend - need careful re-deployment approach

**Fix Scope**: 
1. Correct file structure and constant definition for Zeitwerk compliance
2. Add missing test coverage to prevent future regressions  
3. Verify deployment process works across all environments
4. Restore DLQ bot GitHub issue creation functionality

**Expected Outcome**: 
- DLQ bot functionality restored and working in production
- Robust test coverage prevents similar Zeitwerk issues
- Clear deployment verification process for Rails constant loading

**Repository**: event-management-system (EMS) - same repo where deployment failed

**Next Steps**: 
1. Investigate current file structure and constant definition issues
2. Fix Zeitwerk compliance in `lib/dlqbot/create_issue.rb`
3. Add comprehensive test coverage for DLQ bot functionality
4. Test deployment process thoroughly before production rollout

---
*Work Order #022 - Forest Manufacturing System*