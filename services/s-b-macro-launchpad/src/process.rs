use crate::bus::AppLaunchCmd;

#[cfg(windows)]
mod win {
    use std::ffi::c_void;
    use windows_sys::Win32::System::Threading::*;
    use windows_sys::Win32::Foundation::*;
    use windows_sys::Win32::UI::WindowsAndMessaging::*;

    pub struct ProcessHandle(HANDLE);

    impl Drop for ProcessHandle {
        fn drop(&mut self) {
            unsafe { CloseHandle(self.0); }
        }
    }

    pub fn launch_process(path: &str, args: &[String], focus: bool) -> Result<u32, String> {
        let mut cmd_line = format!("\"{path}\"");
        for arg in args {
            cmd_line.push(' ');
            cmd_line.push_str(arg);
        }

        let wide_cmd: Vec<u16> = cmd_line.encode_utf16().chain(std::iter::once(0)).collect();
        let wide_path: Vec<u16> = std::path::Path::new(path)
            .parent()
            .map(|p| p.to_string_lossy().to_string())
            .unwrap_or_default();
        let wide_dir: Vec<u16> = wide_path.encode_utf16().chain(std::iter::once(0)).collect();

        let mut si: STARTUPINFOW = unsafe { std::mem::zeroed() };
        si.cb = std::mem::size_of::<STARTUPINFOW>() as u32;

        let mut pi: PROCESS_INFORMATION = unsafe { std::mem::zeroed() };

        let ok = unsafe {
            CreateProcessW(
                std::ptr::null(),
                wide_cmd.as_ptr(),
                std::ptr::null(),
                std::ptr::null(),
                0,
                0,
                std::ptr::null(),
                if wide_dir.is_empty() { std::ptr::null() } else { wide_dir.as_ptr() },
                &mut si,
                &mut pi,
            )
        };

        if ok == 0 {
            return Err(format!("CreateProcess fehlgeschlagen für {path}"));
        }

        let _ph = ProcessHandle(pi.hProcess);

        if focus {
            unsafe {
                AllowSetForegroundWindow(pi.dwProcessId);
                let fg_hwnd = GetForegroundWindow();
                if !fg_hwnd.is_null() {
                    SetForegroundWindow(fg_hwnd);
                }
            }
        }

        Ok(pi.dwProcessId)
    }

    pub fn focus_process_by_name(name: &str) -> Result<(), String> {
        use windows_sys::Win32::UI::WindowsAndMessaging::*;

        struct EnumData {
            name: String,
            found: bool,
        }

        unsafe extern "system" fn enum_callback(hwnd: HWND, lparam: LPARAM) -> BOOL {
            let data = &mut *(lparam as *mut EnumData);
            let mut buf = [0u16; 512];
            let len = GetWindowTextW(hwnd, buf.as_mut_ptr(), 512);
            if len > 0 {
                let title = String::from_utf16_lossy(&buf[..len as usize]);
                if title.to_lowercase().contains(&data.name.to_lowercase()) {
                    ShowWindow(hwnd, SW_RESTORE);
                    SetForegroundWindow(hwnd);
                    data.found = true;
                    return 0;
                }
            }
            1
        }

        let mut data = EnumData {
            name: name.to_string(),
            found: false,
        };

        unsafe {
            EnumWindows(Some(enum_callback), &mut data as *mut EnumData as LPARAM);
        }

        if data.found {
            Ok(())
        } else {
            Err(format!("Fenster '{name}' nicht gefunden"))
        }
    }
}

pub fn launch(cmd: &AppLaunchCmd) -> Result<u32, String> {
    #[cfg(windows)]
    {
        let pid = win::launch_process(&cmd.path, &cmd.args, cmd.focus)?;
        println!("[s-b] Prozess gestartet: {} (PID {})", cmd.path, pid);
        Ok(pid)
    }

    #[cfg(not(windows))]
    {
        let _ = cmd;
        println!("[s-b] Prozess gestartet (Stub): {}", cmd.path);
        Ok(0)
    }
}

pub fn focus_by_name(name: &str) -> Result<(), String> {
    #[cfg(windows)]
    {
        win::focus_process_by_name(name)
    }

    #[cfg(not(windows))]
    {
        let _ = name;
        println!("[s-b] Fokus (Stub): {name}");
        Ok(())
    }
}
