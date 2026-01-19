use std::collections::HashSet;
use std::rc::Rc;
use std::cell::RefCell;

use crate::authority::types::*;
use crate::authority::errors::*;
use crate::authority::gate::{
    FreezeStateReader,
    FreezeStateWriter,
    CommitWriter,
    ReceiptWriter,
    IdentityBinder,
};

/// Inner mutable fake backend (pure state)
#[derive(Default)]
pub struct FakeAuthorityBackend {
    // (domain, scope) frozen
    frozen: HashSet<(String, String)>,

    // accepted commits (targets)
    commits: Vec<String>,

    // all receipts appended
    pub receipts: Vec<Receipt>,

    // valid actors (Nullus)
    valid_actors: HashSet<String>,
}

impl FakeAuthorityBackend {
    pub fn with_actor(mut self, id: &str) -> Self {
        self.valid_actors.insert(id.to_string());
        self
    }
    
    pub fn last_receipt(&self) -> Option<&Receipt> {
        self.receipts.last()
    }
}

/// Shared backend handle used by AuthorityKernel
pub type SharedBackend = Rc<RefCell<FakeAuthorityBackend>>;

// ==============================
// IdentityBinder (Nullus)
// ==============================

impl IdentityBinder for SharedBackend {
    fn validate_actor(&self, actor: &ActorRef) -> Result<(), AuthorityError> {
        let backend = self.borrow();
        if backend.valid_actors.contains(&actor.id) {
            Ok(())
        } else {
            Err(AuthorityError::MissingIdentityBinding)
        }
    }
}

// ==============================
// FreezeStateReader
// ==============================

impl FreezeStateReader for SharedBackend {
    fn is_frozen(
        &self,
        domain: &DomainRef,
        scope: &Scope,
    ) -> Result<bool, AuthorityError> {
        let backend = self.borrow();
        Ok(backend
            .frozen
            .contains(&(domain.id.clone(), scope.key.clone())))
    }
}

// ==============================
// FreezeStateWriter
// ==============================

impl FreezeStateWriter for SharedBackend {
    fn apply_freeze_op(
        &mut self,
        domain: &DomainRef,
        scope: &Scope,
        op: FreezeOp,
        _ttl_hint_seconds: Option<u64>,
    ) -> Result<String, AuthorityError> {
        let mut backend = self.borrow_mut();
        let key = (domain.id.clone(), scope.key.clone());

        match op {
            FreezeOp::Freeze => {
                backend.frozen.insert(key);
            }
            FreezeOp::Unfreeze | FreezeOp::BreakGlass => {
                backend.frozen.remove(&key);
            }
        }

        Ok("freeze-state-ref".to_string())
    }
}

// ==============================
// CommitWriter
// ==============================

impl CommitWriter for SharedBackend {
    fn accept_commit(
        &mut self,
        _domain: &DomainRef,
        _scope: &Scope,
        _intent: &Intent,
        target: &str,
    ) -> Result<String, AuthorityError> {
        let mut backend = self.borrow_mut();
        backend.commits.push(target.to_string());
        Ok("commit-ref".to_string())
    }
}

// ==============================
// ReceiptWriter (critical)
// ==============================

impl ReceiptWriter for SharedBackend {
    fn append_receipt(&mut self, receipt: &Receipt) -> Result<(), AuthorityError> {
        let mut backend = self.borrow_mut();
        backend.receipts.push(receipt.clone());
        Ok(())
    }
}


use crate::authority::gate::ReceiptIdMint;
use crate::authority::types::ReceiptRef;

#[derive(Default)]
pub struct TestReceiptMinter {
    next: u64,
}

impl ReceiptIdMint for TestReceiptMinter {
    fn mint_receipt_id(&mut self) -> ReceiptRef {
        self.next += 1;
        ReceiptRef {
            id: format!("rcpt-{}", self.next),
        }
    }
}

use crate::authority::gate::ReceiptParentReader;

impl ReceiptParentReader for SharedBackend {
    fn get_receipt(&self, id: &ReceiptRef) -> Result<Option<Receipt>, AuthorityError> {
        let b = self.borrow();
        Ok(b.receipts.iter().find(|r| &r.id == id).cloned())
    }
}
