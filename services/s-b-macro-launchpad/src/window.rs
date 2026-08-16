use crate::bus::WindowMoveCmd;

#[cfg(windows)]
mod win {
    use windows_sys::Win32::UI::WindowsAndMessaging::*;
    use windows_sys::Win32::Foundation::*;

    pub fn move_window(title: &str, x: i32, y: i32, width: u32, height: u32) -> Result<(), String> {
        struct EnumData {
            title: String,
            found: bool,
        }

        unsafe extern "system" fn enum_callback(hwnd: HWND, lparam: LPARAM) -> BOOL {
            let data = &mut *(lparam as *mut EnumData);
            let mut buf = [0u16; 512];
            let len = GetWindowTextW(hwnd, buf.as_mut_ptr(), 512);
            if len > 0 {
                let window_title = String::from_utf16_lossy(&buf[..len as usize]);
                if window_title.to_lowercase().contains(&data.title.to_lowercase()) {
                    SetWindowPos(
                        hwnd,
                        std::ptr::null_mut(),
                        data.found as i32 * 0 + x,
                        data.found as i32 * 0 + y,
                        width as i32,
                        height as i32,
                        SWP_NOZORDER | SWP_NOACTIVATE,
                    );
                    data.found = true;
                    return 0;
                }
            }
            1
        }

        let mut data = EnumData {
            title: title.to_string(),
            found: false,
        };

        unsafe {
            EnumWindows(Some(enum_callback), &mut data as *mut EnumData as LPARAM);
        }

        if data.found {
            Ok(())
        } else {
            Err(format!("Fenster '{title}' nicht gefunden"))
        }
    }

    pub fn get_window_rect(title: &str) -> Result<(i32, i32, u32, u32), String> {
        struct EnumData {
            title: String,
            rect: Option<(i32, i32, u32, u32)>,
        }

        unsafe extern "system" fn enum_callback(hwnd: HWND, lparam: LPARAM) -> BOOL {
            let data = &mut *(lparam as *mut EnumData);
            let mut buf = [0u16; 512];
            let len = GetWindowTextW(hwnd, buf.as_mut_ptr(), 512);
            if len > 0 {
                let window_title = String::from_utf16_lossy(&buf[..len as usize]);
                if window_title.to_lowercase().contains(&data.title.to_lowercase()) {
                    let mut rect: RECT = std::mem::zeroed();
                    GetWindowRect(hwnd, &mut rect);
                    data.rect = Some((
                        rect.left,
                        rect.top,
                        (rect.right - rect.left) as u32,
                        (rect.bottom - rect.top) as u32,
                    ));
                    return 0;
                }
            }
            1
        }

        let mut data = EnumData {
            title: title.to_string(),
            rect: None,
        };

        unsafe {
            EnumWindows(Some(enum_callback), &mut data as *mut EnumData as LPARAM);
        }

        data.rect.ok_or_else(|| format!("Fenster '{title}' nicht gefunden"))
    }
}

pub fn move_window(cmd: &WindowMoveCmd) -> Result<(), String> {
    #[cfg(windows)]
    {
        let result = win::move_window(&cmd.window_title, cmd.x, cmd.y, cmd.width, cmd.height);
        if result.is_ok() {
            println!(
                "[s-b] Fenster verschoben: '{}' -> ({}, {}) {}x{}",
                cmd.window_title, cmd.x, cmd.y, cmd.width, cmd.height
            );
        }
        result
    }

    #[cfg(not(windows))]
    {
        let _ = cmd;
        println!("[s-b] Fenster verschoben (Stub): {}", cmd.window_title);
        Ok(())
    }
}
