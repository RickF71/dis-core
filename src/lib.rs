//! TRIAD
//! Core structural reality of DIS.

// ============================================================
// CORE — Authority & Truth
// ============================================================

// Structural spine (no behavior)
pub mod spine;

// Process-local runtime state (no authority)
pub mod context;

// Identity & continuity
pub mod identity;

// Sealed capability containers
pub mod capsule;

// Persistent storage substrate
pub mod store;

// Domain ontology & artifacts
pub mod domain;

// Canonical ID types
pub mod id;

// ============================================================
// OBSERVER — Awareness & Witness
// ============================================================

// Projection & observation layer
pub mod taiji;
pub mod chat;


// ============================================================
// RUNTIME — Process Composition (No Authority)
// ============================================================

// Binary entry composition
pub mod app;


// ============================================================
// NODE-FACING ADAPTERS (Translation Only)
// ============================================================

// Read-only HTTP views
pub mod api;

// WebSocket adapters (observe / command / totem)
pub mod ws;

// Transport overlays (e.g. JikkaPipe)
pub mod runtime;
