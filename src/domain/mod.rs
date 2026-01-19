// src/domain/mod.rs

pub mod actions;
pub mod artifact;
pub mod domain;
pub mod parent_link;
pub mod receipts;
pub mod reflective_surface;
pub mod storage;

// ---- DOMAIN COORDINATE SYSTEM (formerly clock / layer6) ----
pub mod lattice_axis;
pub mod lattice;

// ---- DOMAIN DATA & MECHANICS ----
pub mod payload;
pub mod echo_store;

// ---- PUBLIC DOMAIN IDENTITY ----
pub use domain::Domain;


// ---- NETWORK / HUMAN SURFACES (ADAPTERS) ----
pub mod http;


// ---- RUNTIME (Phase 3+) ----
pub mod runtime;
