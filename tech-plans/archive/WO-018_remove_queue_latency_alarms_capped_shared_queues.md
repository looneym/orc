# Work Order #018: Remove Queue Latency Alarms from Capped Shared Queue Workers

**Created**: 2025-08-20  
**Category**: 🔧 Infrastructure  
**Priority**: Medium  
**Effort**: S  
**IMP Assignment**: Unassigned

## Problem Statement

During ops-pagerduty channel investigation, El Presidente discovered that multiple Ruby worker classes are configured with `queue_latency_maximum_queue_latency` worker options despite running on capped shared queues. This creates inappropriate queue latency alarms that don't make architectural sense for workers operating in this constraint model.

**Root Issue**: Queue latency alarms are meaningless for workers running on capped shared queues since the queues are intentionally constrained and latency is expected behavior, not an actionable alert condition.

**Evidence Source**: Honeycomb ASG rag status analysis: `https://ui.honeycomb.io/intercomops/datasets/asg-rag-status/result/ipAV2uDdvuN`

## Acceptance Criteria

### Phase 1: Worker Class Identification
- [ ] **Honeycomb Analysis**: Extract complete list of Ruby worker classes from the provided Honeycomb result
- [ ] **Code Location**: Locate each identified worker class in the intercom codebase
- [ ] **Configuration Audit**: Identify which workers currently have `queue_latency_maximum_queue_latency` options configured
- [ ] **Impact Assessment**: Document current alarm behavior for each affected worker

### Phase 2: Configuration Cleanup
- [ ] **Remove Latency Options**: Delete `queue_latency_maximum_queue_latency` worker options from identified classes
- [ ] **Verify Architecture**: Confirm each worker is indeed running on capped shared queues
- [ ] **Code Review**: Ensure removal doesn't affect other worker functionality
- [ ] **Documentation Update**: Update any relevant comments or documentation

### Phase 3: Validation and Deployment
- [ ] **Testing**: Verify worker classes continue to function correctly without latency alarms
- [ ] **Alarm Verification**: Confirm inappropriate latency alarms are eliminated
- [ ] **Monitoring**: Ensure other monitoring and alerting remains intact
- [ ] **Team Communication**: Document the architectural reasoning for future reference

## Technical Context

**Repository**: intercom (main application repository)

**Target Configuration**: Ruby worker classes with inappropriate queue latency alarm settings

**Configuration Pattern to Remove**:
```ruby
# REMOVE this from worker options in identified classes:
queue_latency_maximum_queue_latency: <value>
```

**Architectural Context**:
- Workers are running on capped shared queues (constrained by design)
- Latency is expected behavior due to queue architecture, not a failure condition
- Current alarms create noise without actionable insights
- Infrastructure is capped at current size due to datastore risk constraints

**Worker Identification Source**: 
- Honeycomb dataset: `asg-rag-status`
- Query result: `ipAV2uDdvuN`
- Focus: Ruby worker classes appearing in this analysis

## Resources & References

- **Investigation Thread**: #ops-pagerduty channel discussion on 2025-08-20
- **Evidence Query**: https://ui.honeycomb.io/intercomops/datasets/asg-rag-status/result/ipAV2uDdvuN
- **Architecture Context**: Capped shared queue model with datastore risk constraints
- **El Presidente's Analysis**: "none of these worker classes should have queue latency alarms if they are effectively running on a capped shared queue"

## Implementation Notes

**Investigation Workflow**:
1. **Extract Worker List**: Parse Honeycomb result for all Ruby worker class names
2. **Code Search**: Use grep/search to locate worker class definitions in codebase
3. **Configuration Review**: Identify `queue_latency_maximum_queue_latency` usage patterns
4. **Surgical Removal**: Remove only the inappropriate latency configuration options

**Search Patterns**:
```bash
# Find worker classes with queue latency configuration
grep -r "queue_latency_maximum_queue_latency" app/workers/
```

**Expected Impact**:
- Reduced alert noise from inappropriate queue latency alarms
- Cleaner monitoring aligned with actual architectural constraints
- Better operational signal-to-noise ratio for meaningful alerts

**Safety Considerations**:
- Only remove latency alarms, preserve other worker options
- Verify each worker is confirmed to be on capped shared queue architecture
- Maintain other monitoring and alerting mechanisms

---

## Work Order Lifecycle

### Status History
- **2025-08-20**: Created → 02-NEXT (ops investigation completed, ready for implementation)

### IMP Notes
**Status**: 📅 **NEXT** - Queue latency alarm cleanup ready for implementation

**El Presidente's Investigation**: Discovered through ops-pagerduty channel analysis that multiple Ruby worker classes have inappropriate `queue_latency_maximum_queue_latency` configurations despite running on capped shared queues.

**Key Insight**: Queue latency alarms are architecturally meaningless for capped shared queue workers since latency is expected constraint behavior, not actionable alert condition.

**Scope**: Specific Ruby worker classes identified in Honeycomb ASG rag status result `ipAV2uDdvuN`

**Expected Outcome**: 
- Eliminated inappropriate queue latency alarms 
- Reduced operational noise
- Better aligned monitoring with actual infrastructure architecture

**Implementation Focus**: Surgical removal of `queue_latency_maximum_queue_latency` options from identified worker classes only

**Next Steps**: 
1. Extract worker class list from Honeycomb result
2. Locate and audit worker configurations in intercom codebase  
3. Remove inappropriate latency alarm configurations
4. Verify workers continue functioning without latency options
5. Validate alarm noise reduction

---
*Work Order #018 - Forest Manufacturing System*