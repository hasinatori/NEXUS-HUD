using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Runtime.InteropServices;
using System.Threading.Tasks;
using Microsoft.UI;
using Microsoft.UI.Input;
using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using Microsoft.UI.Xaml.Input;
using Microsoft.UI.Xaml.Media;
using NexusHud.UI.Bus;
using NexusHud.UI.ViewModels;
using NexusHud.UI.Widgets;
using Windows.UI;

namespace NexusHud.UI;

// Zuletzt geändert: 2026-08-17
public sealed partial class MainWindow : Window
{
    private static readonly TimeSpan KeepAliveInterval = TimeSpan.FromSeconds(5);
    private static readonly TimeSpan KeepAliveTimeout = TimeSpan.FromSeconds(15);

    private readonly BusClient _bus = new(
        new Uri("ws://127.0.0.1:49152/"),
        keepAliveInterval: KeepAliveInterval,
        keepAliveTimeout: KeepAliveTimeout);

    public MainViewModel Vm { get; } = new();
    public DashboardViewModel Dashboard { get; } = new();

    private readonly Dictionary<string, FrameworkElement> _widgetElements = new();

    public MainWindow()
    {
        InitializeComponent();
        SetupOverlay();
        LoadDashboard();
        Connect();
    }

    private void SetupOverlay()
    {
        var hwnd = GetWindowHandle();
        if (hwnd != IntPtr.Zero)
        {
            SetWindowPos(hwnd, HWND_TOPMOST, 0, 0, 0, 0,
                SWP_NOMOVE | SWP_NOSIZE | SWP_NOACTIVATE);

            var exStyle = GetWindowLong(hwnd, GWL_EXSTYLE);
            SetWindowLong(hwnd, GWL_EXSTYLE, exStyle | WS_EX_TOOLWINDOW);
        }
    }

    private void LoadDashboard()
    {
        try
        {
            var layoutPath = Path.Combine(AppContext.BaseDirectory, "layout.json");
            if (File.Exists(layoutPath))
            {
                Dashboard.Load(layoutPath);
            }
            else
            {
                Dashboard.ApplyLayout(WidgetLayout.Default());
            }
        }
        catch
        {
            Dashboard.ApplyLayout(WidgetLayout.Default());
        }

        Dashboard.Widgets.CollectionChanged += (_, _) => RebuildWidgetGrid();
        foreach (var widget in Dashboard.Widgets)
        {
            widget.PropertyChanged += OnWidgetPropertyChanged;
        }
        RebuildWidgetGrid();
    }

    private void RebuildWidgetGrid()
    {
        WidgetGrid.Children.Clear();
        WidgetGrid.RowDefinitions.Clear();
        WidgetGrid.ColumnDefinitions.Clear();
        _widgetElements.Clear();

        for (int r = 0; r < Dashboard.Rows; r++)
            WidgetGrid.RowDefinitions.Add(new RowDefinition { Height = new GridLength(1, GridUnitType.Star) });
        for (int c = 0; c < Dashboard.Columns; c++)
            WidgetGrid.ColumnDefinitions.Add(new ColumnDefinition { Width = new GridLength(1, GridUnitType.Star) });

        foreach (var widget in Dashboard.Widgets)
        {
            var element = CreateWidgetElement(widget);
            _widgetElements[widget.Id] = element;

            Grid.SetRow(element, widget.Row);
            Grid.SetColumn(element, widget.Column);
            if (widget.RowSpan > 1)
                Grid.SetRowSpan(element, widget.RowSpan);
            if (widget.ColumnSpan > 1)
                Grid.SetColumnSpan(element, widget.ColumnSpan);

            WidgetGrid.Children.Add(element);
        }
    }

    private FrameworkElement CreateWidgetElement(WidgetViewModel widget)
    {
        var border = new Border
        {
            Background = new SolidColorBrush(Color.FromArgb(40, 0, 229, 255)),
            CornerRadius = new CornerRadius(5),
            BorderBrush = new SolidColorBrush(Color.FromArgb(60, 0, 229, 255)),
            BorderThickness = new Thickness(1),
            Padding = new Thickness(6),
            Margin = new Thickness(2),
        };

        var stack = new StackPanel { Spacing = 2 };

        var titleBlock = new TextBlock
        {
            Text = widget.Title,
            FontSize = 10,
            FontWeight = Microsoft.UI.Text.FontWeights.SemiBold,
            Foreground = new SolidColorBrush(Color.FromArgb(255, 0, 229, 255)),
        };
        stack.Children.Add(titleBlock);

        var kindBlock = new TextBlock
        {
            Text = widget.KindLabel,
            FontSize = 8,
            Foreground = new SolidColorBrush(Color.FromArgb(255, 107, 114, 128)),
        };
        stack.Children.Add(kindBlock);

        var statusBlock = new TextBlock
        {
            Text = string.IsNullOrEmpty(widget.Status) ? "—" : widget.Status,
            FontSize = 18,
            FontWeight = Microsoft.UI.Text.FontWeights.Bold,
            Foreground = new SolidColorBrush(Color.FromArgb(255, 229, 231, 235)),
            TextWrapping = TextWrapping.Wrap,
        };
        stack.Children.Add(statusBlock);

        border.Child = stack;
        return border;
    }

    private void OnWidgetPropertyChanged(object? sender, System.ComponentModel.PropertyChangedEventArgs e)
    {
        if (sender is WidgetViewModel widget && e.PropertyName == nameof(WidgetViewModel.Status)
            && _widgetElements.TryGetValue(widget.Id, out var element)
            && element is Border border && border.Child is StackPanel stack
            && stack.Children.Count >= 3 && stack.Children[2] is TextBlock statusBlock)
        {
            statusBlock.Text = string.IsNullOrEmpty(widget.Status) ? "—" : widget.Status;
        }
    }

    private async void Connect()
    {
        _bus.Connected += () => DispatcherQueue.TryEnqueue(() =>
        {
            Vm.State = "Bus: verbunden";
            StatusBar.Text = "";
        });
        _bus.Disconnected += () => DispatcherQueue.TryEnqueue(() =>
        {
            Vm.State = "Bus: getrennt";
            StatusBar.Text = "Verbindung getrennt — Reconnect läuft…";
        });
        _bus.MessageReceived += OnMessage;

        await _bus.RunAsync(
            () => _bus.SendHelloAsync(Protocol.Source, Protocol.ServiceId, "0.2.0-s-a.1", Protocol.ProtocolVersion));
    }

    private void OnMessage(string json)
    {
        DispatcherQueue.TryEnqueue(() =>
        {
            if (json.Contains("event.system.hello"))
            {
                Vm.HelloCount++;
            }
            Vm.LastEvent = json;
            DispatchEventToWidgets(json);
        });
    }

    private void DispatchEventToWidgets(string json)
    {
        try
        {
            using var doc = System.Text.Json.JsonDocument.Parse(json);
            var root = doc.RootElement;
            if (!root.TryGetProperty("method", out var methodProp))
                return;

            var method = methodProp.GetString() ?? "";
            var ps = root.TryGetProperty("params", out var paramsProp) ? paramsProp : default;

            switch (method)
            {
                case "event.system.metrics":
                    UpdateMetricsWidgets(ps);
                    break;
                case "event.git.status":
                    UpdateGitWidget(ps);
                    break;
                case "event.build.succeeded":
                case "event.build.failed":
                    UpdateBuildWidget(ps, method);
                    break;
                case "event.media.state":
                    UpdateMediaWidget(ps);
                    break;
            }
        }
        catch { }
    }

    private void UpdateMetricsWidgets(System.Text.Json.JsonElement ps)
    {
        if (ps.TryGetProperty("cpu", out var cpuProp))
            Dashboard.SetStatus("cpu-gauge", $"{cpuProp.GetDouble():F1}%");

        if (ps.TryGetProperty("ram", out var ramProp)
            && ramProp.TryGetProperty("used_mb", out var usedProp)
            && ramProp.TryGetProperty("total_mb", out var totalProp))
        {
            Dashboard.SetStatus("ram-gauge", $"{usedProp.GetDouble() / 1024:F1}/{totalProp.GetDouble() / 1024:F1} GB");
        }
    }

    private void UpdateGitWidget(System.Text.Json.JsonElement ps)
    {
        var branch = ps.TryGetProperty("branch", out var b) ? b.GetString() ?? "" : "";
        var staged = ps.TryGetProperty("staged", out var s) ? s.GetInt32() : 0;
        var uncommitted = ps.TryGetProperty("uncommitted", out var u) ? u.GetInt32() : 0;
        var ahead = ps.TryGetProperty("ahead", out var a) ? a.GetInt32() : 0;
        var behind = ps.TryGetProperty("behind", out var bh) ? bh.GetInt32() : 0;

        var parts = new List<string>();
        if (!string.IsNullOrEmpty(branch)) parts.Add(branch);
        if (staged > 0) parts.Add($"staged:{staged}");
        if (uncommitted > 0) parts.Add($"uncommitted:{uncommitted}");
        if (ahead > 0) parts.Add($"ahead:{ahead}");
        if (behind > 0) parts.Add($"behind:{behind}");

        Dashboard.SetStatus("git-status", parts.Count > 0 ? string.Join(" · ", parts) : "clean");
    }

    private void UpdateBuildWidget(System.Text.Json.JsonElement ps, string method)
    {
        var project = ps.TryGetProperty("project", out var p) ? p.GetString() ?? "" : "";
        var ok = method == "event.build.succeeded";
        Dashboard.SetStatus("build-badge", $"{(ok ? "✓" : "✗")} {project}");
    }

    private void UpdateMediaWidget(System.Text.Json.JsonElement ps)
    {
        var playing = ps.TryGetProperty("playing", out var pl) && pl.GetBoolean();
        var track = ps.TryGetProperty("track", out var t) ? t.GetString() ?? "" : "";
        var artist = ps.TryGetProperty("artist", out var ar) ? ar.GetString() ?? "" : "";
        var text = string.IsNullOrEmpty(track) ? "Kein Track" : $"{(playing ? "▶" : "⏸")} {track}";
        if (!string.IsNullOrEmpty(artist)) text += $" — {artist}";
        Dashboard.SetStatus("media-card", text);
    }

    // --- cmd.* Button Handlers ---

    private async void OnMediaToggle(object sender, RoutedEventArgs e)
    {
        await SendCommandAsync(Protocol.CmdMediaToggle(), "media.toggle");
    }

    private async void OnMediaNext(object sender, RoutedEventArgs e)
    {
        await SendCommandAsync(Protocol.CmdMediaNext(), "media.next");
    }

    private async void OnAppLaunch(object sender, RoutedEventArgs e)
    {
        var dialog = new ContentDialog
        {
            Title = "App starten",
            PrimaryButtonText = "Starten",
            CloseButtonText = "Abbrechen",
            DefaultButton = ContentDialogButton.Primary,
            XamlRoot = this.Content.XamlRoot,
        };

        var pathBox = new TextBox { PlaceholderText = "Pfad oder Programmname", Margin = new Thickness(0, 8, 0, 0) };
        var argsBox = new TextBox { PlaceholderText = "Argumente (optional, kommagetrennt)", Margin = new Thickness(0, 8, 0, 0) };
        var focusBox = new CheckBox { Content = "Fenster fokussieren", IsChecked = true, Margin = new Thickness(0, 8, 0, 0) };

        var stack = new StackPanel();
        stack.Children.Add(new TextBlock { Text = "Programm- oder Dateipfad:" });
        stack.Children.Add(pathBox);
        stack.Children.Add(argsBox);
        stack.Children.Add(focusBox);
        dialog.Content = stack;

        var result = await dialog.ShowAsync();
        if (result == ContentDialogResult.Primary && !string.IsNullOrWhiteSpace(pathBox.Text))
        {
            var args = string.IsNullOrWhiteSpace(argsBox.Text)
                ? null
                : argsBox.Text.Split(',', StringSplitOptions.RemoveEmptyEntries | StringSplitOptions.TrimEntries);
            await SendCommandAsync(Protocol.CmdAppLaunch(pathBox.Text, args, focusBox.IsChecked == true), "app.launch");
        }
    }

    private async void OnAutomationRun(object sender, RoutedEventArgs e)
    {
        var dialog = new ContentDialog
        {
            Title = "Automation starten",
            PrimaryButtonText = "Starten",
            CloseButtonText = "Abbrechen",
            DefaultButton = ContentDialogButton.Primary,
            XamlRoot = this.Content.XamlRoot,
        };

        var taskBox = new TextBox { PlaceholderText = "Task-Name (aus S-C-Konfiguration)", Margin = new Thickness(0, 8, 0, 0) };
        var stack = new StackPanel();
        stack.Children.Add(new TextBlock { Text = "Name des Automation-Tasks:" });
        stack.Children.Add(taskBox);
        dialog.Content = stack;

        var result = await dialog.ShowAsync();
        if (result == ContentDialogResult.Primary && !string.IsNullOrWhiteSpace(taskBox.Text))
        {
            await SendCommandAsync(Protocol.CmdAutomationRun(taskBox.Text), "automation.run");
        }
    }

    private async void OnReconnect(object sender, RoutedEventArgs e)
    {
        StatusBar.Text = "Verbindung wird hergestellt…";
        await _bus.StopAsync();
        Connect();
    }

    private async void OnCloseClick(object sender, RoutedEventArgs e)
    {
        await _bus.StopAsync();
        Close();
    }

    private async Task SendCommandAsync(string json, string label)
    {
        try
        {
            await _bus.SendAsync(json);
            StatusBar.Text = $"✓ cmd.{label} gesendet";
        }
        catch (Exception ex)
        {
            StatusBar.Text = $"✗ cmd.{label} fehlgeschlagen: {ex.Message}";
        }
    }

    // --- Window Management ---

    private void OnHeaderPointerPressed(object sender, PointerRoutedEventArgs e)
    {
        if (e.Pointer.PointerDeviceType == Microsoft.UI.Input.PointerDeviceType.Mouse)
        {
            var pos = e.GetCurrentPoint(null);
            if (pos.Properties.IsLeftButtonPressed)
            {
                DragMove();
            }
        }
    }

    private void OnRootPointerPressed(object sender, PointerRoutedEventArgs e) { }

    private IntPtr GetWindowHandle()
    {
        var windowNative = this.As<IWindowNative>();
        return windowNative.WindowHandle;
    }

    [DllImport("user32.dll")]
    private static extern bool SetWindowPos(IntPtr hWnd, IntPtr hWndInsertAfter, int X, int Y, int cx, int cy, uint uFlags);

    [DllImport("user32.dll")]
    private static extern int GetWindowLong(IntPtr hWnd, int nIndex);

    [DllImport("user32.dll")]
    private static extern int SetWindowLong(IntPtr hWnd, int nIndex, int dwNewLong);

    [ComImport]
    [Guid("EECDBF0E-19C6-4578-9165-4F3F1516A1E3")]
    [InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
    private interface IWindowNative
    {
        IntPtr WindowHandle { get; }
    }

    private const int GWL_EXSTYLE = -20;
    private const int WS_EX_TOOLWINDOW = 0x00000080;
    private const uint SWP_NOMOVE = 0x0002;
    private const uint SWP_NOSIZE = 0x0001;
    private const uint SWP_NOACTIVATE = 0x0010;
    private static readonly IntPtr HWND_TOPMOST = new(-1);

    private void DragMove()
    {
        var hwnd = GetWindowHandle();
        if (hwnd == IntPtr.Zero) return;
        ReleaseCapture();
        SendMessage(hwnd, 0x00A1, 0x0002, 0);
    }

    [DllImport("user32.dll")]
    private static extern bool ReleaseCapture();

    [DllImport("user32.dll")]
    private static extern int SendMessage(IntPtr hWnd, int Msg, int wParam, int lParam);
}
