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

mod model_hub;
mod util;

use std::env;

use model_hub::huggingface_hub::Huggingface;
use model_hub::model_hub_trait::{ModelHub, MODEL_LOCAL_PATH};
use util::logging;

fn main() {
    logging::init();

    let model_hub_name = env::var("MODEL_HUB_NAME").unwrap();
    let model_id = env::var("MODEL_ID").unwrap();

    let hub = get_model_hub(&model_hub_name, MODEL_LOCAL_PATH);
    hub.download_model(&model_id);
}

fn get_model_hub(hub_name: &str, model_path: &str) -> Box<dyn ModelHub> {
    match hub_name {
        "Huggingface" => Box::new(Huggingface::new(model_path)),
        _ => panic!("unsupported model hub"),
    }
}
