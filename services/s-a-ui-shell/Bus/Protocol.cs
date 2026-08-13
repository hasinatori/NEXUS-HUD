using System;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace NexusHud.UI.Bus;

// Zuletzt geändert: 2026-08-13
// JSON-RPC-2.0-Hello für den Dev-Bus. Feldnamen folgen exakt dem Schema
// (schema/events.schema.json): jsonrpc, method, params + camel/snake wie vom Bus erwartet.

public static class Protocol
{
    public const int ProtocolVersion = 1;
    public const string Source = "S-A";
    public const string ServiceId = "s-a-ui-shell";

    public static HelloMessage Hello(string source, string serviceId, string version, int protocolVersion)
    {
        return new HelloMessage(
            "2.0",
            "event.system.hello",
            new HelloParams(
                source,
                protocolVersion,
                serviceId,
                version,
                DateTime.UtcNow.ToString("yyyy-MM-dd'T'HH:mm:ss'Z'")));
    }
}

public sealed record HelloMessage(
    [property: JsonPropertyName("jsonrpc")] string JsonRpc,
    [property: JsonPropertyName("method")] string Method,
    [property: JsonPropertyName("params")] HelloParams Params);

public sealed record HelloParams(
    [property: JsonPropertyName("source")] string Source,
    [property: JsonPropertyName("protocol_version")] int ProtocolVersion,
    [property: JsonPropertyName("service_id")] string ServiceId,
    [property: JsonPropertyName("version")] string Version,
    [property: JsonPropertyName("ts")] string Ts);
