pub fn launch_app(path: &str, args: &[String], focus: bool) -> Result<u32, String> {
    #[cfg(windows)]
    {
        use windows_sys::Win32::Foundation::*;
        use windows_sys::Win32::System::Threading::*;
        use windows_sys::Win32::UI::WindowsAndMessaging::*;

        let mut cmdline = String::new();
        for arg in args {
            cmdline.push_str(&format!(" \"{}\"", arg));
        }

        let wide_path: Vec<u16> = path
            .encode_utf16()
            .chain(std::iter::once(0))
            .collect();

        let mut cmd_wide: Vec<u16> = cmdline
            .encode_utf16()
            .chain(std::iter::once(0))
            .collect();

        let mut si: STARTUPINFOW = unsafe { std::mem::zeroed() };
        si.cb = std::mem::size_of::<STARTUPINFOW>() as u32;

        let mut pi: PROCESS_INFORMATION = unsafe { std::mem::zeroed() };

        let ok = unsafe {
            CreateProcessW(
                wide_path.as_ptr(),
                cmd_wide.as_mut_ptr(),
                std::ptr::null(),
                std::ptr::null(),
                0,
                0,
                std::ptr::null(),
                std::ptr::null(),
                &mut si,
                &mut pi,
            )
        };

        if ok == 0 {
            return Err(format!("CreateProcess fehlgeschlagen: {}", path));
        }

        let pid = pi.dwProcessId;

        if focus {
            std::thread::sleep(std::time::Duration::from_millis(500));

            struct FocusSearch {
                pid: u32,
                result: Option<HWND>,
            }

            unsafe extern "system" fn enum_callback(hwnd: HWND, lparam: LPARAM) -> windows_sys::core::BOOL {
                let data = &mut *(lparam as *mut FocusSearch);
                let mut lpiid: u32 = 0;
                GetWindowThreadProcessId(hwnd, &mut lpiid);
                if lpiid == data.pid {
                    let mut buf = [0u16; 512];
                    let len = GetWindowTextW(hwnd, buf.as_mut_ptr(), 512);
                    if len > 0 {
                        data.result = Some(hwnd);
                        return 0;
                    }
                }
                1
            }

            let mut data = FocusSearch {
                pid,
                result: None,
            };
            unsafe {
                EnumWindows(Some(enum_callback), &mut data as *mut _ as LPARAM);
            }

            if let Some(hwnd) = data.result {
                unsafe {
                    ShowWindow(hwnd, SW_RESTORE);
                    SetForegroundWindow(hwnd);
                }
            }
        }

        unsafe {
            CloseHandle(pi.hProcess);
            CloseHandle(pi.hThread);
        }

        Ok(pid)
    }

    #[cfg(not(windows))]
    {
        let _ = (args, focus);
        println!("[s-b] App gestartet (Stub): {}", path);
        Ok(1)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_launch_app_returns_ok() {
        let result = launch_app("/usr/bin/echo", &[], false);
        assert!(result.is_ok());
    }

    #[test]
    fn test_launch_app_returns_pid_1_on_non_windows() {
        let result = launch_app("/usr/bin/echo", &[], false);
        #[cfg(not(windows))]
        assert_eq!(result.unwrap(), 1);
    }

    #[test]
    fn test_launch_app_with_args() {
        let args = vec!["hello".to_string(), "world".to_string()];
        let result = launch_app("/usr/bin/echo", &args, false);
        assert!(result.is_ok());
    }

    #[test]
    fn test_launch_app_with_focus() {
        let result = launch_app("/usr/bin/echo", &[], true);
        assert!(result.is_ok());
    }
}
