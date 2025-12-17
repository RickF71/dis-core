use warp::Filter;
use serde::Serialize;
use uuid::Uuid;

use crate::kernel::{Kernel, CommitResult, Decision, IdentityRef};

use dis_core::domain::domain_id::DomainId;
use dis_core::spine::capsule::Capsule;

#[derive(Serialize)]
struct CommitReply {
    outcome: &'static str,
    decision_ref: String,
}

pub fn routes(
    kernel: Kernel,
) -> impl Filter<Extract = impl warp::Reply, Error = warp::Rejection> + Clone {
    let kernel_filter = warp::any().map(move || kernel.clone());

    let commit_test = warp::path!("api" / "commit" / "test")
        .and(warp::post())
        .and(kernel_filter)
        .map(|kernel: Kernel| {
            let domain = DomainId(Uuid::nil());

            let capsule = Capsule::<()>::empty();

            let identity = IdentityRef::actor("actor.test");


            let decision = Decision::deny_by_default(domain.clone());

            let result = kernel.commit(&domain, capsule, &identity, decision);

            let reply = match result {
                CommitResult::Applied { decision_ref } => CommitReply {
                    outcome: "applied",
                    decision_ref,
                },
                CommitResult::Denied { decision_ref } => CommitReply {
                    outcome: "denied",
                    decision_ref,
                },
            };

            warp::reply::json(&reply)
        });

    commit_test
}
