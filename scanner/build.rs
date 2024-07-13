use std::env;
use std::error::Error;
use std::path::PathBuf;

fn main() -> Result<(), Box<dyn Error>> {
    let out_dir = PathBuf::from(env::var("OUT_DIR").unwrap());

    let proto_files = ["../api/service.proto"];
    let proto_include_dirs = ["../api"];

    // tonic_build::configure()
    //     .file_descriptor_set_path(out_dir.join("scanner_descriptor.bin"))
    //     .compile(&proto_files, &proto_include_dirs)?;

    // Ok(())

    Ok(())
}