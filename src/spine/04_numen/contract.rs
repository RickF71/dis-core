use super::meaning::MeaningId;

#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub struct ContractId(pub String);

#[derive(Debug, Clone)]
pub struct NumenContract {
    pub id: ContractId,

    /// What this contract applies to
    pub applies_to: MeaningId,

    /// Human-readable explanation
    pub description: String,
}
