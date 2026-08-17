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

pub fn make_call_initiated(call_id: &str, to: &str) -> String {
    let msg = JsonRpcMessage {
        jsonrpc: "2.0".into(),
        id: None,
        method: "event.call.initiated".into(),
        params: serde_json::json!({
            "source": SOURCE,
            "protocol_version": PROTOCOL_VERSION,
            "call_id": call_id,
            "to": to,
            "ts": Utc::now().to_rfc3339_opts(SecondsFormat::Secs, true),
        }),
    };
    serde_json::to_string(&msg).unwrap()
}

pub fn make_call_ended(call_id: &str, duration_sec: u32) -> String {
    let msg = JsonRpcMessage {
        jsonrpc: "2.0".into(),
        id: None,
        method: "event.call.ended".into(),
        params: serde_json::json!({
            "source": SOURCE,
            "protocol_version": PROTOCOL_VERSION,
            "call_id": call_id,
            "duration_sec": duration_sec,
            "ts": Utc::now().to_rfc3339_opts(SecondsFormat::Secs, true),
        }),
    };
    serde_json::to_string(&msg).unwrap()
}

pub fn make_profile_switched(profile: &str) -> String {
    let msg = JsonRpcMessage {
        jsonrpc: "2.0".into(),
        id: None,
        method: "event.profile.switched".into(),
        params: serde_json::json!({
            "source": SOURCE,
            "protocol_version": PROTOCOL_VERSION,
            "profile": profile,
            "ts": Utc::now().to_rfc3339_opts(SecondsFormat::Secs, true),
        }),
    };
    serde_json::to_string(&msg).unwrap()
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::Value;

    #[test]
    fn test_make_hello_structure() {
        let msg = make_hello("1.0.0");
        let v: Value = serde_json::from_str(&msg).unwrap();

        assert_eq!(v["jsonrpc"], "2.0");
        assert!(v["id"].is_null());
        assert_eq!(v["method"], "event.system.hello");
        assert_eq!(v["params"]["source"], SOURCE);
        assert_eq!(v["params"]["protocol_version"], PROTOCOL_VERSION);
        assert_eq!(v["params"]["service_id"], SERVICE_ID);
        assert_eq!(v["params"]["version"], "1.0.0");
        assert!(v["params"]["ts"].is_string());
    }

    #[test]
    fn test_make_hotkey_triggered() {
        let msg = make_hotkey_triggered("hk-1");
        let v: Value = serde_json::from_str(&msg).unwrap();

        assert_eq!(v["method"], "event.hotkey.triggered");
        assert_eq!(v["params"]["source"], SOURCE);
        assert_eq!(v["params"]["hotkey_id"], "hk-1");
        assert!(v["params"]["ts"].is_string());
    }

    #[test]
    fn test_make_process_started() {
        let msg = make_process_started(42, "/usr/bin/test");
        let v: Value = serde_json::from_str(&msg).unwrap();

        assert_eq!(v["method"], "event.process.started");
        assert_eq!(v["params"]["pid"], 42);
        assert_eq!(v["params"]["name"], "/usr/bin/test");
        assert_eq!(v["params"]["source"], SOURCE);
    }

    #[test]
    fn test_make_window_moved() {
        let msg = make_window_moved("MyWindow", 100, 200, 800, 600);
        let v: Value = serde_json::from_str(&msg).unwrap();

        assert_eq!(v["method"], "event.window.moved");
        assert_eq!(v["params"]["window_title"], "MyWindow");
        assert_eq!(v["params"]["x"], 100);
        assert_eq!(v["params"]["y"], 200);
        assert_eq!(v["params"]["width"], 800);
        assert_eq!(v["params"]["height"], 600);
    }

    #[test]
    fn test_make_clipboard_changed() {
        let msg = make_clipboard_changed("hello world", 11);
        let v: Value = serde_json::from_str(&msg).unwrap();

        assert_eq!(v["method"], "event.clipboard.changed");
        assert_eq!(v["params"]["preview"], "hello world");
        assert_eq!(v["params"]["length"], 11);
        assert_eq!(v["params"]["content_type"], "text");
    }

    #[test]
    fn test_make_call_initiated() {
        let msg = make_call_initiated("call-abc", "+1234567890");
        let v: Value = serde_json::from_str(&msg).unwrap();

        assert_eq!(v["method"], "event.call.initiated");
        assert_eq!(v["params"]["call_id"], "call-abc");
        assert_eq!(v["params"]["to"], "+1234567890");
        assert_eq!(v["params"]["source"], SOURCE);
    }

    #[test]
    fn test_make_call_ended() {
        let msg = make_call_ended("call-abc", 120);
        let v: Value = serde_json::from_str(&msg).unwrap();

        assert_eq!(v["method"], "event.call.ended");
        assert_eq!(v["params"]["call_id"], "call-abc");
        assert_eq!(v["params"]["duration_sec"], 120);
    }

    #[test]
    fn test_make_profile_switched() {
        let msg = make_profile_switched("gaming");
        let v: Value = serde_json::from_str(&msg).unwrap();

        assert_eq!(v["method"], "event.profile.switched");
        assert_eq!(v["params"]["profile"], "gaming");
    }

    #[test]
    fn test_all_messages_are_valid_jsonrpc() {
        let messages = vec![
            make_hello("0.1.0"),
            make_hotkey_triggered("hk-1"),
            make_process_started(1, "test"),
            make_window_moved("w", 0, 0, 100, 100),
            make_clipboard_changed("x", 1),
            make_call_initiated("c1", "+123"),
            make_call_ended("c1", 0),
            make_profile_switched("dev"),
        ];

        for raw in messages {
            let v: Value = serde_json::from_str(&raw).unwrap();
            assert_eq!(v["jsonrpc"], "2.0");
            assert!(v["method"].is_string());
            assert!(v["params"].is_object());
            assert!(v["params"]["source"].is_string());
            assert!(v["params"]["protocol_version"].is_number());
            assert!(v["params"]["ts"].is_string());
        }
    }

    #[test]
    fn test_jsonrpc_message_serialization_roundtrip() {
        let msg = JsonRpcMessage {
            jsonrpc: "2.0".into(),
            id: Some("test-123".into()),
            method: "test.method".into(),
            params: serde_json::json!({"key": "value"}),
        };
        let serialized = serde_json::to_string(&msg).unwrap();
        let deserialized: JsonRpcMessage = serde_json::from_str(&serialized).unwrap();

        assert_eq!(deserialized.jsonrpc, "2.0");
        assert_eq!(deserialized.id, Some("test-123".into()));
        assert_eq!(deserialized.method, "test.method");
        assert_eq!(deserialized.params["key"], "value");
    }

    #[test]
    fn test_jsonrpc_message_optional_id_omitted_when_none() {
        let msg = JsonRpcMessage {
            jsonrpc: "2.0".into(),
            id: None,
            method: "test.notification".into(),
            params: serde_json::json!({}),
        };
        let serialized = serde_json::to_string(&msg).unwrap();
        assert!(!serialized.contains("\"id\""));
    }
}
