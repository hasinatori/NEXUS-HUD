using System.ComponentModel;
using NexusHud.UI.ViewModels;
using Xunit;

namespace NexusHud.UI.Tests;

// Zuletzt geändert: 2026-08-14
// MainViewModel: PropertyChanged + Zähler/Text-Ableitungen.

public sealed class MainViewModelTests
{
    [Fact]
    public void Initialzustand()
    {
        var vm = new MainViewModel();
        Assert.Equal("Bus: getrennt", vm.State);
        Assert.Equal(0, vm.HelloCount);
        Assert.Equal("Hellos: 0", vm.HelloCountString);
        Assert.Equal("Warte auf Events vom Bus…", vm.LastEvent);
    }

    [Fact]
    public void PropertyChanged_Feuert_Bei_Zustand()
    {
        var vm = new MainViewModel();
        var changes = new List<string?>();
        vm.PropertyChanged += (_, e) => changes.Add(e.PropertyName);

        vm.State = "Bus: verbunden";
        Assert.Equal("Bus: verbunden", vm.State);
        Assert.Contains(nameof(vm.State), changes);
    }

    [Fact]
    public void HelloCount_Aktualisiert_TextAbleitung()
    {
        var vm = new MainViewModel();
        var changes = new List<string?>();
        vm.PropertyChanged += (_, e) => changes.Add(e.PropertyName);

        vm.HelloCount = 3;
        Assert.Equal(3, vm.HelloCount);
        Assert.Equal("Hellos: 3", vm.HelloCountString);
        Assert.Contains(nameof(vm.HelloCountString), changes);
    }

    [Fact]
    public void LastEvent_SetUndErkannt()
    {
        var vm = new MainViewModel();
        vm.LastEvent = "{\"method\":\"event.system.hello\"}";
        Assert.Equal("{\"method\":\"event.system.hello\"}", vm.LastEvent);
    }
}
