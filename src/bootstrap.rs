use std::fs;
use serde::Deserialize;

#[derive(Debug, Deserialize)]
pub struct BootstrapColors {
    pub colors: std::collections::HashMap<String, String>,
}

#[derive(Debug, Deserialize)]
pub struct BootstrapLayers {
    pub layers: std::collections::HashMap<String, String>,
}

pub fn load_colors() -> BootstrapColors {
    let data = fs::read_to_string("bootstrap/colors.json")
        .expect("missing bootstrap/colors.json");
    serde_json::from_str(&data).expect("invalid colors.json")
}

pub fn load_layers() -> BootstrapLayers {
    let data = fs::read_to_string("bootstrap/layers.json")
        .expect("missing bootstrap/layers.json");
    serde_json::from_str(&data).expect("invalid layers.json")
}
