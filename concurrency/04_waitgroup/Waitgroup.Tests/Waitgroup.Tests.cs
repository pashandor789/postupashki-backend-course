using System.Threading;
using DefaultNamespace;
using Xunit;

public class WaitgroupTests
{
    [Fact]
    public async Task WaitOnZeroReturnsAtOnce()
    {
        var wg = new MyWaitgroup();

        var waiting = Task.Run(() => wg.Wait());

        var completed = await Task.WhenAny(
            waiting,
            Task.Delay(TimeSpan.FromSeconds(1))
        );

        Assert.Same(waiting, completed);
    }

    [Fact]
    public async Task WaitsForEveryone()
    {
        var wg = new MyWaitgroup();

        int done = 0;

        const int workers = 20;

        wg.Add(workers);

        for (int i = 0; i < workers; i++)
        {
            Task.Run(() =>
            {
                Thread.Sleep(20);

                Interlocked.Increment(ref done);

                wg.Done();
            });
        }

        var waiting = Task.Run(() => wg.Wait());

        var completed = await Task.WhenAny(
            waiting,
            Task.Delay(TimeSpan.FromSeconds(5))
        );

        Assert.Same(waiting, completed);

        Assert.Equal(workers, Volatile.Read(ref done));
    }

    [Fact]
    public async Task ManyWaiters()
    {
        var wg = new MyWaitgroup();

        wg.Add(1);

        const int waiters = 30;

        var tasks = Enumerable.Range(0, waiters)
            .Select(_ => Task.Run(() => wg.Wait()))
            .ToArray();

        await Task.Delay(50);

        wg.Done();

        var allTasks = Task.WhenAll(tasks);

        var completed = await Task.WhenAny(
            allTasks,
            Task.Delay(TimeSpan.FromSeconds(5))
        );

        Assert.Same(allTasks, completed);
    }

    [Fact]
    public async Task WaitBlocksUntilDone()
    {
        var wg = new MyWaitgroup();

        wg.Add(1);

        var waiting = Task.Run(() => wg.Wait());

        await Task.Delay(50);

        Assert.False(waiting.IsCompleted);

        wg.Done();

        var completed = await Task.WhenAny(
            waiting,
            Task.Delay(TimeSpan.FromSeconds(2))
        );

        Assert.Same(waiting, completed);
    }

    [Fact]
    public async Task Reuse()
    {
        var wg = new MyWaitgroup();

        for (int round = 0; round < 5; round++)
        {
            wg.Add(4);

            for (int i = 0; i < 4; i++)
            {
                Task.Run(() => wg.Done());
            }

            var waiting = Task.Run(() => wg.Wait());

            var completed = await Task.WhenAny(
                waiting,
                Task.Delay(TimeSpan.FromSeconds(2))
            );

            Assert.Same(waiting, completed);
        }
    }

    [Fact]
    public void NegativeCounterPanics()
    {
        var wg = new MyWaitgroup();

        Assert.Throws<Exception>(() => wg.Done());
    }

    [Fact]
    public async Task Stress()
    {
        for (int round = 0; round < 200; round++)
        {
            var wg = new MyWaitgroup();

            int counter = 0;

            wg.Add(8);

            for (int i = 0; i < 8; i++)
            {
                Task.Run(() =>
                {
                    Interlocked.Increment(ref counter);
                    wg.Done();
                });
            }

            var waiting = Task.Run(() => wg.Wait());

            var completed = await Task.WhenAny(
                waiting,
                Task.Delay(TimeSpan.FromSeconds(5))
            );

            Assert.Same(waiting, completed);

            Assert.Equal(
                8,
                Volatile.Read(ref counter)
            );
        }
    }
}