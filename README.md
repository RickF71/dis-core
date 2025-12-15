# DIS-Core (Layer6 Model)

This repository contains the canonical DIS-Core implementation,
written in Rust and based on the deterministic Layer6 model.

DIS-Core is responsible for:
- advancing global ticks
- committing atomic state transitions
- validating contracts
- persisting authoritative facts

Meaning, interpretation, and life emerge above this layer.

Earlier implementations are deprecated and non-authoritative.
