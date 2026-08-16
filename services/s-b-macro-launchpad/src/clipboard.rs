use std::collections::VecDeque;
use std::sync::{Arc, Mutex};

use chrono::{SecondsFormat, Utc};
use serde::{Deserialize, Serialize};

use crate::bus::{JsonRpcMessage, SOURCE, PROTOCOL_VERSION};

const MAX_HISTORY: usize = 50;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ClipboardEntry {
    pub content_type: String,
    pub preview: String,
    pub length: Option<usize>,
    pub ts: String,
}

pub struct ClipboardManager {
    history: VecDeque<ClipboardEntry>,
    last_content: String,
}

impl ClipboardManager {
    pub fn new() -> Self {
        Self {
            history: VecDeque::with_capacity(MAX_HISTORY),
            last_content: String::new(),
        }
    }

    pub fn record_change(&mut self, content_type: &str, preview: &str, length: Option<usize>) {
        let entry = ClipboardEntry {
            content_type: content_type.to_string(),
            preview: preview.to_string(),
            length,
            ts: Utc::now().to_rfc3339_opts(SecondsFormat::Secs, true),
        };

        if self.history.len() >= MAX_HISTORY {
            self.history.pop_front();
        }
        self.history.push_back(entry);
    }

    pub fn set_last_content(&mut self, content: String) {
        self.last_content = content;
    }

    pub fn get_last_content(&self) -> &str {
        &self.last_content
    }

    pub fn get_history(&self) -> Vec<&ClipboardEntry> {
        self.history.iter().collect()
    }

    pub fn has_changed(&self, new_content: &str) -> bool {
        self.last_content != new_content
    }
}

pub fn make_clipboard_changed_event(preview: &str, length: usize) -> String {
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

#[cfg(windows)]
mod win {
    use std::sync::{Arc, Mutex};
    use windows_sys::Win32::UI::WindowsAndMessaging::*;
    use windows_sys::Win32::Foundation::*;
    use windows_sys::Win32::System::Memory::*;
    use super::ClipboardManager;

    pub const WM_CLIPBOARDUPDATE: u32 = 0x031D;

    pub unsafe fn read_clipboard_text() -> Option<String> {
        if OpenClipboard(0) == 0 {
            return None;
        }

        let data = GetClipboardData(CF_UNICODETEXT.0);
        if data.is_null() {
            CloseClipboard();
            return None;
        }

        let ptr = GlobalLock(data) as *const u16;
        if ptr.is_null() {
            CloseClipboard();
            return None;
        }

        let mut len = 0;
        while *ptr.add(len) != 0 {
            len += 1;
        }

        let slice = std::slice::from_raw_parts(ptr, len);
        let text = String::from_utf16_lossy(slice);

        GlobalUnlock(data);
        CloseClipboard();

        Some(text)
    }

    pub unsafe fn set_clipboard_text(text: &str) -> bool {
        if OpenClipboard(0) == 0 {
            return false;
        }

        EmptyClipboard();

        let wide: Vec<u16> = text.encode_utf16().chain(std::iter::once(0)).collect();
        let size = wide.len() * std::mem::size_of::<u16>();
        let h_mem = GlobalAlloc(GMEM_MOVEABLE, size);

        if h_mem.is_null() {
            CloseClipboard();
            return false;
        }

        let ptr = GlobalLock(h_mem) as *mut u16;
        if ptr.is_null() {
            GlobalFree(h_mem);
            CloseClipboard();
            return false;
        }

        std::ptr::copy_nonoverlapping(wide.as_ptr(), ptr, wide.len());
        GlobalUnlock(h_mem);

        SetClipboardData(CF_UNICODETEXT.0, h_mem);
        CloseClipboard();

        true
    }

    pub fn spawn_clipboard_monitor(
        shared_mgr: Arc<Mutex<ClipboardManager>>,
        event_tx: tokio::sync::mpsc::UnboundedSender<String>,
    ) -> std::thread::JoinHandle<()> {
        std::thread::spawn(move || {
            unsafe {
                let class_name: Vec<u16> = "NexusClipboardMsg\0".encode_utf16().collect();
                let wnd_class = WNDCLASSEXW {
                    cbSize: std::mem::size_of::<WNDCLASSEXW>() as u32,
                    style: 0,
                    lpfnWndProc: Some(clipboard_proc),
                    cbClsExtra: 0,
                    cbWndExtra: 0,
                    hInstance: windows_sys::Win32::System::LibraryLoader::GetModuleHandleW(std::ptr::null()),
                    hIcon: 0,
                    hCursor: 0,
                    hbrBackground: 0,
                    lpszMenuName: std::ptr::null(),
                    lpszClassName: class_name.as_ptr(),
                    hIconSm: 0,
                };

                RegisterClassExW(&wnd_class);

                let hwnd = CreateWindowExW(
                    0,
                    class_name.as_ptr(),
                    std::ptr::null(),
                    0,
                    0, 0, 0, 0,
                    HWND_MESSAGE,
                    std::ptr::null_mut(),
                    windows_sys::Win32::System::LibraryLoader::GetModuleHandleW(std::ptr::null()),
                    std::ptr::null(),
                );

                if hwnd.is_null() {
                    eprintln!("[s-b] Clipboard: Konnte Message-Window nicht erstellen");
                    return;
                }

                AddClipboardFormatListener(hwnd);

                let mut msg: MSG = std::mem::zeroed();
                while GetMessageW(&mut msg, hwnd, 0, 0) > 0 {
                    TranslateMessage(&msg);
                    DispatchMessageW(&msg);
                }

                RemoveClipboardFormatListener(hwnd);
            }
        })
    }

    unsafe extern "system" fn clipboard_proc(
        hwnd: HWND,
        msg: u32,
        _wparam: WPARAM,
        _lparam: LPARAM,
    ) -> LRESULT {
        if msg == WM_CLIPBOARDUPDATE {
            if let Some(text) = read_clipboard_text() {
                if text.len() > 0 {
                    let mut mgr = SHARED_MANAGER.lock().unwrap();
                    if mgr.has_changed(&text) {
                        let preview = if text.len() > 200 {
                            format!("{}...", &text[..200])
                        } else {
                            text.clone()
                        };
                        mgr.record_change("text", &preview, Some(text.len()));
                        mgr.set_last_content(text);

                        if let Some(tx) = EVENT_SENDER.lock().unwrap().as_ref() {
                            let event = super::make_clipboard_changed_event(&preview, preview.len());
                            let _ = tx.send(event);
                        }
                    }
                }
            }
            return 0;
        }
        DefWindowProcW(hwnd, msg, _wparam, _lparam)
    }

    use std::sync::Mutex;

    static SHARED_MANAGER: Mutex<Option<Arc<Mutex<ClipboardManager>>>> = Mutex::new(None);
    static EVENT_SENDER: Mutex<Option<tokio::sync::mpsc::UnboundedSender<String>>> = Mutex::new(None);

    pub fn set_shared_manager(mgr: Arc<Mutex<ClipboardManager>>) {
        *SHARED_MANAGER.lock().unwrap() = Some(mgr);
    }

    pub fn set_event_sender(tx: tokio::sync::mpsc::UnboundedSender<String>) {
        *EVENT_SENDER.lock().unwrap() = Some(tx);
    }
}

#[cfg(not(windows))]
mod win {
    use std::sync::{Arc, Mutex};
    use super::ClipboardManager;

    pub unsafe fn read_clipboard_text() -> Option<String> {
        None
    }

    pub unsafe fn set_clipboard_text(_text: &str) -> bool {
        false
    }

    pub fn spawn_clipboard_monitor(
        _shared_mgr: Arc<Mutex<ClipboardManager>>,
        _event_tx: tokio::sync::mpsc::UnboundedSender<String>,
    ) -> std::thread::JoinHandle<()> {
        std::thread::spawn(move || {
            loop {
                std::thread::sleep(std::time::Duration::from_secs(3600));
            }
        })
    }

    pub fn set_shared_manager(_mgr: Arc<Mutex<ClipboardManager>>) {}
    pub fn set_event_sender(_tx: tokio::sync::mpsc::UnboundedSender<String>) {}
}

pub fn read_text() -> Option<String> {
    unsafe { win::read_clipboard_text() }
}

pub fn set_text(text: &str) -> bool {
    unsafe { win::set_clipboard_text(text) }
}

pub fn spawn_monitor(
    shared_mgr: Arc<Mutex<ClipboardManager>>,
    event_tx: tokio::sync::mpsc::UnboundedSender<String>,
) -> std::thread::JoinHandle<()> {
    win::set_shared_manager(shared_mgr);
    win::set_event_sender(event_tx.clone());
    win::spawn_clipboard_monitor(shared_mgr, event_tx)
}
