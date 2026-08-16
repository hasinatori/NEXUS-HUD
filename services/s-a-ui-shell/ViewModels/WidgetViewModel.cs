using System.ComponentModel;
using System.Runtime.CompilerServices;
using NexusHud.UI.Widgets;

namespace NexusHud.UI.ViewModels;

// Zuletzt geändert: 2026-08-16
// Eine Widget-Zeile im Dashboard-Grid für x:Bind: feste Platzierung (aus dem
// Layout) + veränderlicher Status. Pure Darstellung ohne WinUI-Abhängigkeit.

public sealed class WidgetViewModel : INotifyPropertyChanged
{
    private string _status = "";

    public WidgetViewModel(WidgetPlacement placement)
    {
        Id = placement.Id;
        Title = placement.Title;
        Kind = placement.Kind;
        Row = placement.Row;
        Column = placement.Column;
        RowSpan = placement.RowSpan;
        ColumnSpan = placement.ColumnSpan;
    }

    public event PropertyChangedEventHandler? PropertyChanged;

    public string Id { get; }
    public string Title { get; }
    public WidgetKind Kind { get; }
    public int Row { get; }
    public int Column { get; }
    public int RowSpan { get; }
    public int ColumnSpan { get; }

    public string KindLabel => Kind switch
    {
        WidgetKind.Gauge => "Gauge",
        WidgetKind.BuildBadge => "Build",
        WidgetKind.MediaCard => "Media",
        WidgetKind.GitWidget => "Git",
        _ => Kind.ToString(),
    };

    public string Status
    {
        get => _status;
        set
        {
            if (_status != value)
            {
                _status = value;
                OnPropertyChanged();
            }
        }
    }

    private void OnPropertyChanged([CallerMemberName] string? name = null)
        => PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(name));
}
