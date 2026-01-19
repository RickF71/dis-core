// src/runtime/record_test.rs

use crate::runtime::record::record_commit;
use crate::runtime::commit::CommitKind;
use crate::runtime::coord6::Coord6;
use crate::store::{Store, artifact::Artifact};

#[tokio::test]
async fn record_commit_writes_artifact() {
    // 1. Open the store
    let store = Store::open().await.expect("store opens");

    // 2. Create a Coord6 (already advanced)
    let coord6 = Coord6 {
        n: 0,
        a: 1,
        t: 0,
        nu: 0,
        l: 0,
        c: 0,
    };

    // 3. Create an artifact
    let artifact = Artifact {
        id: "test-artifact-1".to_string(),
        observed_at: None,
        domain: "domain.test".to_string(),
        store: "store.test".to_string(),
        kind: "receipt".to_string(),
        action: "test.memory.append.v1".to_string(),
        coord6,
        reason: "allow:test".to_string(),
        refs: serde_json::Value::Null,
    };

    // 4. Record the commit
    record_commit(
        &store,
        CommitKind::MemoryAppend,
        artifact.clone(),
    )
    .await
    .expect("record succeeds");

    // 5. Read it back
    let fetched = store
        .get(&artifact.id)
        .await
        .expect("read succeeds");

    // 6. Assert it exists
    assert!(fetched.is_some(), "artifact was recorded");

    let fetched = fetched.unwrap();
    assert_eq!(fetched.id, artifact.id);
    assert_eq!(fetched.domain, artifact.domain);
}
