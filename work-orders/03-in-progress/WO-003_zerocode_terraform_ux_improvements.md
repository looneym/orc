# Work Order #003: ZeroCode Terraform UX Improvements

**Created**: 2025-08-15  
**Category**: 🔧 Enhancement  
**Priority**: High  
**Effort**: XL  
**IMP Assignment**: IMP-ZEROCODE (active - backported from existing worktree)

## Problem Statement

ZeroCode Terraform tool has significant UX pain points based on user feedback from Miles McGuire. First-time users experience confusion about the workflow timing and manual steps required. The tool needs UX improvements to make it feel more like the intuitive "old muster workflow" while maintaining Terraform benefits.

## Acceptance Criteria

- [ ] **Auto-merge implementation**: Enable automatic PR merging to eliminate manual step
- [ ] **Smarter change batching**: Batch multiple UI changes into single commits to reduce noise
- [ ] **Reduced flush interval**: Decrease waiting time from 10 minutes to ~2 minutes
- [ ] **Better visual feedback**: Improve user awareness with enhanced modal dialog and progress indicators
- [ ] **Enhanced notification system**: Make notifications more prominent than current UI banner
- [ ] **User workflow streamlining**: Transform experience to feel like old muster workflow

## Technical Context

**Dependencies**: 
- Understanding of ZeroCode Terraform architecture across multiple repos
- Miles McGuire user feedback and conversation transcript
- Current flush/batching logic in event-management-system
- Muster UI components and notification systems

**Repositories**: 
- **event-management-system**: Backend flush/batching logic
- **muster**: Frontend UI components and user experience
- **infrastructure**: Any configuration changes needed

**Complexity Notes**: 
- Multi-repository feature requiring coordinated changes
- User experience redesign with backend workflow modifications
- Integration with existing GitHub PR creation and Slack notification systems
- Need to maintain reliability while improving user experience

## Resources & References

- **User Feedback Transcript**: Miles McGuire conversation about ZeroCode UX issues
- **Current Workflow**: 10-minute flush → #zero-code-terraform Slack → manual PR merge
- **Target Experience**: Old muster workflow feel with Terraform infrastructure benefits
- **Technical Architecture**: Event-management-system processes changes, Muster provides UI

## Implementation Notes

**Key Improvements Needed:**
1. **2-minute batching system** instead of 10-minute intervals
2. **Auto-merge capability** for ZeroCode PRs to eliminate manual step  
3. **Rich modal dialog** with progress tracking and clear next-steps guidance
4. **Enhanced notification visibility** beyond current UI banner approach

**User Journey Target:**
- Make UI change → See immediate feedback → Automatic processing → Seamless completion
- Eliminate confusion about waiting times and manual PR steps

---

## Work Order Lifecycle

### Status History
- **2025-08-15**: Backported from existing worktree → 02-IN-PROGRESS
- **Previous**: Investigation and implementation work completed by IMP-ZEROCODE

### IMP Notes
**Status**: 🔄 **IN PROGRESS** - Phase 1 delivered, investigating Phase 2

## Phase 1: Quick Wins (COMPLETED)
**Key Actions Completed**:
- **2025-08-15**: Analyzed existing ZeroCode architecture and Miles' feedback
- **2025-08-15**: Implemented 2-minute batching system in event-management-system
- **2025-08-15**: Added auto-merge capability for ZeroCode PRs
- **2025-08-15**: Created rich modal dialog with progress tracking in Muster
- **2025-08-15**: Enhanced notification system with better visibility

**Active PRs**: 
- EMS Backend: https://github.com/intercom/event-management-system/pull/2927
- Muster Frontend: https://github.com/intercom/muster/pull/6057

**Phase 1 Results**: Successfully transformed ZeroCode Terraform UX from confusing 10-minute manual workflow to streamlined 2-minute automated experience.

## Phase 2: Eliminate Batching (INVESTIGATED)
**Objective**: Remove need for batching by combining multiple changes into single message/commit/PR

**Investigation Results**:
- ✅ **Current Architecture Mapped**: Each UI change creates separate SQS message with single attribute
- ✅ **TeraBridge Constraint Identified**: Go binary only supports single attribute per invocation
- ✅ **Message Structure Analyzed**: Simple key-value pairs sent individually to EMS worker
- ✅ **Processing Flow Understood**: Each message = separate commit = multiple commits per PR

**Key Findings**:
- **Current Flow**: Form changes → Individual SQS messages → Separate TeraBridge calls → Multiple commits
- **TeraBridge Limitation**: Binary signature `terrabridge <asg.tf> <cluster> <attribute> <value>` (single attribute only)
- **Opportunity**: Can batch at UI level and process sequentially in single commit transaction

**Recommended Solution: Hybrid Approach**
1. **Collect Changes**: Muster collects all form changes into single comprehensive message
2. **Single Message**: New payload format with multiple attributes array  
3. **Sequential Processing**: EMS worker calls TeraBridge multiple times within single commit
4. **Result**: One message → One commit → One PR with all changes

**New Message Format Design**:
```ruby
{
  asg_file: @asg_file,
  cluster_name: cluster_name,
  attributes: [                          # NEW: Array of changes
    { name: "instance_size", value: "2xlarge" },
    { name: "cpu_high_threshold", value: "80" },
    { name: "busyness_high_threshold", value: "50" }
  ],
  app_stage: app_stage,
  user_name: user_name,
}
```

**Implementation Plan**:
- [ ] **Muster Controller**: Modify to collect all changes into single message
- [ ] **New EMS Worker**: Create `BulkUpdateActionWorker` for attribute arrays
- [ ] **Sequential TeraBridge**: Process multiple attributes within single Git transaction
- [ ] **Testing**: Verify single commit/PR creation with multiple changes
- [ ] **Compatibility**: Maintain existing single-attribute message support during transition

**Benefits**:
- ✅ **Eliminates batching delays** (immediate processing)
- ✅ **Single commit per user action** (cleaner history)
- ✅ **Single PR per change session** (easier review)
- ✅ **Maintains TeraBridge reliability** (no binary changes needed)
- ✅ **Preserves transaction safety** (all-or-nothing Git commits)

**Decision Point**: Ready for implementation or defer to separate work order.

---
*Work Order #003 - Forest Manufacturing System*