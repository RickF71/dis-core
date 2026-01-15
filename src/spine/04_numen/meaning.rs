use std::collections::HashMap;

#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub struct MeaningId(pub String);

#[derive(Debug, Clone)]
pub struct Meaning {
    pub id: MeaningId,
    pub description: String,

    /// Domain-defined semantic tags
    pub tags: HashMap<String, String>,
}
