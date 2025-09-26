# Work Order #005: DLQ Alarm Review Automation

**Created**: 2025-08-15  
**Category**: 🤖 Automation  
**Priority**: High  
**Effort**: XL  
**IMP Assignment**: IMP-DLQ (active - work tree ml-dlq-alarm-review-automation)

## Problem Statement

The DLQ alarm review process is currently manual and toilsome, requiring significant weekly time investment to review alarms, assess their validity, gather context from multiple sources, and make deletion decisions. This creates operational overhead and delays in maintaining clean monitoring systems.

**Weekly Goal Commitment (FY26 Q3 C1 W3)**: "Review paging queue alarms for one Intercom service" with specific deliverables:
- Generate review issue with list of paging alarms for a single service
- Add worker purpose, recent alarm history  
- Review each alarm and make keep/kill decision
- Downgrade non-valuable DLQ alarms if discovered
- **Create Claude command for future automation and document process**

Based on conversation with **Mark Gorman (TPM)**, there's a strong commitment to automating this process to eliminate "toilsome bullshit" work and improve operational efficiency.

## Acceptance Criteria

### Phase 1: Manual Process Foundation (Weekly Goal)
- [ ] **Single Service Review**: Generate review issue with list of paging alarms for one Intercom service
- [ ] **Context Collection**: Add worker purpose and recent alarm history to review
- [ ] **Decision Framework**: Review each alarm and make keep/kill decision with documented rationale
- [ ] **Immediate Actions**: Downgrade non-valuable DLQ alarms if discovered
- [x] **Automation Foundation**: Create Claude command for future automation and document process

**Phase 1 Progress**:
- [x] **Infrastructure Discovery**: Complete mapping of availability_tier 1-2 queues in Terraform (~85 total)
- [x] **Data Source Integration**: Honeycomb ASG-RAG-Status and intercom-production dataset connectivity 
- [x] **Matching Algorithm**: Queue correlation logic between infrastructure and operational data
- [x] **Decision Engine**: Business-aware recommendation system with team priority mapping
- [x] **Command Interface**: Complete CLI with dry-run, service filtering, and issue generation
- [x] **Safety Measures**: Comprehensive audit trails, confidence scoring, manual confirmation flows
- [x] **Documentation**: Full implementation guide with workflow examples and data structure specs

### Phase 2: Automation Development (Future)
- [ ] **Automated Data Collection**: System pulls data from multiple sources:
  - RAG checker dataset for alarm context and history
  - Terraform files for infrastructure configuration context
  - Incident.io data for incident correlation
  - Additional operational data sources as identified
- [ ] **Intelligent Analysis**: Bot processes alarm data and provides actionable recommendations:
  - Identifies potentially safe-to-delete alarms
  - Provides context and justification for recommendations
  - Flags high-risk alarms requiring human review
- [ ] **Automated Actions**: System can safely delete approved alarms without manual intervention
- [ ] **Review Dashboard**: Interface for reviewing bot recommendations and override decisions
- [ ] **Audit Trail**: Complete logging of all automated decisions and manual overrides
- [ ] **Safety Mechanisms**: Comprehensive safeguards to prevent accidental deletion of critical alarms

## Technical Context

**Key Discussion Points from Mark Gorman:**
- Current manual process is time-consuming and doesn't scale
- Need to eliminate repetitive review work while maintaining safety
- Integration with existing monitoring and infrastructure systems is critical
- Must maintain audit trail for compliance and safety

**Dependencies**: 
- Access to RAG checker dataset and API
- Terraform state and configuration access
- Incident.io API integration
- Existing alarm management system integration

**Data Sources Required**:
- **RAG Checker**: Alarm metadata, historical context, firing patterns
- **Terraform**: Infrastructure configuration, resource relationships
- **Incident.io**: Incident correlation, historical alarm impact
- **Monitoring Systems**: Current alarm states, thresholds, ownership

**Safety Requirements**:
- Dry-run mode for testing recommendations
- Multi-tier approval process for high-risk deletions
- Rollback capability for accidentally deleted alarms
- Integration with existing change management processes

## Implementation Strategy

**Phase 1: Manual Process Foundation (Week 3 Deliverable)**
- Select target Intercom service for initial review
- Generate comprehensive review issue with alarm inventory
- Collect worker purpose and alarm history context
- Execute manual review with documented decision rationale
- Create foundational Claude command for process automation
- Document learnings and process improvements

**Phase 2: Data Collection Architecture (Future)**
- Set up MCP integrations for primary data sources
- Build unified data model for alarm analysis
- Implement data freshness and quality checks

**Phase 3: Analysis Engine (Future)** 
- Develop ML/rule-based system for alarm classification
- Create recommendation engine with confidence scoring
- Build context aggregation from multiple data sources

**Phase 4: Automation Pipeline (Future)**
- Implement automated deletion workflow with safeguards
- Create review interface for human oversight
- Add comprehensive logging and audit trails

**Phase 5: Dashboard & Monitoring (Future)**
- Build review dashboard for bot recommendations
- Implement monitoring for bot performance and accuracy
- Create alerting for system failures or anomalies

## Resources & References

- **Coda Weekly Goal**: "Review paging queue alarms for one Intercom service" (FY26 Q3 C1 W3)
- **Slack Discussion**: Conversation with Mark Gorman (TPM) about automation requirements  
- **Weekly Commitment**: El Presidente's commitment to tackle alarm review automation
- **Existing Systems**: 
  - RAG checker dataset and tooling
  - Current alarm review processes
  - Terraform infrastructure management
  - Incident.io integration patterns

## Risk Assessment

**High Impact Risks**:
- Accidental deletion of critical monitoring alarms
- Data quality issues leading to incorrect recommendations
- Integration failures with external systems

**Mitigation Strategies**:
- Comprehensive testing in staging environment
- Phased rollout with manual oversight
- Circuit breakers for system failures
- Detailed rollback procedures

---

## Work Order Lifecycle

### Status History
- **2025-08-15**: Created → 01-BACKLOG (awaiting IMP assignment)
- **2025-08-15**: IMP-DLQ assigned → 02-IN-PROGRESS (work tree created)
- **2025-08-15**: Moved to 02-NEXT (scheduled for FY26 Q3 C1 W3)
- **2025-08-15**: Moved to 03-IN-PROGRESS → active development started

### IMP Notes
**Status**: 🟡 **INBOX SERVICE REVIEW IN PROGRESS** - Phase 1 execution underway with systematic queue analysis

**Immediate Priority**: Phase 1 (Weekly Goal) must be completed for FY26 Q3 C1 W3

**Key Actions Completed**:
- **2025-08-15**: Work tree ml-dlq-alarm-review-automation created with infrastructure + intercom repos
- **2025-08-15**: CLAUDE.md investigation context established  
- **2025-08-15**: Work order symlink integrated for progress tracking
- **2025-08-18**: ✅ **Infrastructure mapping complete** - Located all availability_tier 1-2 queues in Terraform
- **2025-08-18**: ✅ **ASG-RAG-Status integration** - Successfully matched Terraform queues with Honeycomb operational data
- **2025-08-18**: ✅ **Claude command framework built** - Complete DLQ review command structure created in `.claude/commands/`
- **2025-08-18**: ✅ **Matching logic implemented** - Ruby-based algorithm for queue discovery and data correlation
- **2025-08-18**: ✅ **Decision framework complete** - Business-aware recommendation engine with safety measures
- **2025-08-18**: ✅ **Three-dataset methodology established** - Terraform → asg-rag-status → intercom-production integration
- **2025-08-18**: ✅ **Correct DLQ error analysis** - Implemented `is_dlq_error = true` methodology for accurate error detection
- **2025-08-18**: ✅ **GitHub issue structure finalized** - Clean intro + individual queue comments format validated
- **2025-08-18**: ✅ **Inbox service review executed** - 4 of 6 tier 1 queues completed with comprehensive production data
- **2025-08-18**: ✅ **Production GitHub issue created** - https://github.com/intercom/intercom/issues/425511

**Technical Implementation Details**:
- **Infrastructure Discovery**: Automated scanning of `/infrastructure/prod/services/intercom/` for tier 1-2 queues
- **Data Integration**: Cross-reference with Honeycomb `asg-rag-status` dataset using `queue_names` field matching
- **Activity Analysis**: Integration with `intercom-production` dataset for queue processing metrics
- **Recommendation Engine**: Smart categorization (Keep/Downgrade/Kill) based on team priority and business criticality
- **Command Structure**: Complete CLI interface with dry-run mode, service filtering, and issue generation

**Files Created**:
- `intercom/.claude/commands/dlq-alarm-review.md` - Main command documentation and implementation guide
- `intercom/.claude/commands/dlq-queue-matcher.rb` - Core matching algorithm with demo functionality  
- `intercom/.claude/commands/example-dlq-review.md` - Complete workflow example with sample data

**Key Findings from Investigation**:
- **Queue Distribution**: ~25 tier-1 queues, ~60+ tier-2 queues across inbox/channels/fin/misc services
- **Data Quality**: High match rate between Terraform config and ASG-RAG-Status operational data
- **Regional Architecture**: US/EU/AU deployment pattern with consistent naming conventions
- **Team Ownership**: Clear mapping to responsible teams (infra-platform, inbox, billing, ml, etc.)

**Active Work Tree**: `~/src/worktrees/ml-dlq-alarm-review-automation/`

**Current Status**: 🟡 **INBOX SERVICE REVIEW 67% COMPLETE** - 4/6 tier 1 queues analyzed, systematic methodology proven

**Active GitHub Issue**: [425511 - DLQ Alarm Review: inbox Service](https://github.com/intercom/intercom/issues/425511)
- Clean issue structure with individual queue comments
- Real production data using correct `is_dlq_error` methodology  
- Complete infrastructure context and decision frameworks
- Team-ready for voting and consensus gathering

**Phase 1 Progress**:
- ✅ **Queue Analysis Method** - Three-dataset integration proven reliable
- ✅ **Production Data Quality** - 423M+ jobs analyzed with 0% error rates on tier 1 queues  
- ✅ **Infrastructure Integration** - Exact line numbers, costs, capacity data captured
- ✅ **Team Review Structure** - Individual comments enable focused decision-making

**Next Steps**: 
1. ✅ Create foundational Claude command and process documentation
2. ✅ Execute inbox service review - 4/6 tier 1 queues complete 
3. ✅ Generate review issue - GitHub issue 425511 active
4. **Complete remaining 2 tier 1 + 8 tier 2 inbox queues** - Full service coverage
5. **Gather team consensus** - Use GitHub issue voting for decisions
6. **Apply Terraform changes** - Implement approved paging modifications
7. **Document lessons learned** - Capture process improvements for Phase 2 automation

---
*Work Order #005 - Forest Manufacturing System*
