using System.Collections.Concurrent;
using System.Net;
using System.Net.Sockets;
using System.Net.WebSockets;

namespace NexusHud.UI.Tests;

// Zuletzt geändert: 2026-08-14
// Test-WebSocket-Server auf HttpListener (funktioniert auf Linux/Windows).
// Nimmt beliebig viele Verbindungen an, sammelt empfangene Frames und kann
// eine Verbindung gezielt abbrechen (für Reconnect-Tests).

public sealed class TestWsServer : IAsyncDisposable
{
    private readonly HttpListener _listener = new();
    private readonly CancellationTokenSource _cts = new();
    private readonly Task _acceptLoop;
    private readonly List<WebSocket> _sockets = new();

    public int Port { get; }
    public ConcurrentBag<string> Received { get; } = new();
    public int ConnectionCount { get; private set; }

    public TestWsServer()
    {
        Port = FreePort();
        _listener.Prefixes.Add($"http://127.0.0.1:{Port}/");
        _listener.Start();
        _acceptLoop = AcceptLoopAsync();
    }

    public Uri Uri => new($"ws://127.0.0.1:{Port}/");

    private static int FreePort()
    {
        using var tcp = new TcpListener(IPAddress.Loopback, 0);
        tcp.Start();
        return ((IPEndPoint)tcp.LocalEndpoint).Port;
    }

    private async Task AcceptLoopAsync()
    {
        try
        {
            while (!_cts.IsCancellationRequested)
            {
                var ctx = await _listener.GetContextAsync();
                if (_cts.IsCancellationRequested)
                {
                    ctx.Response.Abort();
                    return;
                }
                if (!ctx.Request.IsWebSocketRequest)
                {
                    ctx.Response.StatusCode = (int)HttpStatusCode.BadRequest;
                    ctx.Response.Close();
                    continue;
                }
                var ws = (await ctx.AcceptWebSocketAsync(null)).WebSocket;
                lock (_sockets)
                {
                    _sockets.Add(ws);
                }
                ConnectionCount++;
                _ = ReadLoopAsync(ws);
            }
        }
        catch (Exception)
        {
            // Listener beendet
        }
    }

    private async Task ReadLoopAsync(WebSocket ws)
    {
        var buffer = new byte[16 * 1024];
        try
        {
            while (ws.State == WebSocketState.Open)
            {
                var result = await ws.ReceiveAsync(buffer, _cts.Token);
                if (result.MessageType == WebSocketMessageType.Close)
                {
                    break;
                }
                Received.Add(System.Text.Encoding.UTF8.GetString(buffer, 0, result.Count));
            }
        }
        catch
        {
            // Verbindung beendet
        }
    }

    /// <summary>Schließt die zuletzt angenommene Verbindung sauber (Bus-Ausfall).
    /// Auf Linux erkennt der Client nur echte Close-Frames zuverlässig
    /// (Abort/Netzabriss wird dort erst per Keepalive bemerkt).</summary>
    public async Task CloseLastConnectionAsync()
    {
        WebSocket? ws;
        lock (_sockets)
        {
            ws = _sockets.Count > 0 ? _sockets[^1] : null;
        }
        if (ws == null)
        {
            return;
        }
        try
        {
            // CloseOutputAsync sendet den Close-Frame ohne auf die Handshake-
            // Antwort zu warten — der Client erkennt das zuverlässig.
            await ws.CloseOutputAsync(WebSocketCloseStatus.NormalClosure, "bus stop", CancellationToken.None);
        }
        catch (Exception)
        {
            // bereits geschlossen
        }
    }

    /// <summary>Sendet einen Text-Frame an die zuletzt angenommene Verbindung.</summary>
    public async Task SendAsync(string json)
    {
        WebSocket? ws;
        lock (_sockets)
        {
            ws = _sockets.Count > 0 ? _sockets[^1] : null;
        }
        if (ws == null)
        {
            return;
        }
        var bytes = System.Text.Encoding.UTF8.GetBytes(json);
        await ws.SendAsync(bytes, WebSocketMessageType.Text, true, _cts.Token);
    }

    public ValueTask DisposeAsync()
    {
        _cts.Cancel();
        try
        {
            _listener.Stop();
        }
        catch
        {
            // bereits gestoppt
        }
        lock (_sockets)
        {
            foreach (var ws in _sockets)
            {
                try
                {
                    ws.Dispose();
                }
                catch
                {
                    // bereits geschlossen
                }
            }
        }
        _listener.Close();
        _cts.Dispose();
        return ValueTask.CompletedTask;
    }
}
