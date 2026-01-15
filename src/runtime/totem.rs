use crate::runtime::coord6::Coord6;
use crate::runtime::commit::CommitKind;

#[derive(Debug, serde::Serialize)]
pub struct Receipt {
    pub id: String,
    pub coord6: Coord6,
    pub ts: u64,
    pub kind: CommitKind,
    pub intent: String,
}

pub struct Totem {
    coord6: Coord6,
}

impl Totem {
    pub fn new() -> Self {
        Self {
            coord6: Coord6::zero(),
        }
    }

    pub fn commit(&mut self, kind: CommitKind, intent: String) -> Receipt {
        // Advance structural coordinate FIRST
        self.coord6.advance(kind);

        let ts = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .expect("time went backwards")
            .as_secs();

        let id = format!(
            "rcpt-{}-{}-{}-{}-{}-{}",
            self.coord6.n,
            self.coord6.a,
            self.coord6.t,
            self.coord6.nu,
            self.coord6.l,
            self.coord6.c,
        );

        Receipt {
            id,
            coord6: self.coord6,
            ts,
            kind,
            intent,
        }
    }
}
