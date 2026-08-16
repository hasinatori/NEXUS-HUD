using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace NexusHud.UI.Widgets;

// Zuletzt geändert: 2026-08-16
// Konfigurierbares Dashboard-Grid (Phase-2-Waypoint): Layout = Zeilen/Spalten +
// Widget-Platzierungen (Kind, Titel, Position, Spans). Pure Datenklassen ohne
// WinUI-Abhängigkeiten -> headless testbar auf Linux (tests/s-a-ui-shell).
// JSON-Format (snake_case) dokumentiert in layout.json.

public enum WidgetKind
{
    Gauge,
    BuildBadge,
    MediaCard,
    GitWidget
}

public sealed class WidgetPlacement
{
    [JsonPropertyName("id")]
    public string Id { get; set; } = "";

    [JsonPropertyName("kind")]
    public WidgetKind Kind { get; set; }

    [JsonPropertyName("title")]
    public string Title { get; set; } = "";

    [JsonPropertyName("row")]
    public int Row { get; set; }

    [JsonPropertyName("column")]
    public int Column { get; set; }

    [JsonPropertyName("row_span")]
    public int RowSpan { get; set; } = 1;

    [JsonPropertyName("column_span")]
    public int ColumnSpan { get; set; } = 1;
}

public sealed class WidgetLayout
{
    public static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNameCaseInsensitive = true,
        Converters = { new JsonStringEnumConverter(JsonNamingPolicy.SnakeCaseLower) },
    };

    [JsonPropertyName("rows")]
    public int Rows { get; set; } = 4;

    [JsonPropertyName("columns")]
    public int Columns { get; set; } = 4;

    [JsonPropertyName("widgets")]
    public List<WidgetPlacement> Widgets { get; set; } = new();

    public static WidgetLayout Default() => new()
    {
        Rows = 4,
        Columns = 4,
        Widgets =
        {
            new WidgetPlacement { Id = "cpu-gauge", Kind = WidgetKind.Gauge, Title = "CPU", Row = 0, Column = 0 },
            new WidgetPlacement { Id = "ram-gauge", Kind = WidgetKind.Gauge, Title = "RAM", Row = 0, Column = 1 },
            new WidgetPlacement { Id = "git-status", Kind = WidgetKind.GitWidget, Title = "Git", Row = 1, Column = 0, ColumnSpan = 2 },
            new WidgetPlacement { Id = "media-card", Kind = WidgetKind.MediaCard, Title = "Media", Row = 2, Column = 0, ColumnSpan = 2 },
            new WidgetPlacement { Id = "build-badge", Kind = WidgetKind.BuildBadge, Title = "Build", Row = 3, Column = 0 },
        },
    };

    public static WidgetLayout FromJson(string json)
    {
        try
        {
            var layout = JsonSerializer.Deserialize<WidgetLayout>(json, JsonOptions)
                ?? throw new JsonException("WidgetLayout ist leer/null.");
            layout.Validate();
            return layout;
        }
        catch (JsonException ex)
        {
            throw new FormatException($"Ungültiges Widget-Layout: {ex.Message}", ex);
        }
        catch (InvalidOperationException ex)
        {
            throw new FormatException($"Ungültiges Widget-Layout: {ex.Message}", ex);
        }
    }

    public static WidgetLayout Load(string path) => FromJson(File.ReadAllText(path));

    /// <summary>Validierung der Grid-Geometrie: Rastermaße, eindeutige IDs,
    /// Platzierungen innerhalb des Rasters.</summary>
    public void Validate()
    {
        if (Rows < 1 || Columns < 1)
        {
            throw new InvalidOperationException($"Rastermaße ungültig (rows={Rows}, columns={Columns}).");
        }

        var seen = new HashSet<string>();
        foreach (var w in Widgets)
        {
            if (string.IsNullOrWhiteSpace(w.Id))
            {
                throw new InvalidOperationException("Widget ohne id.");
            }
            if (!seen.Add(w.Id))
            {
                throw new InvalidOperationException($"Widget-id doppelt: '{w.Id}'.");
            }
            if (w.Row < 0 || w.Column < 0)
            {
                throw new InvalidOperationException($"Widget '{w.Id}' liegt außerhalb des Rasters (row={w.Row}, column={w.Column}).");
            }
            if (w.RowSpan < 1 || w.ColumnSpan < 1)
            {
                throw new InvalidOperationException($"Widget '{w.Id}' hat ungültige Spans (row_span={w.RowSpan}, column_span={w.ColumnSpan}).");
            }
            if (w.Row + w.RowSpan > Rows || w.Column + w.ColumnSpan > Columns)
            {
                throw new InvalidOperationException($"Widget '{w.Id}' überschreitet das Raster ({Rows}x{Columns}).");
            }
        }
    }
}
