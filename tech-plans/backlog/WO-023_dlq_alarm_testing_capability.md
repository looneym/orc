# Work Order #023: DLQ Alarm Testing Capability

**Created**: 2025-08-26  
**Category**: 🧪 Testing Infrastructure  
**Priority**: Medium  
**Effort**: S  
**IMP Assignment**: Unassigned

## Problem Statement

To effectively develop, test, and validate the DLQ bot automation system (WO-019, WO-022), we need a reliable way to trigger DLQ alarms on command in testing environments. Currently, there's no systematic way to cause controlled DLQ alarm conditions for testing the automated investigation and remediation workflows.

**Testing Gap**: The DLQ bot automation system requires real DLQ alarm scenarios to validate:
- GitHub issue creation automation (WO-019/WO-022) 
- Claude investigation workflows (WO-020)
- Slack thread integration and notifications
- End-to-end alarm → investigation → remediation pipeline

**Solution**: Enhance the existing Einhorn test worker infrastructure to accept parameters that deliberately cause job failures, triggering DLQ alarms in a controlled manner for testing purposes.

## Acceptance Criteria

### Phase 1: Einhorn Test Worker Enhancement
- [ ] **Parameter Support**: Modify Einhorn test worker to accept failure-inducing parameters
- [ ] **Exception Modes**: Implement different types of exceptions (timeout, runtime error, resource exhaustion, etc.)
- [ ] **Queue Targeting**: Allow specification of which queues should receive the failing jobs
- [ ] **Failure Scenarios**: Support various failure patterns that mirror real production issues

### Phase 2: Command Interface
- [ ] **CLI Integration**: Add command-line interface for triggering test DLQ scenarios
- [ ] **Parameter Validation**: Ensure only valid test environments can trigger intentional failures
- [ ] **Safety Guards**: Prevent accidental triggering in production environments
- [ ] **Documentation**: Clear usage instructions for different testing scenarios

### Phase 3: Testing Integration
- [ ] **DLQ Bot Validation**: Use controlled failures to test GitHub issue creation automation
- [ ] **Investigation Testing**: Validate that investigation workflows can handle test scenarios
- [ ] **Alarm Pattern Testing**: Ensure test failures generate appropriate DataDog alarm patterns
- [ ] **Recovery Testing**: Verify that test jobs can be properly cleared/recovered after testing

## Technical Context

**Repositories**: 
- **intercom**: Contains Einhorn test worker infrastructure and job processing logic
- **event-management-system**: May contain related worker coordination and alarm handling

**Existing Infrastructure**: 
- Einhorn test worker system already exists for testing worker functionality
- Current system likely sends test messages but doesn't deliberately cause failures
- Integration with Sidekiq job processing and queue management

**Enhancement Approach**:
```ruby
# Example enhancement to Einhorn test worker
class EinhornTestWorker
  def perform(test_type: :success, failure_mode: nil, target_queue: nil)
    case failure_mode
    when :runtime_error
      raise "Intentional test failure for DLQ alarm testing"
    when :timeout
      sleep(300) # Cause timeout
    when :resource_exhaustion
      # Simulate resource issues
    else
      # Normal successful test operation
    end
  end
end
```

**Integration Points**:
- Sidekiq job queues and DLQ processing
- DataDog monitoring and alarm generation  
- Existing Einhorn test worker command infrastructure
- DLQ bot automation system for validation testing

## Resources & References

- **Related Work**: 
  - WO-019/WO-022: DLQ bot GitHub issue creation (needs testing capability)
  - WO-020: Claude investigation workflows (needs real alarm scenarios)
  - WO-012: DLQ bot foundations (overall automation system)

- **Infrastructure**: Existing Einhorn test worker system in intercom repository
- **Testing Context**: Required for validating end-to-end DLQ automation workflows

## Implementation Notes

**Einhorn Test Worker Location**: 
- Likely in intercom repository under job/worker infrastructure
- May have existing CLI command for triggering test scenarios

**Failure Mode Implementation**:
1. **Runtime Exceptions**: Standard Ruby exceptions that cause job failures
2. **Timeout Scenarios**: Long-running operations that exceed timeouts
3. **Resource Issues**: Memory/connection exhaustion patterns
4. **Data Issues**: Invalid data scenarios that cause processing failures

**Safety Considerations**:
- Environment checks to prevent production usage
- Rate limiting to prevent overwhelming test environments  
- Clear logging of intentional test failures
- Easy cleanup/recovery mechanisms

**Command Interface Design**:
```bash
# Example usage patterns
einhorn_test_worker --failure-mode=runtime_error --queue=test-dlq
einhorn_test_worker --failure-mode=timeout --count=5
einhorn_test_worker --failure-mode=resource_exhaustion --target-queue=processing-dlq
```

**Success Metrics**:
- Can reliably trigger DLQ alarms in test environments on command
- Different failure modes generate appropriate alarm patterns
- DLQ bot automation can be validated against controlled failure scenarios
- Test scenarios are isolated and don't impact production systems

---

## Work Order Lifecycle

### Status History
- **2025-08-26**: Created → 03-IN-PROGRESS (immediate need for DLQ bot testing)

### IMP Notes
**Status**: 📅 **NEXT** - DLQ alarm testing capability development

**Immediate Need**: Required for validating DLQ bot automation system development (WO-019, WO-022, WO-020)

**Implementation Scope**: Enhance Einhorn test worker to support controlled failure scenarios:
1. Add parameter support for different failure modes
2. Implement safety guards for test environment usage  
3. Create command interface for triggering test DLQ scenarios
4. Validate integration with DLQ bot automation workflows

**Multi-Repo Development**: 
- **intercom**: Primary repository for Einhorn test worker enhancement
- **event-management-system**: May contain related worker coordination logic
- **infrastructure**: Terraform configuration for DLQ monitoring and alarm infrastructure

**Expected Outcome**: 
- Reliable capability to trigger DLQ alarms for testing automation workflows
- Validation framework for DLQ bot GitHub issue creation and investigation
- Foundation for comprehensive testing of end-to-end DLQ automation system

**Integration**: Essential for testing and validating the broader DLQ automation initiative

**Next Steps**: 
1. Locate existing Einhorn test worker implementation
2. Design parameter interface for failure mode specification
3. Implement controlled failure scenarios with appropriate safety guards
4. Test integration with DLQ bot automation workflows

---
*Work Order #023 - Forest Manufacturing System*