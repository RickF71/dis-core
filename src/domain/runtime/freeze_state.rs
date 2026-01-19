// src/domain/state/freeze.rs

use sqlx::{PgPool, Row};

pub enum FreezeState {
    Active,
    Frozen { reason: Option<String> },
}

pub async fn get_freeze_state(
    pool: &PgPool,
    domain_id: &str,
) -> FreezeState {
    let row = sqlx::query(
        r#"
        SELECT frozen, reason
        FROM domain_freeze
        WHERE domain_id = $1
        "#
    )
    .bind(domain_id)
    .fetch_optional(pool)
    .await;

    if let Ok(Some(row)) = row {
        let frozen: bool = row.try_get("frozen").unwrap_or(false);
        if frozen {
            let reason: Option<String> = row.try_get("reason").ok();
            return FreezeState::Frozen { reason };
        }
    }
    FreezeState::Active
}
