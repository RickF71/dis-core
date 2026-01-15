// src/config/ports.rs

pub fn dis_core_port() -> u16 {
    std::env::var("DIS_CORE_PORT")
        .ok()
        .and_then(|v| v.parse().ok())
        .unwrap_or(8710)
}
