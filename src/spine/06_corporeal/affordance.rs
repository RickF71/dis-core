use crate::spine::numen::MeaningId;

#[derive(Debug, Clone)]
pub struct Affordance {
    pub meaning: MeaningId,
    pub label: String,
    pub hint: Option<String>,
}
