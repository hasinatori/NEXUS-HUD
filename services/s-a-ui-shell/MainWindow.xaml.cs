using System;
using Microsoft.UI.Xaml;
using NexusHud.UI.Bus;
using NexusHud.UI.ViewModels;

namespace NexusHud.UI;

// Zuletzt geändert: 2026-08-13
public sealed partial class MainWindow : Window
{
    private readonly BusClient _bus = new(new Uri("ws://127.0.0.1:49152/"));

    public MainViewModel Vm { get; } = new();

    public MainWindow()
    {
        InitializeComponent();
        Connect();
    }

    private async void Connect()
    {
        Vm.State = "Bus: verbinde…";
        try
        {
            await _bus.ConnectAsync();
        }
        catch (Exception ex)
        {
            Vm.State = $"Bus: getrennt ({ex.GetType().Name})";
            return;
        }

        Vm.State = "Bus: verbunden";
        _bus.MessageReceived += OnMessage;
        await _bus.SendHelloAsync(Protocol.Source, Protocol.ServiceId, "0.2.0-s-a.1", Protocol.ProtocolVersion);
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
