# Work Order #006: ZeroCode Batching Elimination

**Created**: 2025-08-15  
**Category**: 🔧 Enhancement  
**Priority**: Medium  
**Effort**: L  
**IMP Assignment**: Unassigned

## Problem Statement

ZeroCode Terraform currently uses 2-minute batching to combine multiple UI changes into single PRs, but this still creates user-perceived delays and multiple commits per change session. Users must wait up to 2 minutes even for single changes, and multiple form changes result in multiple commits within the same PR.

**Architecture Investigation Completed**: IMP-ZEROCODE has fully mapped the current system and designed a solution to eliminate batching delays entirely by processing multiple changes as single message/commit/PR combinations.

## Acceptance Criteria

- [ ] **Immediate Processing**: Eliminate 2-minute batching delay for single and multiple changes
- [ ] **Single Commit per Session**: Multiple form changes result in one commit instead of multiple
- [ ] **Single PR per Session**: All changes from one user session create one clean PR
- [ ] **Bulk Message Format**: New message structure supporting multiple attribute arrays
- [ ] **Sequential TeraBridge Processing**: Process multiple attributes within single Git transaction
- [ ] **Backward Compatibility**: Maintain existing single-attribute message support during transition
- [ ] **Transaction Safety**: Ensure all-or-nothing processing for attribute groups

## Technical Context

**Current Architecture (Post-WO-003)**:
- Form changes → Individual SQS messages → Separate TeraBridge calls → Multiple commits per PR
- 2-minute batching window combines messages into single PR but still creates multiple commits
- TeraBridge binary signature: `terrabridge <asg.tf> <cluster> <attribute> <value>` (single attribute only)

**Proposed Architecture**:
- Form changes → Single comprehensive SQS message → Sequential TeraBridge calls → Single commit per PR
- New `BulkUpdateActionWorker` processes attribute arrays within single Git transaction
- Modified Muster controller collects all form changes before sending

**Dependencies**: 
- WO-003 Phase 1 implementation (2-minute batching system)
- Understanding of current EMS worker architecture and TeraBridge integration
- Muster form handling and SQS message creation logic
- **WO-008 Auto-Merge Integration**: Include fixing broken auto-merge functionality as part of comprehensive workflow improvements

**Repositories**: 
- **event-management-system**: New `BulkUpdateActionWorker` for attribute array processing
- **muster**: Modified controller to collect changes into single message
- **infrastructure**: No changes needed (TeraBridge binary unchanged)

## Implementation Strategy

**New Message Format**:
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
1. **Muster Controller Enhancement**: 
   - Collect all form changes before submission
   - Create single message with attributes array
   - Maintain UI responsiveness during collection

2. **New EMS Worker Creation**:
   - `BulkUpdateActionWorker` for processing attribute arrays
   - Sequential TeraBridge calls within single Git transaction
   - Error handling for partial failures

3. **Sequential Processing Logic**:
   - Process multiple attributes within single commit
   - Maintain transaction safety (all-or-nothing)
   - Preserve existing error handling and rollback capabilities

4. **Compatibility Layer**:
   - Support both old single-attribute and new bulk messages during transition
   - Gradual migration strategy
   - Monitoring and observability for both message types

## Resources & References

- **Architecture Investigation**: Complete technical analysis from WO-003 Phase 2
- **Current Implementation**: WO-003 2-minute batching system foundation
- **TeraBridge Documentation**: Binary interface and Git transaction handling
- **User Experience Goal**: Immediate processing with clean Git history

## Benefits Analysis

**User Experience**:
- ✅ **Zero wait time** for immediate processing
- ✅ **Clean Git history** with single commit per change session
- ✅ **Simplified PR review** with consolidated changes

**Technical Benefits**:
- ✅ **Maintains TeraBridge reliability** (no binary changes required)
- ✅ **Preserves transaction safety** with all-or-nothing Git operations
- ✅ **Backward compatibility** during migration period
- ✅ **Cleaner architecture** with purpose-built bulk processing

**Operational Benefits**:
- ✅ **Reduced infrastructure complexity** (eliminate batching timers)
- ✅ **Better error handling** with atomic transactions
- ✅ **Improved monitoring** with clearer message/commit relationships

## Risk Assessment

**Low Risk Implementation**:
- Uses existing TeraBridge binary without modifications
- Builds on proven 2-minute batching foundation from WO-003
- Maintains existing error handling patterns
- Supports gradual rollout with compatibility layer

**Mitigation Strategies**:
- Comprehensive testing with both single and multiple attribute scenarios
- Gradual migration from batching to immediate processing
- Rollback capability to 2-minute batching system if needed

---

## Work Order Lifecycle

### Status History
- **2025-08-15**: Created from WO-003 Phase 2 investigation → 01-BACKLOG

### IMP Notes
**Status**: 📋 **BACKLOG** - Awaiting assignment, architecture investigation complete

**Technical Foundation**: Complete architecture design available from WO-003 Phase 2 investigation

**Next Steps**: 
1. Assign IMP and create work tree environment (or continue with existing ZeroCode IMP)
2. Implement new message format and Muster controller changes
3. Create BulkUpdateActionWorker in EMS
4. Test single commit/PR generation with multiple changes
5. Deploy with compatibility layer for gradual migration

---
*Work Order #006 - Forest Manufacturing System*