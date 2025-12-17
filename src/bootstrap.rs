use std::fs;
use std::collections::HashMap;
use serde::{Serialize,Deserialize};

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ColorDef {
    pub hex: String,
}


#[derive(Debug, Deserialize)]
pub struct BootstrapColors {
    #[allow(dead_code)]
    pub version: String,
    pub colors: HashMap<String, ColorDef>,
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
