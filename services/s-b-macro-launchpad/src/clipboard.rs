#[cfg(windows)]
mod win {
    use std::sync::{Arc, Mutex};
    use windows_sys::Win32::Foundation::*;
    use windows_sys::Win32::System::LibraryLoader::GetModuleHandleW;
    use windows_sys::Win32::UI::WindowsAndMessaging::*;

    pub struct ClipboardManager {
        last_content: String,
        event_tx: Option<std::sync::mpsc::Sender<String>>,
    }

    impl ClipboardManager {
        pub fn new() -> Self {
            Self {
                last_content: String::new(),
                event_tx: None,
            }
        }

        pub fn has_changed(&self, new_content: &str) -> bool {
            self.last_content != new_content
        }

        pub fn record_change(&mut self, content_type: &str, preview: &str, _length: Option<usize>) {
            println!(
                "[s-b] Clipboard: type={}, preview={}",
                content_type,
                &preview[..preview.len().min(100)]
            );
        }

        pub fn set_last_content(&mut self, content: String) {
            self.last_content = content;
        }

        pub fn set_event_sender(&mut self, tx: std::sync::mpsc::Sender<String>) {
            self.event_tx = Some(tx);
        }

        pub fn send_event(&self, event: String) {
            if let Some(ref tx) = self.event_tx {
                let _ = tx.send(event);
            }
        }
    }

    static SHARED_MANAGER: Mutex<Option<Arc<Mutex<ClipboardManager>>>> = Mutex::new(None);

    pub fn set_shared_manager(mgr: Arc<Mutex<ClipboardManager>>) {
        *SHARED_MANAGER.lock().unwrap() = Some(mgr);
    }

    pub fn set_event_sender(tx: std::sync::mpsc::Sender<String>) {
        unsafe extern "system" fn clipboard_window_proc(
            hwnd: HWND,
            msg: u32,
            _wparam: WPARAM,
            _lparam: LPARAM,
        ) -> LRESULT {
            if msg == WM_CLIPBOARDUPDATE {
                if let Ok(text) = get_clipboard_text() {
                    if let Ok(guard) = SHARED_MANAGER.lock() {
                        if let Some(ref mgr_arc) = *guard {
                            if let Ok(mut mgr) = mgr_arc.lock() {
                                if mgr.has_changed(&text) {
                                    let preview: String = text.chars().take(200).collect();
                                    mgr.record_change("text", &preview, Some(text.len()));
                                    mgr.set_last_content(text.clone());
                                    let event =
                                        crate::bus::make_clipboard_changed(&preview, text.len());
                                    mgr.send_event(event);
                                }
                            }
                        }
                    }
                }
            }
            DefWindowProcW(hwnd, msg, _wparam, _lparam)
        }

        use windows_sys::Win32::System::DataExchange::*;

        let class_name: Vec<u16> = "NexusClipboard\0".encode_utf16().collect();
        unsafe {
            let wnd_class = WNDCLASSEXW {
                cbSize: std::mem::size_of::<WNDCLASSEXW>() as u32,
                style: 0,
                lpfnWndProc: Some(clipboard_window_proc),
                cbClsExtra: 0,
                cbWndExtra: 0,
                hInstance: GetModuleHandleW(std::ptr::null()),
                hIcon: std::ptr::null_mut(),
                hCursor: std::ptr::null_mut(),
                hbrBackground: std::ptr::null_mut(),
                lpszMenuName: std::ptr::null(),
                lpszClassName: class_name.as_ptr(),
                hIconSm: std::ptr::null_mut(),
            };

            RegisterClassExW(&wnd_class);

            let hwnd = CreateWindowExW(
                0,
                class_name.as_ptr(),
                std::ptr::null(),
                0,
                0,
                0,
                0,
                0,
                std::ptr::null_mut(),
                std::ptr::null_mut(),
                GetModuleHandleW(std::ptr::null()),
                std::ptr::null(),
            );

            if hwnd.is_null() {
                eprintln!("[s-b] Fehler: Konnte Clipboard-Window nicht erstellen");
                return;
            }

            AddClipboardFormatListener(hwnd);
        }

        std::thread::spawn(move || unsafe {
            let mut msg: MSG = std::mem::zeroed();
            let dummy_hwnd = std::ptr::null_mut() as HWND;
            while GetMessageW(&mut msg, dummy_hwnd, 0, 0) != 0 {
                TranslateMessage(&msg);
                DispatchMessageW(&msg);
            }
        });

        let _ = tx;
    }

    unsafe fn get_clipboard_text() -> Result<String, ()> {
        use windows_sys::Win32::System::DataExchange::*;
        use windows_sys::Win32::System::Memory::*;
        use windows_sys::Win32::System::Ole::CF_UNICODETEXT;

        if OpenClipboard(std::ptr::null_mut()) == 0 {
            return Err(());
        }

        let data = GetClipboardData(CF_UNICODETEXT as u32);
        if data.is_null() {
            CloseClipboard();
            return Err(());
        }

        let ptr = GlobalLock(data);
        if ptr.is_null() {
            CloseClipboard();
            return Err(());
        }

        let mut len = 0;
        let mut p = ptr as *const u16;
        while *p != 0 {
            len += 1;
            p = p.add(1);
        }

        let slice = std::slice::from_raw_parts(ptr as *const u16, len);
        let result = String::from_utf16_lossy(slice);

        GlobalUnlock(data);
        CloseClipboard();

        Ok(result)
    }

    pub fn set_clipboard_text(text: &str) -> Result<(), ()> {
        use windows_sys::Win32::System::DataExchange::*;
        use windows_sys::Win32::System::Memory::*;
        use windows_sys::Win32::System::Ole::CF_UNICODETEXT;

        unsafe {
            if OpenClipboard(std::ptr::null_mut()) == 0 {
                return Err(());
            }

            EmptyClipboard();

            let wide: Vec<u16> = text.encode_utf16().chain(std::iter::once(0)).collect();
            let size = wide.len() * 2;
            let h_mem = GlobalAlloc(GMEM_MOVEABLE, size);
            if h_mem.is_null() {
                CloseClipboard();
                return Err(());
            }

            let ptr = GlobalLock(h_mem);
            if ptr.is_null() {
                GlobalFree(h_mem);
                CloseClipboard();
                return Err(());
            }

            std::ptr::copy_nonoverlapping(wide.as_ptr(), ptr as *mut u16, wide.len());
            GlobalUnlock(h_mem);

            SetClipboardData(CF_UNICODETEXT as u32, h_mem);
            CloseClipboard();
        }

        Ok(())
    }
}

#[cfg(not(windows))]
mod win {
    use std::sync::{Arc, Mutex};

    pub struct ClipboardManager {
        last_content: String,
    }

    impl ClipboardManager {
        pub fn new() -> Self {
            Self {
                last_content: String::new(),
            }
        }

        pub fn has_changed(&self, new_content: &str) -> bool {
            self.last_content != new_content
        }

        pub fn record_change(&mut self, content_type: &str, preview: &str, _length: Option<usize>) {
            println!(
                "[s-b] Clipboard (Stub): type={}, preview={}",
                content_type, preview
            );
        }

        pub fn set_last_content(&mut self, content: String) {
            self.last_content = content;
        }

        pub fn set_event_sender(&mut self, _tx: std::sync::mpsc::Sender<String>) {}

        pub fn send_event(&self, _event: String) {}
    }

    pub fn set_shared_manager(_mgr: Arc<Mutex<ClipboardManager>>) {}
    pub fn set_event_sender(_tx: std::sync::mpsc::Sender<String>) {}
    pub fn spawn_clipboard_monitor(
        _mgr: Arc<Mutex<ClipboardManager>>,
        _event_tx: std::sync::mpsc::Sender<String>,
    ) {
    }
    pub fn set_clipboard_text(_text: &str) -> Result<(), ()> {
        Ok(())
    }
}

use std::sync::{Arc, Mutex};

pub struct ClipboardWatcher {
    manager: Arc<Mutex<win::ClipboardManager>>,
}

impl ClipboardWatcher {
    pub fn new() -> Self {
        Self {
            manager: Arc::new(Mutex::new(win::ClipboardManager::new())),
        }
    }

    pub fn start(&self, event_tx: std::sync::mpsc::Sender<String>) {
        let shared = Arc::clone(&self.manager);
        win::set_shared_manager(shared);
        win::set_event_sender(event_tx);
    }

    pub fn set_text(&self, text: &str) -> Result<(), ()> {
        win::set_clipboard_text(text)
    }
}
