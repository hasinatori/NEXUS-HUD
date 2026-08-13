using System.ComponentModel;
using System.Runtime.CompilerServices;

namespace NexusHud.UI.ViewModels;

// Zuletzt geändert: 2026-08-13
public sealed class MainViewModel : INotifyPropertyChanged
{
    private string _state = "Bus: getrennt";
    private int _helloCount;
    private string _lastEvent = "Warte auf Events vom Bus…";

    public event PropertyChangedEventHandler? PropertyChanged;

    public string State
    {
        get => _state;
        set
        {
            if (_state != value)
            {
                _state = value;
                OnPropertyChanged();
            }
        }
    }

    public int HelloCount
    {
        get => _helloCount;
        set
        {
            if (_helloCount != value)
            {
                _helloCount = value;
                OnPropertyChanged(nameof(HelloCountString));
            }
        }
    }

    public string HelloCountString => $"Hellos: {HelloCount}";

    public string LastEvent
    {
        get => _lastEvent;
        set
        {
            if (_lastEvent != value)
            {
                _lastEvent = value;
                OnPropertyChanged();
            }
        }
    }

    private void OnPropertyChanged([CallerMemberName] string? name = null)
        => PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(name));
}
