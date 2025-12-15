use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct DomainId(pub Uuid);

impl DomainId {
    pub fn self_domain() -> Self {
        // Stub: returns a new UUID for now
        DomainId(Uuid::new_v4())
    }
}
