// $rPg(Rust/Parsing/Rust token parsing, rust, parser)
// $~ A compact parser example.
pub fn parse(value: &str) -> Option<&str> { (!value.is_empty()).then_some(value) }
