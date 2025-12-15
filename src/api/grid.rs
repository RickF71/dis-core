// src/api/grid.rs

use dis_spine::spine::Layer6;
use serde::Serialize;

#[derive(Debug, Clone, Serialize)]
pub struct GridCell {
    pub from: Layer6,
    pub to: Layer6,
}

#[derive(Debug, Clone, Serialize)]
pub struct FullGrid {
    pub layers: Vec<Layer6>,
    pub cells: Vec<GridCell>,
}


pub fn generate_full_grid() -> FullGrid {
    let layers = vec![
        Layer6::Nullus,
        Layer6::Aether,
        Layer6::Terra,
        Layer6::Numen,
        Layer6::Lima,
        Layer6::Corporeal,
    ];

    let mut cells = Vec::new();

    for from in &layers {
        for to in &layers {
            cells.push(GridCell {
                from: *from,
                to: *to,
            });
        }
    }

    FullGrid { layers, cells }
}
