mod bus;
mod clipboard;
mod hotkey;
mod process;
mod window;

use std::env;
use std::sync::{Arc, mpsc};
use std::sync::Mutex;

use futures_util::{SinkExt, StreamExt};
use tokio::time::{interval, Duration};
use tokio_tungstenite::connect_async;

use bus::{AppLaunchCmd, HotkeyRegisterCmd, JsonRpcMessage, WindowMoveCmd};

const HELLO_INTERVAL: Duration = Duration::from_secs(5);

#[tokio::main]
async fn main() {
    let port = env::var("NEXUS_WS_PORT")
        .ok()
        .and_then(|p| p.parse::<u16>().ok())
        .unwrap_or(49152);
    let url = format!("ws://127.0.0.1:{port}/");

    let mut hotkey_mgr = hotkey::HotkeyManager::new();
    let (native_tx, native_rx) = mpsc::channel::<u32>();

    let _hotkey_thread = hotkey::spawn_message_loop(native_tx);

    let (hotkey_async_tx, mut hotkey_async_rx) = tokio::sync::mpsc::unbounded_channel::<u32>();

    std::thread::spawn(move || {
        while let Ok(id) = native_rx.recv() {
            let _ = hotkey_async_tx.send(id);
        }
    });

    let clipboard_mgr = Arc::new(Mutex::new(clipboard::ClipboardManager::new()));
    let (clipboard_event_tx, mut clipboard_event_rx) = tokio::sync::mpsc::unbounded_channel::<String>();

    let _clipboard_thread = clipboard::spawn_monitor(clipboard_mgr.clone(), clipboard_event_tx);

    loop {
        match run(
            &url,
            &mut hotkey_mgr,
            &mut hotkey_async_rx,
            &mut clipboard_event_rx,
        )
        .await
        {
            Ok(()) => break,
            Err(err) => {
                eprintln!("[s-b-macro-launchpad] Verbindungsfehler: {err}; neuer Versuch in 2 s");
                tokio::time::sleep(Duration::from_secs(2)).await;
            }
        }
    }

    hotkey_mgr.unregister_all();
}

async fn run(
    url: &str,
    hotkey_mgr: &mut hotkey::HotkeyManager,
    hotkey_async_rx: &mut tokio::sync::mpsc::UnboundedReceiver<u32>,
    clipboard_event_rx: &mut tokio::sync::mpsc::UnboundedReceiver<String>,
) -> Result<(), Box<dyn std::error::Error>> {
    let (mut ws, _) = connect_async(url).await?;
    println!("[s-b-macro-launchpad] verbunden mit {url}");

    let mut ticker = interval(HELLO_INTERVAL);
    loop {
        tokio::select! {
            _ = ticker.tick() => {
                ws.send(bus::make_hello(env!("CARGO_PKG_VERSION")).into()).await?;
            }

            Some(hotkey_id) = hotkey_async_rx.recv() => {
                if let Some(name) = hotkey_mgr.find_by_id(hotkey_id) {
                    let name_clone = name.clone();
                    ws.send(bus::make_hotkey_triggered(&name_clone).into()).await?;
                    println!("[s-b] Hotkey ausgelöst: {name_clone}");
                }
            }

            Some(clipboard_event) = clipboard_event_rx.recv() => {
                ws.send(clipboard_event.into()).await?;
            }

            msg = ws.next() => match msg {
                Some(Ok(text)) => {
                    if let Ok(raw) = text.to_text() {
                        handle_message(raw, hotkey_mgr, &mut ws).await?;
                    }
                }
                Some(Err(err)) => return Err(err.into()),
                None => return Ok(()),
            },
        }
    }
}

async fn handle_message(
    raw: &str,
    hotkey_mgr: &mut hotkey::HotkeyManager,
    ws: &mut (impl SinkExt<String, Error = Box<dyn std::error::Error + Send + Sync>> + Unpin),
) -> Result<(), Box<dyn std::error::Error>> {
    let msg: JsonRpcMessage = match serde_json::from_str(raw) {
        Ok(m) => m,
        Err(_) => return Ok(()),
    };

    match msg.method.as_str() {
        "cmd.hotkey.register" => {
            if let Ok(cmd) = serde_json::from_value::<HotkeyRegisterCmd>(msg.params) {
                match hotkey_mgr.register(&cmd) {
                    Ok(()) => {
                        println!("[s-b] Hotkey '{}' registriert", cmd.hotkey_id);
                    }
                    Err(e) => {
                        eprintln!("[s-b] Hotkey-Registrierung fehlgeschlagen: {e}");
                    }
                }
            }
        }
        "cmd.app.launch" => {
            if let Ok(cmd) = serde_json::from_value::<AppLaunchCmd>(msg.params) {
                match process::launch(&cmd) {
                    Ok(pid) => {
                        ws.send(bus::make_process_started(pid, &cmd.path).into()).await?;
                    }
                    Err(e) => {
                        eprintln!("[s-b] Prozessstart fehlgeschlagen: {e}");
                    }
                }
            }
        }
        "cmd.window.move" => {
            if let Ok(cmd) = serde_json::from_value::<WindowMoveCmd>(msg.params) {
                match window::move_window(&cmd) {
                    Ok(()) => {
                        ws.send(
                            bus::make_window_moved(
                                &cmd.window_title,
                                cmd.x,
                                cmd.y,
                                cmd.width,
                                cmd.height,
                            )
                            .into(),
                        )
                        .await?;
                    }
                    Err(e) => {
                        eprintln!("[s-b] Fenster-Verschiebung fehlgeschlagen: {e}");
                    }
                }
            }
        }
        "cmd.clipboard.set" => {
            if let Some(content) = msg.params.get("content").and_then(|c| c.as_str()) {
                if clipboard::set_text(content) {
                    println!("[s-b] Clipboard-Inhalt gesetzt ({} Zeichen)", content.len());
                } else {
                    eprintln!("[s-b] Clipboard setzen fehlgeschlagen");
                }
            }
        }
        "cmd.clipboard.get_history" => {
            println!("[s-b] Clipboard-History angefordert");
        }
        "event.system.hello" => {
            if let Some(source) = msg.params.get("source").and_then(|s| s.as_str()) {
                if source != bus::SOURCE {
                    println!("[s-b] Anderer Service verbunden: {source}");
                }
            }
        }
        _ => {}
    }

    Ok(())
}
