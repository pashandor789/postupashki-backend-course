using System.Threading;
using DefaultNamespace;
using Xunit;

public class OnceTests
{
    [Fact]
    public void RunsOnlyOnce()
    {
        var o = new MyOnce();
        int calls = 0;

        for (int i = 0; i < 10; i++)
        {
            o.Do(() => Interlocked.Increment(ref calls));
        }

        Assert.Equal(1, calls);
    }

    [Fact]
    public void Done()
    {
        var o = new MyOnce();

        Assert.False(o.Done());

        o.Do(() => { });

        Assert.True(o.Done());
    }

    [Fact]
    public async Task ConcurrentCallsRunItOnce()
    {
        var o = new MyOnce();
        int calls = 0;

        var tasks = Enumerable.Range(0, 100)
            .Select(_ => Task.Run(() =>
            {
                o.Do(() =>
                {
                    Interlocked.Increment(ref calls);
                    Thread.Sleep(10);
                });
            }))
            .ToArray();

        var allTasks = Task.WhenAll(tasks);

        var completed = await Task.WhenAny(
            allTasks,
            Task.Delay(TimeSpan.FromSeconds(10))
        );

        Assert.Same(allTasks, completed);
        Assert.Equal(1, Volatile.Read(ref calls));
    }

    [Fact]
    public async Task DoWaitsForTheWinner()
    {
        var o = new MyOnce();
        int ready = 0;

        var first = Task.Run(() =>
        {
            o.Do(() =>
            {
                Thread.Sleep(100);
                Volatile.Write(ref ready, 1);
            });
        });

        await Task.Delay(10);

        var second = Task.Run(() =>
        {
            o.Do(() => { });

            Assert.Equal(1, Volatile.Read(ref ready));
        });

        var firstCompleted = await Task.WhenAny(
            first,
            Task.Delay(TimeSpan.FromSeconds(5))
        );

        Assert.Same(first, firstCompleted);

        var secondCompleted = await Task.WhenAny(
            second,
            Task.Delay(TimeSpan.FromSeconds(5))
        );

        Assert.Same(second, secondCompleted);
    }

    [Fact]
    public void PanicCountsAsDone()
    {
        var o = new MyOnce();
        int calls = 0;

        try
        {
            o.Do(() =>
            {
                calls++;
                throw new Exception("упало");
            });
        }
        catch (Exception)
        {
        }

        o.Do(() =>
        {
            calls++;
        });

        Assert.Equal(1, calls);
    }

    [Fact]
    public async Task Stress()
    {
        for (int round = 0; round < 300; round++)
        {
            var o = new MyOnce();
            int calls = 0;

            var tasks = Enumerable.Range(0, 16)
                .Select(_ => Task.Run(() =>
                {
                    o.Do(() =>
                    {
                        Interlocked.Increment(ref calls);
                    });
                }))
                .ToArray();

            await Task.WhenAll(tasks);

            Assert.Equal(1, Volatile.Read(ref calls));
        }
    }
}