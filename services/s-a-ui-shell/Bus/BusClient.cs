using System;
using System.Net.WebSockets;
using System.Text;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;

namespace NexusHud.UI.Bus;

// Zuletzt geändert: 2026-08-14
// WebSocket-Client zum Dev-Bus (Phase-1-Transport: Local WebSocket, Port 49152).
// Events laufen als JSON-Textframes (eine Nachricht = ein Frame).
// RunAsync verbindet, ruft onConnectedAsync auf (z. B. Hello senden) und
// verbindet nach einem Verbindungsabbruch automatisch neu (reconnectDelay).

public sealed class BusClient : IAsyncDisposable
{
    private readonly Uri _uri;
    private readonly TimeSpan _reconnectDelay;
    private readonly CancellationTokenSource _cts = new();
    private readonly SemaphoreSlim _sendLock = new(1, 1);

    private ClientWebSocket? _socket;
    private Task? _runTask;

    public event Action<string>? MessageReceived;
    public event Action? Connected;
    public event Action? Disconnected;

    public BusClient(Uri uri, TimeSpan? reconnectDelay = null)
    {
        _uri = uri;
        _reconnectDelay = reconnectDelay ?? TimeSpan.FromSeconds(2);
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
            try
            {
                _socket = socket;
                await socket.ConnectAsync(_uri, ct);
                Connected?.Invoke();
                await onConnectedAsync();
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

            string text = Encoding.UTF8.GetString(buffer, 0, result.Count);
            MessageReceived?.Invoke(text);
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
