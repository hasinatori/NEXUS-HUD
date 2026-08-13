using System;
using System.Runtime.InteropServices;
using System.Threading;
using Microsoft.UI.Dispatching;
using Microsoft.UI.Xaml;

namespace NexusHud.UI;

// Zuletzt geändert: 2026-08-13
public partial class App : Application
{
    private Window? _window;

    public App()
    {
        InitializeComponent();
    }

    protected override void OnLaunched(LaunchActivatedEventArgs args)
    {
        _window = new MainWindow();
        _window.Activate();
    }
}

internal static class Program
{
    [DllImport("Microsoft.ui.xaml.dll")]
    private static extern void XamlCheckProcessRequirements();

    [STAThread]
    private static void Main(string[] args)
    {
        // Windows App SDK Runtime 1.6 bootstrappen (unpackaged deployment).
        uint hr = Microsoft.Windows.ApplicationModel.DynamicDependency.Bootstrap.TryInitialize(0x00010006, out _);
        if (hr != 0)
        {
            throw new InvalidOperationException(
                $"Windows App SDK (Runtime 1.6) nicht initialisierbar (HRESULT 0x{hr:X8}). " +
                "Bitte Runtime installieren: https://aka.ms/windowsappsdk/1.6/latest/windowsappruntimeinstall-x64.exe");
        }

        XamlCheckProcessRequirements();
        WinRT.ComWrappersSupport.InitializeComWrappers();

        Application.Start(_ =>
        {
            var context = new DispatcherQueueSynchronizationContext(
                DispatcherQueue.GetForCurrentThread());
            SynchronizationContext.SetSynchronizationContext(context);
            new App();
        });

        Microsoft.Windows.ApplicationModel.DynamicDependency.Bootstrap.Shutdown();
    }
}
