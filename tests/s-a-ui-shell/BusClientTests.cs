using System.Linq;
using System.Net.WebSockets;
using System.Text.Json;
using NexusHud.UI.Bus;
using Xunit;

namespace NexusHud.UI.Tests;

// Zuletzt geändert: 2026-08-16
// BusClient: Hello senden, Empfang von Server-Frames, Auto-Reconnect nach Abriss,
// Heartbeat-Senden und Watchdog-Erkennung bei hartem Netzabriss (ohne Close-Frame).

public sealed class BusClientTests
{
    private const string HelloJson = "{\"jsonrpc\":\"2.0\",\"method\":\"event.system.hello\",\"params\":{\"source\":\"bus\"}}";

    private static async Task UntilAsync(Func<bool> condition, int timeoutMs = 5000)
    {
        var deadline = DateTime.UtcNow.AddMilliseconds(timeoutMs);
        while (!condition())
        {
            if (DateTime.UtcNow > deadline)
            {
                throw new TimeoutException("Bedingung nicht erfüllt innerhalb des Timeouts.");
            }
            await Task.Delay(50);
        }
    }

    [Fact]
    public async Task Hello_Connect_Und_MessageReceived_Funktionieren()
    {
        await using var server = new TestWsServer();
        var bus = new BusClient(server.Uri, reconnectDelay: TimeSpan.FromMilliseconds(100));
        var received = new List<string>();

        bus.MessageReceived += received.Add;
        var connected = new TaskCompletionSource();
        bus.Connected += () => connected.TrySetResult();

        var runTask = bus.RunAsync(() => bus.SendHelloAsync("S-A", "s-a-ui-shell", "0.2.0-s-a.1", 1));

        await connected.Task.WaitAsync(TimeSpan.FromSeconds(5));
        await UntilAsync(() => server.Received.Count >= 1);

        var hello = server.Received.Single();
        using var doc = JsonDocument.Parse(hello);
        Assert.Equal("2.0", doc.RootElement.GetProperty("jsonrpc").GetString());
        Assert.Equal("event.system.hello", doc.RootElement.GetProperty("method").GetString());
        var p = doc.RootElement.GetProperty("params");
        Assert.Equal("S-A", p.GetProperty("source").GetString());
        Assert.Equal(1, p.GetProperty("protocol_version").GetInt32());
        Assert.Equal("s-a-ui-shell", p.GetProperty("service_id").GetString());
        Assert.Equal("0.2.0-s-a.1", p.GetProperty("version").GetString());
        Assert.False(string.IsNullOrEmpty(p.GetProperty("ts").GetString()));

        await server.SendAsync(HelloJson);
        await UntilAsync(() => received.Count >= 1);
        Assert.Contains("event.system.hello", received[0]);

        await bus.StopAsync();
        await runTask;
    }

    [Fact]
    public async Task AutoReconnect_Nach_Abriss()
    {
        await using var server = new TestWsServer();
        var bus = new BusClient(server.Uri, reconnectDelay: TimeSpan.FromMilliseconds(200));
        long helloCount = 0;
        long connectedCount = 0;
        long disconnectedCount = 0;

        bus.Connected += () => Interlocked.Increment(ref connectedCount);
        bus.Disconnected += () => Interlocked.Increment(ref disconnectedCount);
        bus.MessageReceived += _ => Interlocked.Increment(ref helloCount);

        using var cts = new CancellationTokenSource();
        var runTask = bus.RunAsync(
            () =>
            {
                Interlocked.Increment(ref helloCount);
                return bus.SendHelloAsync("S-A", "s-a-ui-shell", "0.2.0-s-a.1", 1, cts.Token);
            });

        // Erste Verbindung + Hello.
        await UntilAsync(() => Interlocked.Read(ref helloCount) >= 1);
        await UntilAsync(() => server.ConnectionCount >= 1);

        // Ausfall -> Reconnect -> zweites Hello.
        await server.CloseLastConnectionAsync();
        await UntilAsync(() => server.ConnectionCount >= 2, 8000);
        await UntilAsync(() => Interlocked.Read(ref helloCount) >= 2, 8000);

        Assert.True(Interlocked.Read(ref disconnectedCount) >= 1);
        Assert.True(Interlocked.Read(ref connectedCount) >= 2);

        await bus.StopAsync();
        await runTask;
    }

    [Fact]
    public async Task SendOhneVerbindung_Wirft()
    {
        var bus = new BusClient(new Uri("ws://127.0.0.1:1/"));
        await Assert.ThrowsAsync<InvalidOperationException>(() => bus.SendAsync("{}"));
    }

    [Fact]
    public async Task Disconnected_Wird_Nach_Abriss_Gemeldet()
    {
        await using var server = new TestWsServer();
        var bus = new BusClient(server.Uri, reconnectDelay: TimeSpan.FromMilliseconds(100));
        var disconnected = new TaskCompletionSource();

        bus.Disconnected += () => disconnected.TrySetResult();
        var runTask = bus.RunAsync(() => Task.CompletedTask);

        await UntilAsync(() => server.ConnectionCount >= 1);
        await server.CloseLastConnectionAsync();

        await disconnected.Task.WaitAsync(TimeSpan.FromSeconds(5));
        await bus.StopAsync();
        await runTask;
    }

    [Fact]
    public async Task Heartbeat_Wird_Periodisch_Gesendet()
    {
        await using var server = new TestWsServer();
        var bus = new BusClient(server.Uri, reconnectDelay: TimeSpan.FromMilliseconds(100), keepAliveInterval: TimeSpan.FromMilliseconds(100));
        var runTask = bus.RunAsync(() => bus.SendHelloAsync("S-A", "s-a-ui-shell", "0.2.0-s-a.1", 1));

        await UntilAsync(() => server.ConnectionCount >= 1);

        // Ohne Watchdog kein Inbound -> Heartbeats fließen vom Client zum Server.
        await UntilAsync(() => server.Received.Count(m => m.Contains("\"event.system.heartbeat\"")) >= 2);

        var heartbeat = server.Received.First(m => m.Contains("\"event.system.heartbeat\""));
        using var doc = JsonDocument.Parse(heartbeat);
        Assert.Equal("event.system.heartbeat", doc.RootElement.GetProperty("method").GetString());
        var p = doc.RootElement.GetProperty("params");
        Assert.Equal("S-A", p.GetProperty("source").GetString());
        Assert.Equal(1, p.GetProperty("protocol_version").GetInt32());
        Assert.Equal("s-a-ui-shell", p.GetProperty("service_id").GetString());
        Assert.False(string.IsNullOrEmpty(p.GetProperty("ts").GetString()));

        await bus.StopAsync();
        await runTask;
    }

    [Fact]
    public async Task HarteNetzunterbrechung_Ohne_Close_Frame_Fuehrt_Zu_Reconnect()
    {
        await using var server = new TestWsServer();
        var bus = new BusClient(
            server.Uri,
            reconnectDelay: TimeSpan.FromMilliseconds(100),
            keepAliveTimeout: TimeSpan.FromMilliseconds(400));
        long helloCount = 0;
        long disconnectedCount = 0;

        bus.Disconnected += () => Interlocked.Increment(ref disconnectedCount);

        var runTask = bus.RunAsync(
            () =>
            {
                Interlocked.Increment(ref helloCount);
                return bus.SendHelloAsync("S-A", "s-a-ui-shell", "0.2.0-s-a.1", 1);
            });

        await UntilAsync(() => server.ConnectionCount >= 1);
        await UntilAsync(() => Interlocked.Read(ref helloCount) >= 1);

        // Server bleibt verbunden, sendet aber nie etwas (simulierter harter Abriss:
        // der Client bekommt keinen Close-Frame). Der Watchdog muss die Verbindung
        // nach keepAliveTimeout per Abort beenden -> Reconnect.
        await UntilAsync(() => server.ConnectionCount >= 2, 8000);
        await UntilAsync(() => Interlocked.Read(ref helloCount) >= 2, 8000);

        Assert.True(Interlocked.Read(ref disconnectedCount) >= 1);

        await bus.StopAsync();
        await runTask;
    }
}
