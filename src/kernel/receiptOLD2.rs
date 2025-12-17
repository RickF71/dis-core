use super::decision::Decision;
use super::identity::IdentityRef;
use dis_core::domain::domain_id::DomainId;

#[derive(Debug, Clone)]
pub struct Receipt {
    pub receipt_id: String,
    pub domain: DomainId,
    pub decision: Decision,
    pub identity: IdentityRef,
    pub signature: String, // placeholder
}

impl Receipt {
    pub fn mint(
        domain: &DomainId,
        decision: Decision,
        identity: &IdentityRef,
    ) -> Self {
        Self {
            receipt_id: "rcpt-placeholder".into(),
            domain: domain.clone(),
            decision,
            identity: identity.clone(),
            signature: "unsigned".into(),
        }
    }
}
