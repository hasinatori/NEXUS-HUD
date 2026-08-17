use chrono::{SecondsFormat, Utc};
use serde::{Deserialize, Serialize};

pub const SOURCE: &str = "S-B";
pub const SERVICE_ID: &str = "s-b-macro-launchpad";
pub const PROTOCOL_VERSION: u32 = 1;

#[derive(Debug, Serialize, Deserialize)]
pub struct JsonRpcMessage {
    pub jsonrpc: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub id: Option<String>,
    pub method: String,
    pub params: serde_json::Value,
}

#[derive(Debug, Deserialize)]
pub struct HotkeyRegisterCmd {
    pub source: String,
    pub protocol_version: u32,
    pub hotkey_id: String,
    pub modifiers: Vec<String>,
    pub key: String,
}

#[derive(Debug, Deserialize)]
pub struct AppLaunchCmd {
    pub source: String,
    pub protocol_version: u32,
    pub path: String,
    #[serde(default)]
    pub args: Vec<String>,
    #[serde(default)]
    pub focus: bool,
}

#[derive(Debug, Deserialize)]
pub struct WindowMoveCmd {
    pub source: String,
    pub protocol_version: u32,
    pub window_title: String,
    pub x: i32,
    pub y: i32,
    pub width: u32,
    pub height: u32,
}

pub fn make_hello(version: &str) -> String {
    let msg = JsonRpcMessage {
        jsonrpc: "2.0".into(),
        id: None,
        method: "event.system.hello".into(),
        params: serde_json::json!({
            "source": SOURCE,
            "protocol_version": PROTOCOL_VERSION,
            "service_id": SERVICE_ID,
            "version": version,
            "ts": Utc::now().to_rfc3339_opts(SecondsFormat::Secs, true),
        }),
    };
    serde_json::to_string(&msg).unwrap()
}

pub fn make_hotkey_triggered(hotkey_id: &str) -> String {
    let msg = JsonRpcMessage {
        jsonrpc: "2.0".into(),
        id: None,
        method: "event.hotkey.triggered".into(),
        params: serde_json::json!({
            "source": SOURCE,
            "protocol_version": PROTOCOL_VERSION,
            "hotkey_id": hotkey_id,
            "ts": Utc::now().to_rfc3339_opts(SecondsFormat::Secs, true),
        }),
    };
    serde_json::to_string(&msg).unwrap()
}

pub fn make_process_started(pid: u32, name: &str) -> String {
    let msg = JsonRpcMessage {
        jsonrpc: "2.0".into(),
        id: None,
        method: "event.process.started".into(),
        params: serde_json::json!({
            "source": SOURCE,
            "protocol_version": PROTOCOL_VERSION,
            "pid": pid,
            "name": name,
            "ts": Utc::now().to_rfc3339_opts(SecondsFormat::Secs, true),
        }),
    };
    serde_json::to_string(&msg).unwrap()
}

pub fn make_window_moved(window_title: &str, x: i32, y: i32, width: u32, height: u32) -> String {
    let msg = JsonRpcMessage {
        jsonrpc: "2.0".into(),
        id: None,
        method: "event.window.moved".into(),
        params: serde_json::json!({
            "source": SOURCE,
            "protocol_version": PROTOCOL_VERSION,
            "window_title": window_title,
            "x": x,
            "y": y,
            "width": width,
            "height": height,
            "ts": Utc::now().to_rfc3339_opts(SecondsFormat::Secs, true),
        }),
    };
    serde_json::to_string(&msg).unwrap()
}

pub fn make_clipboard_changed(preview: &str, length: usize) -> String {
    let msg = JsonRpcMessage {
        jsonrpc: "2.0".into(),
        id: None,
        method: "event.clipboard.changed".into(),
        params: serde_json::json!({
            "source": SOURCE,
            "protocol_version": PROTOCOL_VERSION,
            "content_type": "text",
            "preview": preview,
            "length": length,
            "ts": Utc::now().to_rfc3339_opts(SecondsFormat::Secs, true),
        }),
    };
    serde_json::to_string(&msg).unwrap()
}
