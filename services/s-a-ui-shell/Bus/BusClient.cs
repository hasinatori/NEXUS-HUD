using System;
using System.Net.WebSockets;
using System.Text;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;

namespace NexusHud.UI.Bus;

// Zuletzt geändert: 2026-08-13
// WebSocket-Client zum Dev-Bus (Phase-1-Transport: Local WebSocket, Port 49152).
// Events laufen als JSON-Textframes (eine Nachricht = ein Frame).

public sealed class BusClient : IAsyncDisposable
{
    private readonly Uri _uri;
    private readonly ClientWebSocket _socket = new();
    private readonly CancellationTokenSource _cts = new();

    public event Action<string>? MessageReceived;
    public event Action? Connected;
    public event Action? Disconnected;

    public BusClient(Uri uri)
    {
        _uri = uri;
    }

    public async Task ConnectAsync()
    {
        await _socket.ConnectAsync(_uri, _cts.Token);
        Connected?.Invoke();
        await ReceiveLoopAsync();
    }

    public async Task SendHelloAsync(string source, string serviceId, string version, int protocolVersion)
    {
        string json = JsonSerializer.Serialize(Protocol.Hello(source, serviceId, version, protocolVersion));
        await SendAsync(json);
    }

    public async Task SendAsync(string json)
    {
        byte[] bytes = Encoding.UTF8.GetBytes(json);
        await _socket.SendAsync(bytes, WebSocketMessageType.Text, true, _cts.Token);
    }

    private async Task ReceiveLoopAsync()
    {
        var buffer = new byte[16 * 1024];
        try
        {
            while (_socket.State == WebSocketState.Open)
            {
                var result = await _socket.ReceiveAsync(buffer, _cts.Token);
                if (result.MessageType == WebSocketMessageType.Close)
                {
                    break;
                }

                string text = Encoding.UTF8.GetString(buffer, 0, result.Count);
                MessageReceived?.Invoke(text);
            }
        }
        catch (WebSocketException)
        {
            // Verbindung unterbrochen -> Disconnected unten
        }
        catch (OperationCanceledException)
        {
            // geplanter Shutdown
        }
        finally
        {
            Disconnected?.Invoke();
        }
    }

    public async ValueTask DisposeAsync()
    {
        try
        {
            await _socket.CloseAsync(WebSocketCloseStatus.NormalClosure, "bye", CancellationToken.None);
        }
        catch
        {
            // Socket bereits geschlossen
        }

        _socket.Dispose();
        _cts.Cancel();
        _cts.Dispose();
    }
}
