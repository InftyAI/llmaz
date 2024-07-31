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

use std::os::unix::fs::symlink;
use std::path::Path;
use std::{fs, io};

pub fn copy_symlink_dir(src: &Path, dest: &Path) -> io::Result<()> {
    if dest.exists() {
        fs::remove_dir_all(dest)?;
    }
    fs::create_dir_all(dest)?;

    let entries = fs::read_dir(src)?;
    for entry in entries {
        let entry = entry?;
        let src_path = fs::read_link(entry.path()).unwrap();
        let dest_path = dest.join(entry.file_name());

        symlink(&src_path, dest_path)?;
    }
    Ok(())
}
