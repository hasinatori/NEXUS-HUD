use std::collections::HashMap;
use std::sync::mpsc;

use crate::bus::HotkeyRegisterCmd;

#[cfg(windows)]
mod win {
    use windows_sys::Win32::UI::WindowsAndMessaging::*;
    use windows_sys::Win32::UI::Input::KeyboardAndMouse::*;
    use windows_sys::Win32::Foundation::*;
    use std::ffi::c_void;

    pub const WM_HOTKEY: u32 = 0x0312;

    pub fn parse_modifiers(mods: &[String]) -> u32 {
        let mut result = 0u32;
        for m in mods {
            match m.to_uppercase().as_str() {
                "ALT" | "MENU" => result |= MOD_ALT as u32,
                "CTRL" | "CONTROL" => result |= MOD_CONTROL as u32,
                "SHIFT" => result |= MOD_SHIFT as u32,
                "WIN" | "META" => result |= MOD_WIN as u32,
                _ => {}
            }
        }
        result
    }

    pub fn parse_vkey(key: &str) -> Option<u32> {
        match key.to_uppercase().as_str() {
            "F1" => Some(VK_F1 as u32),
            "F2" => Some(VK_F2 as u32),
            "F3" => Some(VK_F3 as u32),
            "F4" => Some(VK_F4 as u32),
            "F5" => Some(VK_F5 as u32),
            "F6" => Some(VK_F6 as u32),
            "F7" => Some(VK_F7 as u32),
            "F8" => Some(VK_F8 as u32),
            "F9" => Some(VK_F9 as u32),
            "F10" => Some(VK_F10 as u32),
            "F11" => Some(VK_F11 as u32),
            "F12" => Some(VK_F12 as u32),
            "A" => Some(VK_A as u32),
            "B" => Some(VK_B as u32),
            "C" => Some(VK_C as u32),
            "D" => Some(VK_D as u32),
            "E" => Some(VK_E as u32),
            "F" => Some(VK_F as u32),
            "G" => Some(VK_G as u32),
            "H" => Some(VK_H as u32),
            "I" => Some(VK_I as u32),
            "J" => Some(VK_J as u32),
            "K" => Some(VK_K as u32),
            "L" => Some(VK_L as u32),
            "M" => Some(VK_M as u32),
            "N" => Some(VK_N as u32),
            "O" => Some(VK_O as u32),
            "P" => Some(VK_P as u32),
            "Q" => Some(VK_Q as u32),
            "R" => Some(VK_R as u32),
            "S" => Some(VK_S as u32),
            "T" => Some(VK_T as u32),
            "U" => Some(VK_U as u32),
            "V" => Some(VK_V as u32),
            "W" => Some(VK_W as u32),
            "X" => Some(VK_X as u32),
            "Y" => Some(VK_Y as u32),
            "Z" => Some(VK_Z as u32),
            "0" => Some(VK_0 as u32),
            "1" => Some(VK_1 as u32),
            "2" => Some(VK_2 as u32),
            "3" => Some(VK_3 as u32),
            "4" => Some(VK_4 as u32),
            "5" => Some(VK_5 as u32),
            "6" => Some(VK_6 as u32),
            "7" => Some(VK_7 as u32),
            "8" => Some(VK_8 as u32),
            "9" => Some(VK_9 as u32),
            "SPACE" => Some(VK_SPACE as u32),
            "ENTER" | "RETURN" => Some(VK_RETURN as u32),
            "TAB" => Some(VK_TAB as u32),
            "ESC" | "ESCAPE" => Some(VK_ESCAPE as u32),
            "DELETE" | "DEL" => Some(VK_DELETE as u32),
            "INSERT" | "INS" => Some(VK_INSERT as u32),
            "HOME" => Some(VK_HOME as u32),
            "END" => Some(VK_END as u32),
            "PAGEUP" | "PGUP" => Some(VK_PRIOR as u32),
            "PAGEDOWN" | "PGDN" => Some(VK_NEXT as u32),
            "UP" => Some(VK_UP as u32),
            "DOWN" => Some(VK_DOWN as u32),
            "LEFT" => Some(VK_LEFT as u32),
            "RIGHT" => Some(VK_RIGHT as u32),
            _ => None,
        }
    }

    pub unsafe fn hidden_window_proc(hwnd: HWND, msg: u32, wparam: WPARAM, _lparam: LPARAM) -> LRESULT {
        if msg == WM_HOTKEY {
            let id = wparam as u32;
            if let Some(tx) = HOTKEY_SENDER.lock().unwrap().as_ref() {
                let _ = tx.send(id);
            }
            return 0;
        }
        if msg == WM_QUIT {
            return 0;
        }
        DefWindowProcW(hwnd, msg, wparam, _lparam)
    }

    use std::sync::Mutex;

    static HOTKEY_SENDER: Mutex<Option<mpsc::Sender<u32>>> = Mutex::new(None);

    pub fn set_hotkey_sender(tx: mpsc::Sender<u32>) {
        *HOTKEY_SENDER.lock().unwrap() = Some(tx);
    }
}

pub struct HotkeyManager {
    registered: HashMap<String, u32>,
    next_id: u32,
}

impl HotkeyManager {
    pub fn new() -> Self {
        Self {
            registered: HashMap::new(),
            next_id: 1,
        }
    }

    #[cfg(windows)]
    pub fn register(&mut self, cmd: &HotkeyRegisterCmd) -> Result<(), String> {
        use windows_sys::Win32::UI::WindowsAndMessaging::*;

        let modifiers = win::parse_modifiers(&cmd.modifiers);
        let vk = win::parse_vkey(&cmd.key)
            .ok_or_else(|| format!("Unbekannte Taste: {}", cmd.key))?;

        let hotkey_id = self.next_id;
        self.next_id += 1;

        unsafe {
            let ok = RegisterHotKey(std::ptr::null_mut(), hotkey_id as i32, modifiers, vk);
            if ok == 0 {
                return Err(format!(
                    "RegisterHotKey fehlgeschlagen für {} (mods={}, key={})",
                    cmd.hotkey_id, modifiers, vk
                ));
            }
        }

        self.registered.insert(cmd.hotkey_id.clone(), hotkey_id);
        println!("[s-b] Hotkey registriert: {} -> ID {}", cmd.hotkey_id, hotkey_id);
        Ok(())
    }

    #[cfg(not(windows))]
    pub fn register(&mut self, cmd: &HotkeyRegisterCmd) -> Result<(), String> {
        let hotkey_id = self.next_id;
        self.next_id += 1;
        self.registered.insert(cmd.hotkey_id.clone(), hotkey_id);
        println!("[s-b] Hotkey registriert (Stub): {} -> ID {}", cmd.hotkey_id, hotkey_id);
        Ok(())
    }

    #[cfg(windows)]
    pub fn unregister_all(&self) {
        use windows_sys::Win32::UI::WindowsAndMessaging::*;
        for (_, id) in &self.registered {
            unsafe {
                UnregisterHotKey(std::ptr::null_mut(), *id as i32);
            }
        }
    }

    #[cfg(not(windows))]
    pub fn unregister_all(&self) {}

    pub fn find_by_id(&self, win_id: u32) -> Option<&String> {
        self.registered.iter().find(|(_, v)| **v == win_id).map(|(k, _)| k)
    }
}

#[cfg(windows)]
pub fn spawn_message_loop(tx: mpsc::Sender<u32>) -> std::thread::JoinHandle<()> {
    std::thread::spawn(move || {
        unsafe {
            win::set_hotkey_sender(tx);

            let class_name: Vec<u16> = "NexusHotkeyMsg\0".encode_utf16().collect();
            let wnd_class = WNDCLASSEXW {
                cbSize: std::mem::size_of::<WNDCLASSEXW>() as u32,
                style: 0,
                lpfnWndProc: Some(win::hidden_window_proc),
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
                std::ptr::null_mut(),
                std::ptr::null_mut(),
                windows_sys::Win32::System::LibraryLoader::GetModuleHandleW(std::ptr::null()),
                std::ptr::null(),
            );

            if hwnd.is_null() {
                eprintln!("[s-b] Fehler: Konnte Message-Window nicht erstellen");
                return;
            }

            let mut msg: MSG = std::mem::zeroed();
            while GetMessageW(&mut msg, hwnd, 0, 0) > 0 {
                TranslateMessage(&msg);
                DispatchMessageW(&msg);
            }
        }
    })
}

#[cfg(not(windows))]
pub fn spawn_message_loop(_tx: mpsc::Sender<u32>) -> std::thread::JoinHandle<()> {
    std::thread::spawn(move || {
        loop {
            std::thread::sleep(std::time::Duration::from_secs(3600));
        }
    })
}
