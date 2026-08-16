using System;
using System.Net.WebSockets;
using System.Text;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;

namespace NexusHud.UI.Bus;

// Zuletzt geändert: 2026-08-16
// WebSocket-Client zum Dev-Bus (Phase-1-Transport: Local WebSocket, Port 49152).
// Events laufen als JSON-Textframes (eine Nachricht = ein Frame).
// RunAsync verbindet, ruft onConnectedAsync auf (z. B. Hello senden) und
// verbindet nach einem Verbindungsabbruch automatisch neu (reconnectDelay).
//
// Keepalive: Bei aktivem keepAliveInterval sendet der Client periodisch
// event.system.heartbeat an den Bus. Der Watchdog (keepAliveTimeout) erkennt
// harte Netzabriffe ohne Close-Frame: Bleibt ein Inbound-Frame länger als
// keepAliveTimeout aus, wird die Verbindung per Abort beendet und der
// Reconnect-Loop greift. Beide Mechanismen sind per Konstruktor abschaltbar.

public sealed class BusClient : IAsyncDisposable
{
    private readonly Uri _uri;
    private readonly TimeSpan _reconnectDelay;
    private readonly TimeSpan _keepAliveInterval;
    private readonly TimeSpan _keepAliveTimeout;
    private readonly CancellationTokenSource _cts = new();
    private readonly SemaphoreSlim _sendLock = new(1, 1);

    private ClientWebSocket? _socket;
    private Task? _runTask;
    private DateTime _lastReceiveUtc = DateTime.MinValue;

    public event Action<string>? MessageReceived;
    public event Action? Connected;
    public event Action? Disconnected;

    public BusClient(Uri uri, TimeSpan? reconnectDelay = null, TimeSpan? keepAliveInterval = null, TimeSpan? keepAliveTimeout = null)
    {
        _uri = uri;
        _reconnectDelay = reconnectDelay ?? TimeSpan.FromSeconds(2);
        _keepAliveInterval = keepAliveInterval ?? TimeSpan.Zero;
        _keepAliveTimeout = keepAliveTimeout ?? TimeSpan.Zero;
    }

    /// <summary>Verbindungs-Loop: verbindet, meldet Connected, ruft onConnectedAsync
    /// auf (z. B. Hello senden), empfängt, verbindet bei Abbruch erneut. Läuft bis
    /// StopAsync/DisposeAsync (interne CancellationSource).</summary>
    public Task RunAsync(Func<Task> onConnectedAsync)
    {
        _runTask = RunLoopAsync(onConnectedAsync, _cts.Token);
        return _runTask;
    }

    private async Task RunLoopAsync(Func<Task> onConnectedAsync, CancellationToken ct)
    {
        while (!ct.IsCancellationRequested)
        {
            var socket = new ClientWebSocket();
            using var connectionCts = CancellationTokenSource.CreateLinkedTokenSource(ct);
            var connectionToken = connectionCts.Token;
            Task? keepAliveTask = null;
            try
            {
                _socket = socket;
                await socket.ConnectAsync(_uri, ct);
                _lastReceiveUtc = DateTime.UtcNow;
                Connected?.Invoke();
                await onConnectedAsync();
                if (_keepAliveInterval > TimeSpan.Zero || _keepAliveTimeout > TimeSpan.Zero)
                {
                    keepAliveTask = RunKeepAliveLoopAsync(socket, connectionToken);
                }
                await ReceiveLoopAsync(socket, ct);
            }
            catch (OperationCanceledException)
            {
                // Abbruch durch StopAsync/ct
            }
            catch (Exception)
            {
                // Verbindungsfehler -> Reconnect-Versuch unten
            }
            finally
            {
                connectionCts.Cancel();
                if (keepAliveTask != null)
                {
                    try
                    {
                        await keepAliveTask;
                    }
                    catch (OperationCanceledException)
                    {
                        // erwarteter Abbruch
                    }
                }
                if (_socket == socket)
                {
                    _socket = null;
                }
                try
                {
                    socket.Dispose();
                }
                catch
                {
                    // bereits geschlossen
                }
                Disconnected?.Invoke();

                if (!ct.IsCancellationRequested)
                {
                    try
                    {
                        await Task.Delay(_reconnectDelay, ct);
                    }
                    catch (OperationCanceledException)
                    {
                        // Abbruch -> Schleifenbedingung beendet die Schleife
                    }
                }
            }
        }
    }

    public async Task SendHelloAsync(string source, string serviceId, string version, int protocolVersion, CancellationToken ct = default)
    {
        string json = JsonSerializer.Serialize(Protocol.Hello(source, serviceId, version, protocolVersion));
        await SendAsync(json, ct);
    }

    public async Task SendHeartbeatAsync(string source, string serviceId, int protocolVersion, CancellationToken ct = default)
    {
        string json = JsonSerializer.Serialize(Protocol.Heartbeat(source, serviceId, protocolVersion));
        await SendAsync(json, ct);
    }

    public async Task SendAsync(string json, CancellationToken ct = default)
    {
        var socket = _socket
            ?? throw new InvalidOperationException("BusClient ist nicht verbunden.");
        if (socket.State != WebSocketState.Open)
        {
            throw new InvalidOperationException($"BusClient-Verbindung ist nicht offen (State: {socket.State}).");
        }

        byte[] bytes = Encoding.UTF8.GetBytes(json);
        await _sendLock.WaitAsync(ct);
        try
        {
            await socket.SendAsync(bytes, WebSocketMessageType.Text, true, ct);
        }
        finally
        {
            _sendLock.Release();
        }
    }

    private async Task ReceiveLoopAsync(ClientWebSocket socket, CancellationToken ct)
    {
        var buffer = new byte[16 * 1024];
        while (socket.State == WebSocketState.Open)
        {
            var result = await socket.ReceiveAsync(buffer, ct);
            if (result.MessageType == WebSocketMessageType.Close)
            {
                break;
            }

            _lastReceiveUtc = DateTime.UtcNow;
            string text = Encoding.UTF8.GetString(buffer, 0, result.Count);
            MessageReceived?.Invoke(text);
        }
    }

    /// <summary>Keepalive-Loop (läuft pro Verbindung): sendet periodisch
    /// event.system.heartbeat und bricht die Verbindung per Abort ab, wenn
    /// länger als keepAliveTimeout kein Inbound-Frame eingetroffen ist
    /// (hartes Netzwerkproblem ohne Close-Frame).</summary>
    private async Task RunKeepAliveLoopAsync(ClientWebSocket socket, CancellationToken ct)
    {
        var tick = TimeSpan.FromSeconds(1);
        if (_keepAliveInterval > TimeSpan.Zero && _keepAliveInterval < tick)
        {
            tick = _keepAliveInterval;
        }
        if (_keepAliveTimeout > TimeSpan.Zero && _keepAliveTimeout / 4 < tick)
        {
            tick = _keepAliveTimeout / 4;
        }

        var lastSendUtc = DateTime.UtcNow;
        while (!ct.IsCancellationRequested)
        {
            try
            {
                await Task.Delay(tick, ct);
            }
            catch (OperationCanceledException)
            {
                return;
            }

            if (_keepAliveTimeout > TimeSpan.Zero
                && _lastReceiveUtc != DateTime.MinValue
                && DateTime.UtcNow - _lastReceiveUtc > _keepAliveTimeout)
            {
                socket.Abort();
                return;
            }

            if (_keepAliveInterval > TimeSpan.Zero && DateTime.UtcNow - lastSendUtc >= _keepAliveInterval)
            {
                lastSendUtc = DateTime.UtcNow;
                await TrySendHeartbeatAsync(ct);
            }
        }
    }

    private async Task TrySendHeartbeatAsync(CancellationToken ct)
    {
        try
        {
            await SendHeartbeatAsync(Protocol.Source, Protocol.ServiceId, Protocol.ProtocolVersion, ct);
        }
        catch (Exception)
        {
            // Verbindung ist bereits defekt -> ReceiveLoop/Reconnect übernimmt.
        }
    }

    /// <summary>Beendet den Verbindungs-Loop und räumt auf (idempotent).</summary>
    public async Task StopAsync()
    {
        _cts.Cancel();
        var runTask = _runTask;
        if (runTask != null)
        {
            try
            {
                await runTask;
            }
            catch (OperationCanceledException)
            {
                // erwarteter Abbruch
            }
        }
        try
        {
            _socket?.Dispose();
        }
        catch
        {
            // bereits geschlossen
        }
        _socket = null;
    }

    public async ValueTask DisposeAsync()
    {
        await StopAsync();
        _sendLock.Dispose();
        _cts.Dispose();
    }
}
