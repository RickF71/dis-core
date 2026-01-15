// src/domain/actions/freeze.rs
//
// Authority primitive: domain.freeze.v1
// This file mutates sovereign state and emits canonical receipts.

use sqlx::{PgPool, Postgres, Transaction};
use crate::domain::receipts::{new_receipt_id, hash_payload};

pub struct FreezeRequest {
    pub domain_id: String,
    pub actor_id: String,
    pub reason: Option<String>,
}

pub struct FreezeResult {
    pub receipt_id: String,
}

pub async fn freeze_domain(
    pool: &PgPool,
    req: FreezeRequest,
) -> Result<FreezeResult, sqlx::Error> {

    let mut tx: Transaction<Postgres> = pool.begin().await?;

    // 1. Write freeze state
    sqlx::query(
        r#"
        INSERT INTO domain_freeze (domain_id, frozen, reason, frozen_by)
        VALUES ($1, true, $2, $3)
        ON CONFLICT (domain_id)
        DO UPDATE SET
            frozen = true,
            reason = EXCLUDED.reason,
            frozen_by = EXCLUDED.frozen_by,
            updated_at = now()
        "#
    )
    .bind(&req.domain_id)
    .bind(&req.reason)
    .bind(&req.actor_id)
    .execute(&mut *tx)
    .await?;

    // 2. Mint receipt
    let receipt_id = new_receipt_id();
    let payload = format!(
        "freeze:{}:{:?}",
        req.domain_id,
        req.reason,
    );

    sqlx::query(
        r#"
        INSERT INTO receipts (id, domain_id, actor_id, action, payload_hash)
        VALUES ($1, $2, $3, 'domain.freeze.v1', $4)
        "#
    )
    .bind(&receipt_id)
    .bind(&req.domain_id)
    .bind(&req.actor_id)
    .bind(&hash_payload(&payload))
    .execute(&mut *tx)
    .await?;

    tx.commit().await?;

    Ok(FreezeResult { receipt_id })
}
