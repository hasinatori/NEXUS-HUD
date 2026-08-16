using System.Collections.ObjectModel;
using System.ComponentModel;
using System.Linq;
using System.Runtime.CompilerServices;
using NexusHud.UI.Widgets;

namespace NexusHud.UI.ViewModels;

// Zuletzt geändert: 2026-08-16
// Dashboard-Grid (Phase-2-Waypoint): hält Rastermaße und die Widget-Zeilen,
// lädt ein WidgetLayout (JSON) und aktualisiert Widget-Status. Pure Darstellung
// ohne WinUI-Abhängigkeit -> headless testbar.

public sealed class DashboardViewModel : INotifyPropertyChanged
{
    private int _rows;
    private int _columns;

    public event PropertyChangedEventHandler? PropertyChanged;

    public int Rows
    {
        get => _rows;
        private set
        {
            if (_rows != value)
            {
                _rows = value;
                OnPropertyChanged();
            }
        }
    }

    public int Columns
    {
        get => _columns;
        private set
        {
            if (_columns != value)
            {
                _columns = value;
                OnPropertyChanged();
            }
        }
    }

    public ObservableCollection<WidgetViewModel> Widgets { get; } = new();

    public void ApplyLayout(WidgetLayout layout)
    {
        layout.Validate();
        Rows = layout.Rows;
        Columns = layout.Columns;
        Widgets.Clear();
        foreach (var placement in layout.Widgets.OrderBy(w => w.Row).ThenBy(w => w.Column))
        {
            Widgets.Add(new WidgetViewModel(placement));
        }
    }

    public void ApplyJson(string json) => ApplyLayout(WidgetLayout.FromJson(json));

    public void Load(string path) => ApplyLayout(WidgetLayout.Load(path));

    public WidgetViewModel? Find(string id) => Widgets.FirstOrDefault(w => w.Id == id);

    /// <summary>Setzt den Status eines Widgets (ohne Fehler, wenn die ID fehlt —
    /// Events von Diensten, die nicht konfiguriert sind, dürfen nicht crashen).</summary>
    public bool SetStatus(string id, string status)
    {
        var widget = Find(id);
        if (widget == null)
        {
            return false;
        }
        widget.Status = status;
        return true;
    }

    private void OnPropertyChanged([CallerMemberName] string? name = null)
        => PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(name));
}
