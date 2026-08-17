use std::collections::HashMap;

#[cfg(windows)]
mod win {
    use std::sync::mpsc;
    use std::sync::Mutex;
    use windows_sys::Win32::Foundation::*;
    use windows_sys::Win32::System::LibraryLoader::GetModuleHandleW;
    use windows_sys::Win32::UI::Input::KeyboardAndMouse::*;
    use windows_sys::Win32::UI::WindowsAndMessaging::*;

    pub const WM_HOTKEY: u32 = 786;

    pub fn parse_modifiers(mods: &[String]) -> HOT_KEY_MODIFIERS {
        let mut result: HOT_KEY_MODIFIERS = 0;
        for m in mods {
            match m.to_uppercase().as_str() {
                "ALT" | "MENU" => result |= MOD_ALT,
                "CTRL" | "CONTROL" => result |= MOD_CONTROL,
                "SHIFT" => result |= MOD_SHIFT,
                "WIN" | "META" => result |= MOD_WIN,
                _ => {}
            }
        }
        result
    }

    pub fn parse_vkey(key: &str) -> Option<u32> {
        match key.to_uppercase().as_str() {
            "F1" => Some(0x7A),
            "F2" => Some(0x7B),
            "F3" => Some(0x7C),
            "F4" => Some(0x7D),
            "F5" => Some(0x7E),
            "F6" => Some(0x7F),
            "F7" => Some(0x80),
            "F8" => Some(0x81),
            "F9" => Some(0x82),
            "F10" => Some(0x83),
            "F11" => Some(0x84),
            "F12" => Some(0x85),
            "A" => Some(0x41),
            "B" => Some(0x42),
            "C" => Some(0x43),
            "D" => Some(0x44),
            "E" => Some(0x45),
            "F" => Some(0x46),
            "G" => Some(0x47),
            "H" => Some(0x48),
            "I" => Some(0x49),
            "J" => Some(0x4A),
            "K" => Some(0x4B),
            "L" => Some(0x4C),
            "M" => Some(0x4D),
            "N" => Some(0x4E),
            "O" => Some(0x4F),
            "P" => Some(0x50),
            "Q" => Some(0x51),
            "R" => Some(0x52),
            "S" => Some(0x53),
            "T" => Some(0x54),
            "U" => Some(0x55),
            "V" => Some(0x56),
            "W" => Some(0x57),
            "X" => Some(0x58),
            "Y" => Some(0x59),
            "Z" => Some(0x5A),
            "0" => Some(0x30),
            "1" => Some(0x31),
            "2" => Some(0x32),
            "3" => Some(0x33),
            "4" => Some(0x34),
            "5" => Some(0x35),
            "6" => Some(0x36),
            "7" => Some(0x37),
            "8" => Some(0x38),
            "9" => Some(0x39),
            "SPACE" => Some(0x20),
            "ENTER" | "RETURN" => Some(0x0D),
            "TAB" => Some(0x09),
            "ESC" | "ESCAPE" => Some(0x1B),
            "DELETE" | "DEL" => Some(0x2E),
            "INSERT" | "INS" => Some(0x2D),
            "HOME" => Some(0x24),
            "END" => Some(0x23),
            "PAGEUP" | "PGUP" => Some(0x21),
            "PAGEDOWN" | "PGDN" => Some(0x22),
            "UP" => Some(0x26),
            "DOWN" => Some(0x28),
            "LEFT" => Some(0x25),
            "RIGHT" => Some(0x27),
            _ => None,
        }
    }

    unsafe extern "system" fn hidden_window_proc(
        hwnd: HWND,
        msg: u32,
        wparam: WPARAM,
        lparam: LPARAM,
    ) -> LRESULT {
        if msg == WM_HOTKEY {
            let id = wparam as u32;
            if let Ok(guard) = HOTKEY_SENDER.lock() {
                if let Some(ref tx) = *guard {
                    let _ = tx.send(id);
                }
            }
            return 0;
        }
        DefWindowProcW(hwnd, msg, wparam, lparam)
    }

    static HOTKEY_SENDER: Mutex<Option<mpsc::Sender<u32>>> = Mutex::new(None);

    pub fn set_hotkey_sender(tx: mpsc::Sender<u32>) {
        *HOTKEY_SENDER.lock().unwrap() = Some(tx);
    }

    pub fn register_hotkey(id: i32, modifiers: HOT_KEY_MODIFIERS, vk: u32) -> Result<(), String> {
        unsafe {
            let ok = RegisterHotKey(std::ptr::null_mut(), id, modifiers, vk);
            if ok == 0 {
                return Err(format!(
                    "RegisterHotKey fehlgeschlagen (mods={modifiers}, key={vk})"
                ));
            }
        }
        Ok(())
    }

    pub fn unregister_hotkey(id: i32) {
        unsafe {
            UnregisterHotKey(std::ptr::null_mut(), id);
        }
    }

    pub fn run_message_loop() {
        unsafe {
            let class_name: Vec<u16> = "NexusHotkeyMsg\0".encode_utf16().collect();
            let wnd_class = WNDCLASSEXW {
                cbSize: std::mem::size_of::<WNDCLASSEXW>() as u32,
                style: 0,
                lpfnWndProc: Some(hidden_window_proc),
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
                eprintln!("[s-b] Fehler: Konnte Hotkey-Window nicht erstellen");
                return;
            }

            let mut msg: MSG = std::mem::zeroed();
            while GetMessageW(&mut msg, hwnd, 0, 0) != 0 {
                TranslateMessage(&msg);
                DispatchMessageW(&msg);
            }
        }
    }
}

pub struct HotkeyManager {
    registered: HashMap<String, i32>,
    next_id: i32,
}

impl HotkeyManager {
    pub fn new() -> Self {
        Self {
            registered: HashMap::new(),
            next_id: 1,
        }
    }

    pub fn register(&mut self, cmd: &crate::bus::HotkeyRegisterCmd) -> Result<(), String> {
        #[cfg(windows)]
        {
            let modifiers = win::parse_modifiers(&cmd.modifiers);
            let vk = win::parse_vkey(&cmd.key)
                .ok_or_else(|| format!("Unbekannte Taste: {}", cmd.key))?;

            let hotkey_id = self.next_id;
            self.next_id += 1;

            win::register_hotkey(hotkey_id, modifiers, vk)?;
            self.registered.insert(cmd.hotkey_id.clone(), hotkey_id);
            println!(
                "[s-b] Hotkey registriert: {} -> ID {}",
                cmd.hotkey_id, hotkey_id
            );
            Ok(())
        }

        #[cfg(not(windows))]
        {
            let hotkey_id = self.next_id;
            self.next_id += 1;
            self.registered
                .insert(cmd.hotkey_id.clone(), hotkey_id);
            println!(
                "[s-b] Hotkey registriert (Stub): {} -> ID {}",
                cmd.hotkey_id, hotkey_id
            );
            Ok(())
        }
    }

    pub fn unregister_all(&self) {
        #[cfg(windows)]
        {
            for (_, id) in &self.registered {
                win::unregister_hotkey(*id);
            }
        }
    }

    pub fn find_by_id(&self, win_id: u32) -> Option<&String> {
        self.registered
            .iter()
            .find(|(_, v)| **v as u32 == win_id)
            .map(|(k, _)| k)
    }
}

pub fn spawn_message_loop(tx: std::sync::mpsc::Sender<u32>) -> std::thread::JoinHandle<()> {
    std::thread::spawn(move || {
        #[cfg(windows)]
        {
            win::set_hotkey_sender(tx);
            win::run_message_loop();
        }
        #[cfg(not(windows))]
        {
            let _ = tx;
            loop {
                std::thread::sleep(std::time::Duration::from_secs(3600));
            }
        }
    })
}
