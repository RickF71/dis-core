pub mod artifact;

use std::path::{Path, PathBuf};
use tokio::fs::{self, File, OpenOptions};
use tokio::io::{AsyncBufReadExt, AsyncWriteExt, BufReader};

use artifact::Artifact;

#[derive(Clone)]
pub struct Store {
    dir: PathBuf,
}

impl Store {
    pub async fn open() -> std::io::Result<Self> {
        let dir = std::env::var("DIS_STORE_DIR")
            .map(PathBuf::from)
            .unwrap_or_else(|_| {
                let home = std::env::var("HOME").unwrap_or_else(|_| ".".to_string());
                PathBuf::from(home).join(".dis").join("store")
            });

        fs::create_dir_all(&dir).await?;
        Ok(Self { dir })
    }

    fn artifacts_log_path(&self) -> PathBuf {
        self.dir.join("artifacts.ndjson")
    }

    pub async fn append(&self, a: &Artifact) -> std::io::Result<()> {
        let mut f = OpenOptions::new()
            .create(true)
            .append(true)
            .open(self.artifacts_log_path())
            .await?;

        let line = serde_json::to_string(a)
            .map_err(|e| std::io::Error::new(std::io::ErrorKind::InvalidData, e))?;
        f.write_all(line.as_bytes()).await?;
        f.write_all(b"\n").await?;
        f.flush().await?;
        Ok(())
    }

    pub async fn get(&self, id: &str) -> std::io::Result<Option<Artifact>> {
        let p = self.artifacts_log_path();
        if !Path::new(&p).exists() {
            return Ok(None);
        }
        let f = File::open(&p).await?;
        let mut lines = BufReader::new(f).lines();

        while let Some(line) = lines.next_line().await? {
            if line.trim().is_empty() { continue; }
            let a: Artifact = match serde_json::from_str(&line) {
                Ok(v) => v,
                Err(_) => continue,
            };
            if a.id == id {
                return Ok(Some(a));
            }
        }
        Ok(None)
    }

    pub async fn tail(
        &self,
        domain: Option<&str>,
        after: Option<&str>,
        limit: usize,
    ) -> std::io::Result<Vec<Artifact>> {
        let p = self.artifacts_log_path();
        if !Path::new(&p).exists() {
            return Ok(vec![]);
        }

        let f = File::open(&p).await?;
        let mut lines = BufReader::new(f).lines();

        let mut out = Vec::new();
        let mut seen_after = after.is_none();

        while let Some(line) = lines.next_line().await? {
            if line.trim().is_empty() { continue; }
            let a: Artifact = match serde_json::from_str(&line) {
                Ok(v) => v,
                Err(_) => continue,
            };

            if let Some(d) = domain {
                if a.domain != d { continue; }
            }

            if let Some(after_id) = after {
                if !seen_after {
                    if a.id == after_id {
                        seen_after = true;
                    }
                    continue;
                }
            }

            out.push(a);
        }

        if out.len() > limit {
            Ok(out.split_off(out.len() - limit))
        } else {
            Ok(out)
        }
    }
}
