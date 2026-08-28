mod bus;
mod clipboard;
mod hotkey;
mod process;
mod window;

use futures_util::{SinkExt, StreamExt};
use tokio_tungstenite::tungstenite::Message;

async fn handle_message(
    raw: String,
    hotkey_mgr: &mut hotkey::HotkeyManager,
    ws: &mut (impl SinkExt<Message, Error = tokio_tungstenite::tungstenite::Error> + Unpin),
) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let v: serde_json::Value = serde_json::from_str(&raw)?;
    let method = v.get("method").and_then(|m| m.as_str()).unwrap_or("");

    match method {
        "cmd.hotkey.register" => {
            let cmd: bus::HotkeyRegisterCmd =
                serde_json::from_value(v.get("params").cloned().unwrap_or_default())?;
            match hotkey_mgr.register(&cmd) {
                Ok(()) => println!("[s-b] Hotkey registriert: {}", cmd.hotkey_id),
                Err(e) => eprintln!("[s-b] Hotkey-Fehler: {}", e),
            }
        }
        "cmd.app.launch" => {
            let cmd: bus::AppLaunchCmd =
                serde_json::from_value(v.get("params").cloned().unwrap_or_default())?;
            match process::launch_app(&cmd.path, &cmd.args, cmd.focus) {
                Ok(pid) => {
                    let event = bus::make_process_started(pid, &cmd.path);
                    ws.send(Message::Text(event.into())).await?;
                }
                Err(e) => eprintln!("[s-b] App-Start fehlgeschlagen: {}", e),
            }
        }
        "cmd.window.move" => {
            let cmd: bus::WindowMoveCmd =
                serde_json::from_value(v.get("params").cloned().unwrap_or_default())?;
            if let Some(hwnd) = window::find_window(&cmd.window_title) {
                window::set_window_pos(hwnd, cmd.x, cmd.y, cmd.width, cmd.height);
                let event =
                    bus::make_window_moved(&cmd.window_title, cmd.x, cmd.y, cmd.width, cmd.height);
                ws.send(Message::Text(event.into())).await?;
            }
        }
        "cmd.window.focus" => {
            let title = v
                .get("params")
                .and_then(|p| p.get("window_title"))
                .and_then(|t| t.as_str())
                .unwrap_or("");
            if let Some(hwnd) = window::find_window(title) {
                window::focus_window(hwnd);
            }
        }
        "cmd.profile.switch" => {
            let profile = v
                .get("params")
                .and_then(|p| p.get("profile"))
                .and_then(|t| t.as_str())
                .unwrap_or("");
            println!("[s-b] Profil gewechselt: {} (alle Hotkeys zurueckgesetzt)", profile);
            hotkey_mgr.clear_all();
        }
        "cmd.clipboard.set" => {
            let text = v
                .get("params")
                .and_then(|p| p.get("text"))
                .and_then(|t| t.as_str())
                .unwrap_or("");
            let _ = clipboard::ClipboardWatcher::new().set_text(text);
        }
        _ => {
            eprintln!("[s-b] Unbekannte Methode: {}", method);
        }
    }

    Ok(())
}

#[tokio::main]
async fn main() {
    let server_url = std::env::var("SERVER_URL").unwrap_or_else(|_| "ws://localhost:8080/ws".into());
    let version = env!("CARGO_PKG_VERSION");

    println!("[s-b] Starte s-b-macro-launchpad v{}", version);
    println!("[s-b] Verbinde zu {}...", server_url);

    let (ws_stream, _) = tokio_tungstenite::connect_async(&server_url)
        .await
        .expect("[s-b] Verbindung fehlgeschlagen");
    println!("[s-b] Verbunden!");

    let (mut ws_write, mut read) = ws_stream.split();

    let hello = bus::make_hello(version);
    ws_write.send(Message::Text(hello.into())).await.unwrap();

    let mut hotkey_mgr = hotkey::HotkeyManager::new();

    let (std_event_tx, std_event_rx) = std::sync::mpsc::channel::<String>();
    let clipboard_watcher = clipboard::ClipboardWatcher::new();
    clipboard_watcher.start(std_event_tx);

    let (std_hotkey_tx, std_hotkey_rx) = std::sync::mpsc::channel::<u32>();
    let _msg_loop_handle = hotkey::spawn_message_loop(std_hotkey_tx);

    let (tokio_event_tx, mut tokio_event_rx) = tokio::sync::mpsc::channel::<String>(32);
    let (tokio_hotkey_tx, mut tokio_hotkey_rx) = tokio::sync::mpsc::channel::<u32>(32);

    std::thread::spawn(move || {
        while let Ok(event) = std_event_rx.recv() {
            if tokio_event_tx.blocking_send(event).is_err() {
                break;
            }
        }
    });

    std::thread::spawn(move || {
        while let Ok(id) = std_hotkey_rx.recv() {
            if tokio_hotkey_tx.blocking_send(id).is_err() {
                break;
            }
        }
    });

    loop {
        tokio::select! {
            msg = read.next() => {
                match msg {
                    Some(Ok(Message::Text(text))) => {
                        match handle_message(text.to_string(), &mut hotkey_mgr, &mut ws_write).await {
                            Ok(()) => {}
                            Err(e) => eprintln!("[s-b] Fehler: {}", e),
                        }
                    }
                    Some(Ok(Message::Close(_))) => {
                        println!("[s-b] Verbindung geschlossen");
                        break;
                    }
                    Some(Err(e)) => {
                        eprintln!("[s-b] WebSocket-Fehler: {}", e);
                        break;
                    }
                    None => {
                        println!("[s-b] Stream beendet");
                        break;
                    }
                    _ => {}
                }
            }
            Some(event) = tokio_event_rx.recv() => {
                if ws_write.send(Message::Text(event.into())).await.is_err() {
                    eprintln!("[s-b] WebSocket Send-Fehler");
                    break;
                }
            }
            Some(hotkey_id) = tokio_hotkey_rx.recv() => {
                if let Some(hotkey_id_str) = hotkey_mgr.find_by_id(hotkey_id) {
                    let id = hotkey_id_str.clone();
                    let event = bus::make_hotkey_triggered(&id);
                    if ws_write.send(Message::Text(event.into())).await.is_err() {
                        eprintln!("[s-b] WebSocket Send-Fehler");
                    }
                }
            }
        }
    }
}
