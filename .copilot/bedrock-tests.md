# /task Add BedrockBootstrap Test Suite
# Objective: Create a full Go test suite for BedrockBootstrap, LocalCapsule, and KnowThyselfBedrock.
# These tests ensure the 1D root domain `null` is created exactly once, only with human approval,
# and that BedrockBootstrap behaves correctly in all startup conditions.

# Requirements:
# - All tests run against the dis_test database.
# - Use the existing DB harness test helpers.
# - Each test should isolate state using the test DB or a transaction rollback pattern.
# - Use mocks/fakes where needed (particularly for the capsule interface).
# - Place tests in appropriate locations under internal/core/identity, internal/discapsule, and cmd/dis-core/bootstrap.

# ===================================================================
# Test Suite 1: Environment and Precondition Tests
# ===================================================================

=== create: internal/core/identity/bedrock_test.go ===
Add tests:
- TestBedrock_NoDomainsTable: drop the domains table and ensure RunBedrockBootstrap returns nil and logs skipping.
- TestBedrock_DomainsTableExistsButEmpty: create domains table empty, ensure capsule called and null created.

# ===================================================================
# Test Suite 2: Root Domain Detection
# ===================================================================

Add tests:
- TestBedrock_NullExists: insert domain {name:'null'}, ensure bootstrap is no-op and capsule not called.
- TestBedrock_LegacyDomainNullExists: insert domain {name:'domain.null'}, ensure no-op and capsule not called.

# ===================================================================
# Test Suite 3: Capsule Interaction
# ===================================================================

Add tests with a mock Capsule implementation:
- TestBedrock_CapsuleApproves: fake capsule returns Approved=true; ensure null is created.
- TestBedrock_CapsuleDeclines: fake capsule returns Approved=false; ensure error and no null created.
- TestBedrock_CapsuleBadInput: simulate unexpected input; ensure error and no null created.

# ===================================================================
# Test Suite 4: KnowThyselfBedrock Behavior
# ===================================================================

Add tests:
- TestKnowThyselfBedrock_CreatesNull: ensure a direct call creates domain 'null' exactly once.
- TestKnowThyselfBedrock_Idempotent: ensure calling twice does not duplicate domains.
- TestKnowThyselfBedrock_RejectsUnapprovedGrant: ensure error when Approved=false.

# ===================================================================
# Test Suite 5: Integration Tests
# ===================================================================

Add tests (location: cmd/dis-core/bootstrap/bedrock_integration_test.go):

- TestBedrockBootstrap_EmptyWorld:
    Setup empty schema, mock capsule to auto-approve.
    Expect domain 'null' created, no other domains, no seats, no receipts.

- TestBedrockBootstrap_RunTwice:
    Run bootstrap twice in a row; second run is a no-op.

- TestBedrockBootstrap_AfterBedrock_KnowThyselfStillWorks:
    After bedrock, simulate login/establish or invite/accept.
    Ensure KnowThyselfAtomic works normally atop the bedrock world.

# ===================================================================
# Test Suite 6: Edge Cases
# ===================================================================

Add tests:
- TestBedrockBootstrap_DBInsertFailure: simulate failure inserting null; ensure rollback and no domain created.
- TestBedrockBootstrap_NoReceiptsTable: remove receipts table; ensure bedrock still succeeds (receipts not required yet).
- TestBedrockBootstrap_ConcurrentRuns: two goroutines racing BedrockBootstrap; ensure only one succeeds and no duplicates.

# ===================================================================
# Implementation Notes
# ===================================================================

- Use the existing dis_test DB harness.
- Provide a mock Capsule type in test files that implements PerformBedrockAuth.
- Avoid modifying production code unless strictly necessary for testability (e.g., dependency injection of Capsule).
- Where logs are expected, use the test logger or capture stdout.
- All SQL checks must verify canonical name 'null', not 'domain.null'.

# END OF TASK
