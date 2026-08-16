using System.Text.Json;
using NexusHud.UI.ViewModels;
using NexusHud.UI.Widgets;
using Xunit;

namespace NexusHud.UI.Tests;

// Zuletzt geändert: 2026-08-16
// WidgetLayout (JSON + Validierung) und DashboardViewModel (Grid + Status).

public sealed class WidgetLayoutTests
{
    private const string ValidJson = """
        {
          "rows": 2,
          "columns": 3,
          "widgets": [
            { "id": "cpu", "kind": "gauge", "title": "CPU", "row": 0, "column": 0, "column_span": 2 },
            { "id": "build", "kind": "build_badge", "title": "Build", "row": 1, "column": 2 }
          ]
        }
        """;

    [Fact]
    public void FromJson_Parses_SnakeCaseFelder_Und_EnumStrings()
    {
        var layout = WidgetLayout.FromJson(ValidJson);
        Assert.Equal(2, layout.Rows);
        Assert.Equal(3, layout.Columns);
        Assert.Equal(2, layout.Widgets.Count);

        var cpu = layout.Widgets[0];
        Assert.Equal("cpu", cpu.Id);
        Assert.Equal(WidgetKind.Gauge, cpu.Kind);
        Assert.Equal("CPU", cpu.Title);
        Assert.Equal(0, cpu.Row);
        Assert.Equal(0, cpu.Column);
        Assert.Equal(2, cpu.ColumnSpan);
    }

    [Fact]
    public void Default_Hat_Fuenf_Widgets_Und_Validiert()
    {
        var layout = WidgetLayout.Default();
        layout.Validate();
        Assert.Equal(4, layout.Rows);
        Assert.Equal(4, layout.Columns);
        Assert.Equal(5, layout.Widgets.Count);
    }

    [Fact]
    public void Ungueltiges_Layout_Wirft_FormatException()
    {
        var json = """{ "rows": 2, "columns": 2, "widgets": [ { "id": "x", "kind": "gauge", "row": 5, "column": 0 } ] }""";
        Assert.Throws<FormatException>(() => WidgetLayout.FromJson(json));
    }

    [Fact]
    public void Doppelte_WidgetIds_Werden_Abgelehnt()
    {
        var json = """
            { "rows": 2, "columns": 2, "widgets": [
                { "id": "a", "kind": "gauge", "row": 0, "column": 0 },
                { "id": "a", "kind": "gauge", "row": 1, "column": 1 } ] }
            """;
        var ex = Assert.Throws<FormatException>(() => WidgetLayout.FromJson(json));
        Assert.Contains("doppelt", ex.Message, StringComparison.OrdinalIgnoreCase);
    }

    [Fact]
    public void Unbekannte_WidgetKind_Zeichenkette_Wirft()
    {
        var json = """{ "rows": 2, "columns": 2, "widgets": [ { "id": "x", "kind": "hologramm", "row": 0, "column": 0 } ] }""";
        Assert.Throws<FormatException>(() => WidgetLayout.FromJson(json));
    }

    [Fact]
    public void Default_RoundTrip_Über_JSON()
    {
        var json = JsonSerializer.Serialize(WidgetLayout.Default(), WidgetLayout.JsonOptions);
        var layout = WidgetLayout.FromJson(json);

        Assert.Equal(WidgetLayout.Default().Widgets.Count, layout.Widgets.Count);
        Assert.Equal(WidgetKind.GitWidget, layout.Widgets.First(w => w.Id == "git-status").Kind);
        Assert.Equal(WidgetKind.MediaCard, layout.Widgets.First(w => w.Id == "media-card").Kind);
        Assert.Equal(WidgetKind.BuildBadge, layout.Widgets.First(w => w.Id == "build-badge").Kind);
    }
}

public sealed class DashboardViewModelTests
{
    [Fact]
    public void ApplyLayout_Erzeugt_Sortierte_Widgets()
    {
        var vm = new DashboardViewModel();
        vm.ApplyLayout(WidgetLayout.Default());

        Assert.Equal(4, vm.Rows);
        Assert.Equal(4, vm.Columns);
        Assert.Equal(5, vm.Widgets.Count);
        Assert.Equal("cpu-gauge", vm.Widgets[0].Id);
        Assert.Equal("Gauge", vm.Widgets[0].KindLabel);
    }

    [Fact]
    public void ApplyJson_Und_Find()
    {
        var vm = new DashboardViewModel();
        vm.ApplyJson("""{ "rows": 2, "columns": 2, "widgets": [ { "id": "cpu", "kind": "gauge", "title": "CPU", "row": 0, "column": 0 } ] }""");

        var cpu = vm.Find("cpu");
        Assert.NotNull(cpu);
        Assert.Equal("CPU", cpu!.Title);
        Assert.Null(vm.Find("gibt-es-nicht"));
    }

    [Fact]
    public void SetStatus_Aktualisiert_Widget_Und_Meldet_PropertyChanged()
    {
        var vm = new DashboardViewModel();
        vm.ApplyLayout(WidgetLayout.Default());

        var cpu = vm.Find("cpu-gauge")!;
        var changed = false;
        cpu.PropertyChanged += (_, e) => changed |= e.PropertyName == nameof(cpu.Status);

        Assert.True(vm.SetStatus("cpu-gauge", "42 %"));
        Assert.Equal("42 %", cpu.Status);
        Assert.True(changed);

        Assert.False(vm.SetStatus("unbekannt", "egal"));
    }
}
