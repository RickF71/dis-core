#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub struct IdentityId(String);

impl IdentityId {
    pub fn new(raw: impl Into<String>) -> Self {
        Self(raw.into())
    }
}

impl IdentityRef {
    pub fn actor(raw: impl Into<String>) -> Self {
        IdentityRef::Actor {
            id: IdentityId::new(raw),
        }
    }
}

#[derive(Debug, Clone)]
pub enum IdentityRef {
    Actor { id: IdentityId },
    Seat  { id: IdentityId },
    Totem { id: IdentityId },
}
