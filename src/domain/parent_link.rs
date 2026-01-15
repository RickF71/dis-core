use crate::id::DomainId;

#[derive(Debug, Clone)]
pub struct ParentLink {
    pub parent_domain: DomainId,
    pub contact: ParentContact,
}

#[derive(Debug, Clone)]
pub enum ParentContact {
    Mailbox(String),
    Endpoint(String),
}
