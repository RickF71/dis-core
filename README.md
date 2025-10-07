# 🕊️ DIS-CORE  
**Direct Individual Sovereignty — Core Constitutional Layer**

---

## 📜 Overview

**DIS-CORE** is the foundational layer of the *Direct Individual Sovereignty* (DIS) framework — a trust-anchored system that allows individuals and domains to operate under verifiable, self-declared authority.  

This repository contains the **canonical DIS-CORE v1.0** schema and the **DIS-PERSONAL v0.5.1-core** implementation — the first self-verifying node that boots under its own constitution and proves its integrity at runtime.

---

## 🚀 Features

- 🔐 **Frozen Constitutional Core** — `dis-core.v1.yaml` defines the unalterable foundation of DIS law.  
- 🔏 **Cryptographic Verification** — the runtime computes and logs a SHA-256 hash of its Core schema at startup.  
- 🌐 **Network Sovereignty Runtime** — Go-based service providing a minimal sovereign node capable of networked operation.  
- 🧱 **Configurable Integrity Policy** — core verification is on by default (`--verify-core=true`), but can be skipped for NOTECH or test builds.  
- 🪶 **Zero-Dependency Design** — built entirely on Go’s standard library and a local SQLite store.  

---

## 🧭 Version Lineage

| Component | Version | Description |
|------------|----------|-------------|
| **DIS-CORE** | `v1.0` | Frozen schema (constitutional baseline) |
| **DIS-PERSONAL** | `v0.5.1-core` | First self-verifying node bound to DIS-CORE v1.0 |
| **Hash Verification** | Enabled | Startup integrity check of `schemas/dis-core.v1.yaml` |

---

## ⚙️ Quickstart

### **1️⃣ Build and Run**

```bash
go run main.go --serve
