/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

use std::path::{Path, PathBuf};
use std::sync::mpsc::channel;
use std::time::Instant;

use hf_hub::api::{sync::Api, sync::ApiBuilder};
use log::debug;
use log::info;
use threadpool::ThreadPool;

use crate::model_hub::model_hub_trait::CustomError;
use crate::model_hub::model_hub_trait::ModelHub;
use crate::model_hub::model_hub_trait::MODEL_LOCAL_PATH;
use crate::util::file;

const MAX_CONCURRENT_DOWNLOAD: usize = 8;
const DEFAULT_BRANCH_NAME: &str = "main";

pub struct Huggingface {
    client: Api,
}

impl ModelHub for Huggingface {
    fn info(&self, model_id: &str) -> Result<(Vec<String>, String), CustomError> {
        match self.client.model(model_id.to_string()).info() {
            Ok(value) => {
                debug!("Get model info successfully");
                Ok((
                    value.siblings.into_iter().map(|x| x.rfilename).collect(),
                    value.sha,
                ))
            }
            Err(e) => Err(CustomError::RequestError(e.to_string())),
        }
    }

    fn download_model(&self, model_id: &str) {
        let model_info = self.info(model_id).unwrap();
        let len = model_info.0.len();

        // TODO: download model weights is a synchronous OP, no need to use asynchronous libs like tokio,
        // which can be optimized in the future.
        let pool = ThreadPool::new(MAX_CONCURRENT_DOWNLOAD);
        let (tx, rx) = channel();

        let start = Instant::now();
        // FIXME: once one thread is stuck, the main thread will never exit.
        for file in model_info.0.into_iter() {
            let client = self.client.clone();
            let model_string = String::from(model_id);
            let tx = tx.clone();

            pool.execute(move || {
                debug!("Start to download {}", file);
                client.model(model_string).get(&file).unwrap();
                tx.send(file).unwrap();
            });
        }

        for _ in 0..len {
            let file = rx.recv().unwrap();
            debug!("Download {} successfully", file);
        }
        info!("Download model takes {:?}s", start.elapsed());

        // Copy symlinks to a new folder for easy reference in the inference service.

        let formatted_model_name = model_id.to_string().replace("/", "--");
        let prefix_dir = format!(
            "{}/{}{}",
            MODEL_LOCAL_PATH, "models--", formatted_model_name
        );
        let src_dir = format!("{}/snapshots/{}", prefix_dir, model_info.1);
        // TODO: we only support main branch for now.
        let dest_dir = format!("{}/indices/{}", prefix_dir, DEFAULT_BRANCH_NAME);

        // How the structure looks like:
        // └── 453ed1575b739b5b03ce3758b23befdb0967f40e
        //     ├── LICENSE -> ../../blobs/6634c8cc3133b3848ec74b9f275acaaa1ea618ab
        //     ├── README.md -> ../../blobs/7bed8361ea56b8890ed0ea5c532b4ad11cded8b8
        //     ├── config.json -> ../../blobs/ae477f8df742839e77c615d11b0b1783f6ec2164
        //     ├── generation_config.json -> ../../blobs/cbbb3133034e192527e5321b4c679154e4819ab8
        //     ├── merges.txt -> ../../blobs/20024bfe7c83998e9aeaf98a0cd6a2ce6306c2f0
        //     └── model.safetensors.index.json -> ../../blobs/4bf699d11c6478a4b70fc2adfb405429de22525f
        // FIXME: from the structure, we can't tell the service controller what the sha value is
        // when constructing the Pod container, so we have to copy the symlink again with version name
        // (or branch name) like main, because this is symlink, the cost is acceptable. And the sha is
        // too long, not that friendly. If we find better ways, we should change it.
        file::copy_symlink_dir(Path::new(&src_dir), Path::new(&dest_dir)).unwrap();
        info!("Symlink model weights successfully");
    }
}

impl Huggingface {
    pub fn new(local_path: &str) -> Self {
        let buf = PathBuf::from(local_path);
        Self {
            client: ApiBuilder::new().with_cache_dir(buf).build().unwrap(),
        }
    }
}
