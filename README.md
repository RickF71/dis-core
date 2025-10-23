# 🧾 DIS-Core Changelog

---

## 🧩 **v0.9.8 — “Recreate & Rise”**
**Date:** 2025-10-22
**Status:** ✅ Stable / Bootstrapping milestone

### ✨ Highlights

#### 🔧 Full Database Recreation Pipeline
- Added `--recreate-db` flag to drop and rebuild the database automatically.
- Uses `DISCORE_ADMIN_DSN` for privileged operations and `DISCORE_DSN` for normal runtime.
- Automatically creates the `dis_user` role with password and full privileges.
- Terminates lingering connections before dropping the DB to prevent lock errors.
- Recreates schema from scratch without requiring manual SQL setup.

#### 🌱 Environment Configuration
- Introduced `.env` file support via [`github.com/joho/godotenv`](https://github.com/joho/godotenv).
- Added persistent environment variables for configuration:
  - `DISCORE_ADMIN_DSN`
  - `DISCORE_DSN`
  - `DISCORE_DB_NAME`
  - `DISCORE_APP_USER`
  - `DISCORE_APP_PASS`
- `.env` is loaded automatically at startup—no more manual `export` needed.

#### 🧱 Schema Consistency
- Unified all table creation timestamps under:
  ```sql
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
