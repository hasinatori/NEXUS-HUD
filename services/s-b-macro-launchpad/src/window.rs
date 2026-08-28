#[cfg(windows)]
mod win {
    use windows_sys::Win32::Foundation::*;
    use windows_sys::Win32::UI::WindowsAndMessaging::*;

    pub unsafe fn find_window_by_title(title: &str) -> Option<HWND> {
        struct SearchData {
            target: String,
            result: Option<HWND>,
        }

        unsafe extern "system" fn callback(hwnd: HWND, lparam: LPARAM) -> windows_sys::core::BOOL {
            let data = &mut *(lparam as *mut SearchData);
            let mut buf = [0u16; 512];
            let len = GetWindowTextW(hwnd, buf.as_mut_ptr(), 512);
            if len > 0 {
                let window_title = String::from_utf16_lossy(&buf[..len as usize]);
                if window_title == data.target {
                    data.result = Some(hwnd);
                }
            }
            1
        }

        let mut data = SearchData {
            target: title.to_string(),
            result: None,
        };
        EnumWindows(Some(callback), &mut data as *mut _ as LPARAM);
        data.result
    }
}

#[cfg(not(windows))]
mod win {
    pub unsafe fn find_window_by_title(_title: &str) -> Option<*mut core::ffi::c_void> {
        None
    }
}

pub fn find_window(title: &str) -> Option<usize> {
    unsafe { win::find_window_by_title(title).map(|h| h as usize) }
}

pub fn get_window_rect(hwnd: usize) -> Option<(i32, i32, u32, u32)> {
    #[cfg(windows)]
    {
        use windows_sys::Win32::Foundation::*;
        use windows_sys::Win32::UI::WindowsAndMessaging::*;

        let mut rect: RECT = unsafe { std::mem::zeroed() };
        let ok = unsafe { GetWindowRect(hwnd as HWND, &mut rect) };
        if ok != 0 {
            Some((
                rect.left,
                rect.top,
                (rect.right - rect.left) as u32,
                (rect.bottom - rect.top) as u32,
            ))
        } else {
            None
        }
    }

    #[cfg(not(windows))]
    {
        let _ = hwnd;
        None
    }
}

pub fn set_window_pos(hwnd: usize, x: i32, y: i32, width: u32, height: u32) -> bool {
    #[cfg(windows)]
    {
        use windows_sys::Win32::Foundation::*;
        use windows_sys::Win32::UI::WindowsAndMessaging::*;
        let ok = unsafe {
            SetWindowPos(
                hwnd as HWND,
                std::ptr::null_mut(),
                x,
                y,
                width as i32,
                height as i32,
                SWP_NOZORDER,
            )
        };
        ok != 0
    }

    #[cfg(not(windows))]
    {
        let _ = (hwnd, x, y, width, height);
        false
    }
}

pub fn focus_window(hwnd: usize) {
    #[cfg(windows)]
    {
        use windows_sys::Win32::Foundation::*;
        use windows_sys::Win32::UI::WindowsAndMessaging::*;
        unsafe {
            ShowWindow(hwnd as HWND, SW_RESTORE);
            SetForegroundWindow(hwnd as HWND);
        }
    }

    #[cfg(not(windows))]
    {
        let _ = hwnd;
    }
}

pub fn get_foreground_window() -> usize {
    #[cfg(windows)]
    {
        use windows_sys::Win32::UI::WindowsAndMessaging::*;
        unsafe { GetForegroundWindow() as usize }
    }

    #[cfg(not(windows))]
    {
        0
    }
}

pub fn get_window_title(hwnd: usize) -> String {
    #[cfg(windows)]
    {
        use windows_sys::Win32::Foundation::*;
        use windows_sys::Win32::UI::WindowsAndMessaging::*;
        let mut buf = [0u16; 512];
        let len = unsafe { GetWindowTextW(hwnd as HWND, buf.as_mut_ptr(), 512) };
        if len > 0 {
            String::from_utf16_lossy(&buf[..len as usize])
        } else {
            String::new()
        }
    }

    #[cfg(not(windows))]
    {
        let _ = hwnd;
        String::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_find_window_returns_none_on_non_windows() {
        assert_eq!(find_window("Nonexistent"), None);
    }

    #[test]
    fn test_get_window_rect_returns_none_on_non_windows() {
        assert_eq!(get_window_rect(0), None);
    }

    #[test]
    fn test_set_window_pos_returns_false_on_non_windows() {
        assert_eq!(set_window_pos(0, 100, 200, 800, 600), false);
    }

    #[test]
    fn test_focus_window_does_not_panic() {
        focus_window(0);
    }

    #[test]
    fn test_get_foreground_window_returns_zero_on_non_windows() {
        assert_eq!(get_foreground_window(), 0);
    }

    #[test]
    fn test_get_window_title_returns_empty_on_non_windows() {
        assert_eq!(get_window_title(0), String::new());
    }

    #[test]
    fn test_set_window_pos_with_various_values() {
        assert_eq!(set_window_pos(0, 0, 0, 1, 1), false);
        assert_eq!(set_window_pos(0, -100, -200, 1920, 1080), false);
        assert_eq!(set_window_pos(42, 50, 50, 400, 300), false);
    }
}
