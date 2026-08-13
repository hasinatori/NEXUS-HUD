use std::env;

use chrono::{SecondsFormat, Utc};
use futures_util::{SinkExt, StreamExt};
use tokio::time::{interval, Duration};
use tokio_tungstenite::connect_async;

const SOURCE: &str = "S-B";
const SERVICE_ID: &str = "s-b-macro-launchpad";
const HELLO_INTERVAL: Duration = Duration::from_secs(5);

fn hello() -> String {
    format!(
        "{{\"jsonrpc\":\"2.0\",\"method\":\"event.system.hello\",\"params\":{{\
         \"source\":\"{SOURCE}\",\"protocol_version\":1,\"service_id\":\"{SERVICE_ID}\",\
         \"version\":\"{}\",\"ts\":\"{}\"}}}}",
        env!("CARGO_PKG_VERSION"),
        Utc::now().to_rfc3339_opts(SecondsFormat::Secs, true)
    )
}

#[tokio::main]
async fn main() {
    let port = env::var("NEXUS_WS_PORT")
        .ok()
        .and_then(|p| p.parse::<u16>().ok())
        .unwrap_or(49152);
    let url = format!("ws://127.0.0.1:{port}/");

    loop {
        match run(&url).await {
            Ok(()) => break,
            Err(err) => {
                eprintln!("[{SERVICE_ID}] Verbindungsfehler: {err}; neuer Versuch in 2 s");
                tokio::time::sleep(Duration::from_secs(2)).await;
            }
        }
    }
}

async fn run(url: &str) -> Result<(), Box<dyn std::error::Error>> {
    let (mut ws, _) = connect_async(url).await?;
    println!("[{SERVICE_ID}] verbunden mit {url}");

    let mut ticker = interval(HELLO_INTERVAL);
    loop {
        tokio::select! {
            _ = ticker.tick() => {
                ws.send(hello().into()).await?;
            }
            msg = ws.next() => match msg {
                Some(Ok(_)) => {}
                Some(Err(err)) => return Err(err.into()),
                None => return Ok(()),
            },
        }
    }
}
