using System;
using System.Threading.Tasks;
using Microsoft.UI.Xaml;
using NexusHud.UI.Bus;
using NexusHud.UI.ViewModels;

namespace NexusHud.UI;

// Zuletzt geändert: 2026-08-16
public sealed partial class MainWindow : Window
{
    private static readonly TimeSpan KeepAliveInterval = TimeSpan.FromSeconds(5);
    private static readonly TimeSpan KeepAliveTimeout = TimeSpan.FromSeconds(15);

    private readonly BusClient _bus = new(
        new Uri("ws://127.0.0.1:49152/"),
        keepAliveInterval: KeepAliveInterval,
        keepAliveTimeout: KeepAliveTimeout);

    public MainViewModel Vm { get; } = new();

    public MainWindow()
    {
        InitializeComponent();
        Connect();
    }

    private async void Connect()
    {
        _bus.Connected += () => Vm.State = "Bus: verbunden";
        _bus.Disconnected += () => Vm.State = "Bus: getrennt";
        _bus.MessageReceived += OnMessage;

        await _bus.RunAsync(
            () => _bus.SendHelloAsync(Protocol.Source, Protocol.ServiceId, "0.2.0-s-a.1", Protocol.ProtocolVersion));
    }

    private void OnMessage(string json)
    {
        if (json.Contains("event.system.hello"))
        {
            Vm.HelloCount++;
        }
        Vm.LastEvent = json;
    }
}
