using System.Threading;
using Xunit;
using DefaultNamespace;

public class BarrierTests
{
    private static Task Finished(Action action)
    {
        return Task.Run(action);
    }

    [Fact]
    public async Task NobodyPassesEarly()
    {
        var b = new MyBarrier(3);
        int passed = 0;

        for (int i = 0; i < 2; i++)
        {
            _ = Task.Run(() =>
            {
                b.Wait();
                Interlocked.Increment(ref passed);
            });
        }

        await Task.Delay(100);

        Assert.Equal(0, Volatile.Read(ref passed));

        var last = Finished(() =>
        {
            b.Wait();
            Interlocked.Increment(ref passed);
        });

        var completed = await Task.WhenAny(last, Task.Delay(TimeSpan.FromSeconds(2)));
        Assert.Same(last, completed);

        var deadline = DateTime.UtcNow.AddSeconds(2);

        while (Volatile.Read(ref passed) != 3 && DateTime.UtcNow < deadline)
            await Task.Delay(1);

        Assert.Equal(3, Volatile.Read(ref passed));
    }

    [Fact]
    public async Task SingleParticipant()
    {
        var b = new MyBarrier(1);

        var waiting = Finished(b.Wait);

        var completed = await Task.WhenAny(waiting, Task.Delay(TimeSpan.FromSeconds(1)));

        Assert.Same(waiting, completed);
    }

    [Fact]
    public async Task ReusableAcrossRounds()
    {
        const int parties = 6;
        const int rounds = 50;

        var b = new MyBarrier(parties);
        int round = 0;

        var tasks = Enumerable.Range(0, parties).Select(_ => Task.Run(() =>
        {
            for (int r = 0; r < rounds; r++)
            {
                b.Wait();
                Interlocked.Increment(ref round);
            }
        })).ToArray();

        var allTasks = Task.WhenAll(tasks);
        var completed = await Task.WhenAny(allTasks, Task.Delay(TimeSpan.FromSeconds(30)));

        Assert.Same(allTasks, completed);
        Assert.Equal(parties * rounds, Volatile.Read(ref round));
    }

    [Fact]
    public async Task RoundsDoNotOverlap()
    {
        const int parties = 4;
        const int rounds = 100;

        var b = new MyBarrier(parties);
        int inRound = 0;
        int bad = 0;

        var tasks = Enumerable.Range(0, parties).Select(_ => Task.Run(() =>
        {
            for (int r = 0; r < rounds; r++)
            {
                b.Wait();

                if (Interlocked.Increment(ref inRound) > parties)
                    Interlocked.Exchange(ref bad, 1);

                Thread.Sleep(1);

                Interlocked.Decrement(ref inRound);
            }
        })).ToArray();

        var allTasks = Task.WhenAll(tasks);
        var completed = await Task.WhenAny(allTasks, Task.Delay(TimeSpan.FromSeconds(30)));

        Assert.Same(allTasks, completed);
        Assert.Equal(0, Volatile.Read(ref bad));
    }
}