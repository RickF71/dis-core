#[derive(Debug, Clone)]
pub struct PerceptionContext {
    pub viewport: String,
    pub modality: String, // mouse, touch, voice, etc.
}
