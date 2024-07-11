use std::net::IpAddr;

use async_std::task::block_on;
use rustscan::port_strategy::PortStrategy;
use rustscan::scanner::Scanner;
use tonic::{Request, Response, Status, transport::Server};

use scanner::{ScanPortRequest, ScanPortResponse};
use scanner::port_scanner_service_server::{PortScannerService, PortScannerServiceServer};


pub mod scanner {
    tonic::include_proto!("scanner");

    pub const FILE_DESCRIPTOR_SET: &[u8] =
        tonic::include_file_descriptor_set!("scanner_descriptor");
}

#[derive(Default, Debug, Clone)]
pub struct ScannerService {
    default_timeout: std::time::Duration,
}

impl ScannerService {
    pub fn new(default_timeout: std::time::Duration) -> Self {
        ScannerService { default_timeout }
    }
}

#[tonic::async_trait]
impl PortScannerService for ScannerService {
    async fn scan_ports(&self, request: Request<ScanPortRequest>) -> Result<Response<ScanPortResponse>, Status> {
        let msg = request.get_ref();

        println!("Got a request: {:?}", request);

        let host = match msg.host.parse::<IpAddr>() {
            Ok(host) => { vec![host] }
            Err(_) => { return Err(Status::invalid_argument("Invalid host")) }
        };

        let strategy = PortStrategy::pick(
            &Some(rustscan::input::PortRange {
                start: msg.start_port as u16,
                end: msg.end_port as u16,
            }),
            None,
            rustscan::input::ScanOrder::Random,
        );

        // TODO: by unsafe convert msg.excluded_ports with type Vec<u32> to u16 vector
        let excluded_ports: Vec<u16> = msg.excluded_ports.iter().map(|&x| x as u16).collect();

        let _scanner =
            Scanner::new(
                &host,
                3_000,
                self.default_timeout,
                1,
                false,
                strategy,
                false,
                excluded_ports,
                false,
            );

        let result = block_on(_scanner.run());

        println!("Result {:?}", result);

        let open_ports: Vec<u32> = result.iter().map(|x| x.port() as u32).collect();

        return Ok(Response::new(ScanPortResponse { open_ports }));
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let addr = "127.0.0.1:50051".parse()?;

    println!("addr: {:?}", addr);

    let service = tonic_reflection::server::Builder::configure()
        .register_encoded_file_descriptor_set(scanner::FILE_DESCRIPTOR_SET)
        .build()?;

    Server::builder()
        .add_service(service)
        .add_service(PortScannerServiceServer::new(
            ScannerService::new(std::time::Duration::from_millis(1000),
            )))
        .serve(addr)
        .await?;

    Ok(())
}