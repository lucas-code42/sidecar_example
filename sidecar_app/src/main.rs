use base64::{Engine as _, engine::{general_purpose}};

fn main() {
    let arg = std::env::args().nth(1).expect("no argument given");
    let b64 = general_purpose::STANDARD.encode(arg);

    println!("{}",b64);
}
