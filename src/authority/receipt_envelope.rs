use crate::authority::types::*;

#[derive(Debug, Clone)]
pub struct ReceiptEnvelope {
    pub core: Receipt,
    pub seals: Vec<ReceiptSeal>,
    pub witnesses: Vec<WitnessRef>,
}

pub struct ReceiptEnvelopeBuilder {
    core: Receipt,
    seals: Vec<ReceiptSeal>,
    witnesses: Vec<WitnessRef>,
}

impl ReceiptEnvelopeBuilder {
    pub fn new(core: Receipt) -> Self {
        Self {
            core,
            seals: Vec::new(),
            witnesses: Vec::new(),
        }
    }

    pub fn add_seal(mut self, seal: ReceiptSeal) -> Self {
        self.seals.push(seal);
        self
    }

    pub fn add_witness(mut self, witness: WitnessRef) -> Self {
        self.witnesses.push(witness);
        self
    }

    pub fn build(self) -> ReceiptEnvelope {
        ReceiptEnvelope {
            core: self.core,
            seals: self.seals,
            witnesses: self.witnesses,
        }
    }
}
