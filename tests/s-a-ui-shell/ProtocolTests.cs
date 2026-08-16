using System.Text.Json;
using NexusHud.UI.Bus;
using Xunit;

namespace NexusHud.UI.Tests;

// Zuletzt geändert: 2026-08-16
// Protocol: Hello/Heartbeat-JSON entspricht exakt dem Schema (Feldnamen + Werte).

public sealed class ProtocolTests
{
    [Fact]
    public void Hello_Enthaelt_SchemaKonforme_Felder()
    {
        var hello = Protocol.Hello("S-A", "s-a-ui-shell", "0.2.0-s-a.1", 1);
        var json = JsonSerializer.Serialize(hello);
        using var doc = JsonDocument.Parse(json);

        Assert.Equal("2.0", doc.RootElement.GetProperty("jsonrpc").GetString());
        Assert.Equal("event.system.hello", doc.RootElement.GetProperty("method").GetString());

        var p = doc.RootElement.GetProperty("params");
        Assert.Equal("S-A", p.GetProperty("source").GetString());
        Assert.Equal(1, p.GetProperty("protocol_version").GetInt32());
        Assert.Equal("s-a-ui-shell", p.GetProperty("service_id").GetString());
        Assert.Equal("0.2.0-s-a.1", p.GetProperty("version").GetString());
        Assert.Equal(20, p.GetProperty("ts").GetString()!.Length); // RFC3339 yyyy-MM-ddTHH:mm:ssZ
    }

    [Fact]
    public void Heartbeat_Enthaelt_SchemaKonforme_Felder()
    {
        var heartbeat = Protocol.Heartbeat("S-A", "s-a-ui-shell", 1);
        var json = JsonSerializer.Serialize(heartbeat);
        using var doc = JsonDocument.Parse(json);

        Assert.Equal("2.0", doc.RootElement.GetProperty("jsonrpc").GetString());
        Assert.Equal("event.system.heartbeat", doc.RootElement.GetProperty("method").GetString());

        var p = doc.RootElement.GetProperty("params");
        Assert.Equal("S-A", p.GetProperty("source").GetString());
        Assert.Equal(1, p.GetProperty("protocol_version").GetInt32());
        Assert.Equal("s-a-ui-shell", p.GetProperty("service_id").GetString());
        Assert.Equal(20, p.GetProperty("ts").GetString()!.Length);
    }

    [Fact]
    public void Protocol_Konstanten()
    {
        Assert.Equal(1, Protocol.ProtocolVersion);
        Assert.Equal("S-A", Protocol.Source);
        Assert.Equal("s-a-ui-shell", Protocol.ServiceId);
    }
}
