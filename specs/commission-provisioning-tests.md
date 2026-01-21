# Commission Provisioning Test Matrix

Generated from: `specs/commission-provisioning.yaml`

## Test Categories

### 1. Transition Tests (Happy Path)

| Transition | From | To | Event | Guards | Test Cases |
|------------|------|-----|-------|--------|------------|
| create_commission | initial | active | create | is_orc | ORC creates commission with title only |
| | | | | | ORC creates commission with title and description |
| update_commission | active | active | update | - | Update title only |
| | | | | | Update description only |
| | | | | | Update both title and description |
| | | | | | Update with empty values (no-op) |
| start_commission | active | active | start | is_orc, not_in_orc_source | ORC starts commission, creates workspace |
| | | | | | ORC starts commission with custom workspace path |
| launch_commission | active | active | launch | is_orc, not_in_orc_source | Launch creates workspace and groves directory |
| | | | | | Launch moves groves to standard location |
| | | | | | Launch writes grove configs |
| | | | | | Launch with --tmux creates session |
| | | | | | Launch without --tmux skips TMux |
| | | | | | Launch is idempotent (run twice, same result) |
| pin_commission | active | active | pin | exists | Pin active commission |
| | | | | | Pin already-pinned commission (no-op) |
| unpin_commission | active | active | unpin | exists | Unpin pinned commission |
| | | | | | Unpin unpinned commission (no-op) |
| complete_commission | active | complete | complete | not_pinned | Complete unpinned commission |
| archive_active_commission | active | archived | archive | not_pinned | Archive active unpinned commission |
| archive_complete_commission | complete | archived | archive | not_pinned | Archive completed commission |
| pause_commission | active | paused | update | status_change_to_paused | Pause active commission |
| resume_commission | paused | active | update | status_change_to_active | Resume paused commission |
| delete_active_commission | active | deleted | delete | no_dependents_or_force | Delete commission with no dependents |
| | | | | | Delete commission with --force (succeeds, orphans dependents) |
| delete_archived_commission | archived | deleted | delete | no_dependents_or_force | Delete archived commission |

### 2. Guard Failure Tests (Negative Cases)

| Guard | Transition | Test Case | Expected Error |
|-------|------------|-----------|----------------|
| is_orc | create_commission | IMP blocked from creating commission | "IMPs cannot create commissions - only ORC can create commissions" |
| is_orc | start_commission | IMP blocked from starting commission | "IMPs cannot start commissions - only ORC can start commissions" |
| is_orc | launch_commission | IMP blocked from launching commission | "IMPs cannot launch commissions - only ORC can launch commissions" |
| not_in_orc_source | start_commission | Start from ORC source directory blocked | "Cannot run this command from ORC source directory" |
| not_in_orc_source | launch_commission | Launch from ORC source directory blocked | "Cannot run this command from ORC source directory" |
| not_pinned | complete_commission | Complete pinned commission (error) | "Cannot complete pinned commission {id}. Unpin first with: orc commission unpin {id}" |
| not_pinned | archive_active_commission | Archive pinned commission (error) | "Cannot archive pinned commission {id}. Unpin first with: orc commission unpin {id}" |
| no_dependents_or_force | delete_active_commission | Delete commission with groves (error without --force) | "Commission has {count} groves and {count} shipments. Use --force to delete anyway" |
| no_dependents_or_force | delete_active_commission | Delete commission with shipments (error without --force) | "Commission has {count} groves and {count} shipments. Use --force to delete anyway" |
| exists | pin_commission | Pin non-existent commission (error) | "Commission {id} not found" |

### 3. Edge Case Tests

| Category | Test Case | Expected Behavior |
|----------|-----------|-------------------|
| TMux | Start with existing TMux session | Error (session exists) |
| Idempotency | Launch same commission twice | Same result (idempotent) |
| No-op | Complete already-complete commission | No-op or error |
| No-op | Pin already-pinned commission | No-op |
| No-op | Unpin unpinned commission | No-op |

### 4. Invariant Tests (Property-Based)

| Invariant | Property | Test Strategy |
|-----------|----------|---------------|
| id_format | Commission IDs follow COMM-XXX format | Regex check on all created commissions |
| id_unique | Commission IDs are unique | Attempt duplicate creation |
| status_valid | Status is one of valid values | Attempt invalid status via direct DB |
| pinned_blocks_terminal | Pinned commissions cannot be in terminal states | Attempt to pin then complete/archive |
| completed_has_timestamp | Complete status requires completed_at | Check completed_at after complete transition |
| timestamps_ordered | completed_at >= created_at | Check timestamp ordering after complete |

### 5. State Reachability Tests

| State | Reachable Via | Test |
|-------|---------------|------|
| initial | - | Default state before any commission exists |
| active | create from initial | Create commission, verify status = "active" |
| paused | pause from active | Pause commission, verify status = "paused" |
| complete | complete from active | Complete commission, verify status = "complete" |
| archived | archive from active/complete | Archive commission, verify status = "archived" |
| deleted | delete from active/archived | Delete commission, verify removed from DB |

---

## Test Implementation Priority

### P0 - Core Happy Path
1. create_commission (ORC creates with title)
2. complete_commission (unpinned)
3. archive_complete_commission
4. delete_archived_commission

### P1 - Guard Enforcement
1. is_orc guard (IMP blocked)
2. not_pinned guard (pinned blocks complete/archive)
3. no_dependents_or_force guard

### P2 - Infrastructure Provisioning
1. start_commission (workspace creation)
2. launch_commission (idempotent provisioning)
3. TMux session management

### P3 - Edge Cases & Invariants
1. Idempotency tests
2. No-op handling
3. Property-based invariant tests

---

## Mapping to Existing Tests

| FSM Test Case | Existing Shell Test | Status |
|---------------|---------------------|--------|
| ORC creates commission with title only | 01-test-commission-creation.sh:test_basic | Covered |
| ORC creates commission with title and description | 01-test-commission-creation.sh:test_with_description | Covered |
| Archive active unpinned commission | 01-test-commission-creation.sh:test_archive | Covered |
| IMP blocked from creating commission | - | Gap |
| Complete pinned commission (error) | - | Gap |
| Launch is idempotent | - | Gap |
| Pin/Unpin lifecycle | - | Gap |
| Delete with dependents guard | - | Gap |

## Test Gaps Summary

**High Priority Gaps:**
- Agent identity checks (is_orc guard) - no tests
- Pinned state transitions - no tests
- Deletion guards with dependents - no tests
- Launch idempotency - no tests

**Medium Priority Gaps:**
- Pause/resume workflow (rarely used but defined)
- TMux error cases
- Edge case no-ops
