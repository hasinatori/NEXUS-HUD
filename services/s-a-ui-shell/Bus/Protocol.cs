using System;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace NexusHud.UI.Bus;

// Zuletzt geändert: 2026-08-16
// JSON-RPC-2.0-Hello/-Heartbeat für den Dev-Bus. Feldnamen folgen exakt dem Schema
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

    public static HeartbeatMessage Heartbeat(string source, string serviceId, int protocolVersion)
    {
        return new HeartbeatMessage(
            "2.0",
            "event.system.heartbeat",
            new HeartbeatParams(
                source,
                protocolVersion,
                serviceId,
                DateTime.UtcNow.ToString("yyyy-MM-dd'T'HH:mm:ss'Z'")));
    }

    public static string CmdMediaToggle()
    {
        return JsonSerializer.Serialize(new
        {
            jsonrpc = "2.0",
            method = "cmd.media.toggle",
            @params = new { source = Source, protocol_version = ProtocolVersion, ts = DateTime.UtcNow.ToString("yyyy-MM-dd'T'HH:mm:ss'Z'") }
        });
    }

    public static string CmdMediaNext()
    {
        return JsonSerializer.Serialize(new
        {
            jsonrpc = "2.0",
            method = "cmd.media.next",
            @params = new { source = Source, protocol_version = ProtocolVersion, ts = DateTime.UtcNow.ToString("yyyy-MM-dd'T'HH:mm:ss'Z'") }
        });
    }

    public static string CmdMediaVolume(int volume)
    {
        return JsonSerializer.Serialize(new
        {
            jsonrpc = "2.0",
            method = "cmd.media.volume",
            @params = new { source = Source, protocol_version = ProtocolVersion, volume, ts = DateTime.UtcNow.ToString("yyyy-MM-dd'T'HH:mm:ss'Z'") }
        });
    }

    public static string CmdAppLaunch(string path, string[]? args = null, bool focus = true)
    {
        return JsonSerializer.Serialize(new
        {
            jsonrpc = "2.0",
            method = "cmd.app.launch",
            @params = new { source = Source, protocol_version = ProtocolVersion, path, args = args ?? Array.Empty<string>(), focus, ts = DateTime.UtcNow.ToString("yyyy-MM-dd'T'HH:mm:ss'Z'") }
        });
    }

    public static string CmdHotkeyRegister(string hotkeyId, string[] modifiers, string key)
    {
        return JsonSerializer.Serialize(new
        {
            jsonrpc = "2.0",
            method = "cmd.hotkey.register",
            @params = new { source = Source, protocol_version = ProtocolVersion, hotkey_id = hotkeyId, modifiers, key, ts = DateTime.UtcNow.ToString("yyyy-MM-dd'T'HH:mm:ss'Z'") }
        });
    }

    public static string CmdWindowMove(string windowTitle, int x, int y, int width, int height)
    {
        return JsonSerializer.Serialize(new
        {
            jsonrpc = "2.0",
            method = "cmd.window.move",
            @params = new { source = Source, protocol_version = ProtocolVersion, window_title = windowTitle, x, y, width, height, ts = DateTime.UtcNow.ToString("yyyy-MM-dd'T'HH:mm:ss'Z'") }
        });
    }

    public static string CmdAutomationRun(string task)
    {
        return JsonSerializer.Serialize(new
        {
            jsonrpc = "2.0",
            method = "cmd.automation.run",
            @params = new { source = Source, protocol_version = ProtocolVersion, task, ts = DateTime.UtcNow.ToString("yyyy-MM-dd'T'HH:mm:ss'Z'") }
        });
    }

    public static string CmdClipboardSet(string content)
    {
        return JsonSerializer.Serialize(new
        {
            jsonrpc = "2.0",
            method = "cmd.clipboard.set",
            @params = new { source = Source, protocol_version = ProtocolVersion, content, ts = DateTime.UtcNow.ToString("yyyy-MM-dd'T'HH:mm:ss'Z'") }
        });
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

public sealed record HeartbeatMessage(
    [property: JsonPropertyName("jsonrpc")] string JsonRpc,
    [property: JsonPropertyName("method")] string Method,
    [property: JsonPropertyName("params")] HeartbeatParams Params);

public sealed record HeartbeatParams(
    [property: JsonPropertyName("source")] string Source,
    [property: JsonPropertyName("protocol_version")] int ProtocolVersion,
    [property: JsonPropertyName("service_id")] string ServiceId,
    [property: JsonPropertyName("ts")] string Ts);
