using DefaultNamespace;

namespace a;

using System.Collections.Concurrent;
using System.Threading;
using Xunit;

public class SemaphoreTests
{
    [Fact]
    public void AcquireRelease()
    {
        var s = new MySemaphore(2);

        Assert.Equal(2, s.Available());

        s.Acquire();
        s.Acquire();

        Assert.Equal(0, s.Available());

        s.Release();
        s.Release();

        Assert.Equal(2, s.Available());
    }

    [Fact]
    public async Task BlocksWhenEmpty()
    {
        var s = new MySemaphore(1);

        s.Acquire();

        var second = Task.Run(() =>
        {
            s.Acquire();
            s.Release();
        });

        await Task.Delay(50);

        Assert.False(second.IsCompleted);

        s.Release();

        var completed = await Task.WhenAny(
            second,
            Task.Delay(TimeSpan.FromSeconds(2))
        );

        Assert.Same(second, completed);
    }

    [Fact]
    public async Task NeverMoreThanLimit()
    {
        const int limit = 3;
        var s = new MySemaphore(limit);

        int inside = 0;
        int peak = 0;

        var tasks = Enumerable.Range(0, 40)
            .Select(_ => Task.Run(() =>
            {
                s.Acquire();

                int now = Interlocked.Increment(ref inside);

                while (true)
                {
                    int old = Volatile.Read(ref peak);

                    if (now <= old ||
                        Interlocked.CompareExchange(
                            ref peak,
                            now,
                            old) == old)
                    {
                        break;
                    }
                }

                Thread.Sleep(1);

                Interlocked.Decrement(ref inside);

                s.Release();
            }))
            .ToArray();

        var completed = await Task.WhenAny(
            Task.WhenAll(tasks),
            Task.Delay(TimeSpan.FromSeconds(10))
        );

        Assert.True(completed == tasks[0] || tasks.All(t => t.IsCompleted));

        Assert.True(peak <= limit,
            $"одновременно внутри было {peak}, лимит {limit}");

        Assert.True(peak >= 2,
            $"параллелизма не случилось вовсе: пик {peak}");
    }

    [Fact]
    public void TryAcquire()
    {
        var s = new MySemaphore(1);

        Assert.True(s.TryAcquire());

        Assert.False(s.TryAcquire());

        s.Release();

        Assert.True(s.TryAcquire());
    }

    [Fact]
    public async Task ZeroPermits()
    {
        var s = new MySemaphore(0);

        Assert.False(s.TryAcquire());

        var waiting = Task.Run(() =>
        {
            s.Acquire();
        });

        await Task.Delay(50);

        Assert.False(waiting.IsCompleted);

        s.Release();

        var completed = await Task.WhenAny(
            waiting,
            Task.Delay(TimeSpan.FromSeconds(2))
        );

        Assert.Same(waiting, completed);
    }

    [Fact]
    public async Task AllWaitersWakeUp()
    {
        var s = new MySemaphore(0);

        const int waiters = 30;

        var tasks = Enumerable.Range(0, waiters)
            .Select(_ => Task.Run(() =>
            {
                s.Acquire();
            }))
            .ToArray();

        await Task.Delay(50);

        for (int i = 0; i < waiters; i++)
        {
            s.Release();
        }

        var completed = await Task.WhenAny(
            Task.WhenAll(tasks),
            Task.Delay(TimeSpan.FromSeconds(10))
        );

        Assert.True(tasks.All(t => t.IsCompleted),
            "не все ждущие проснулись");
    }
}