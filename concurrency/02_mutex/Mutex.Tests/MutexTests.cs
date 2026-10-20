using DefaultNamespace;
using Xunit;

public class MutexTests
{
    [Fact]
    public void LockUnlock()
    {
        var m = new MyMutex();

        m.Lock();
        m.Unlock();

        m.Lock();
        m.Unlock();
    }

    [Fact]
    public async Task MutualExclusion()
    {
        const int threads = 16;
        const int iterations = 20_000;

        var m = new MyMutex();
        int counter = 0;

        var tasks = Enumerable.Range(0, threads)
            .Select(_ => Task.Run(() =>
            {
                for (int j = 0; j < iterations; j++)
                {
                    m.Lock();
                    counter++;
                    m.Unlock();
                }
            }))
            .ToArray();

        await Task.WhenAll(tasks);

        Assert.Equal(threads * iterations, counter);
    }

    [Fact]
    public async Task OnlyOneInside()
    {
        var m = new MyMutex();

        int inside = 0;
        bool bad = false;

        var tasks = Enumerable.Range(0, 16)
            .Select(_ => Task.Run(() =>
            {
                for (int j = 0; j < 3000; j++)
                {
                    m.Lock();

                    inside++;

                    if (inside != 1)
                        bad = true;

                    inside--;

                    m.Unlock();
                }
            }))
            .ToArray();

        await Task.WhenAll(tasks);

        Assert.False(bad);
    }

    [Fact]
    public async Task LockWaitsForUnlock()
    {
        var m = new MyMutex();

        m.Lock();

        var second = Task.Run(() =>
        {
            m.Lock();
            m.Unlock();
        });

        await Task.Delay(50);

        Assert.False(second.IsCompleted);

        m.Unlock();

        var completed = await Task.WhenAny(
            second,
            Task.Delay(TimeSpan.FromSeconds(2))
        );

        Assert.Same(second, completed);
    }

    [Fact]
    public async Task EveryWaiterWakesUp()
    {
        var m = new MyMutex();
        const int waiters = 50;

        m.Lock();

        var tasks = Enumerable.Range(0, waiters)
            .Select(_ => Task.Run(() =>
            {
                m.Lock();
                Thread.Sleep(1);
                m.Unlock();
            }))
            .ToArray();

        await Task.Delay(50);

        m.Unlock();

        var completed = await Task.WhenAny(
            Task.WhenAll(tasks),
            Task.Delay(TimeSpan.FromSeconds(10))
        );

        Assert.Equal(TaskStatus.RanToCompletion, completed.Status);
    }

    [Fact]
    public async Task TryLock()
    {
        var m = new MyMutex();

        Assert.True(m.TryLock());

        var task = Task.Run(() => m.TryLock());

        var completed = await Task.WhenAny(
            task,
            Task.Delay(TimeSpan.FromSeconds(1))
        );

        Assert.Same(task, completed);
        Assert.False(await task);

        m.Unlock();

        Assert.True(m.TryLock());

        m.Unlock();
    }

    [Fact]
    public void UnlockWithoutLockThrows()
    {
        var m = new MyMutex();

        Assert.Throws<Exception>(() => m.Unlock());
    }

    [Fact]
    public async Task HandoffUnderLoad()
    {
        var m = new MyMutex();

        var deadline = DateTime.UtcNow.AddMilliseconds(300);

        var tasks = Enumerable.Range(0, 8)
            .Select(_ => Task.Factory.StartNew(() =>
            {
                int passes = 0;

                while (DateTime.UtcNow < deadline)
                {
                    m.Lock();
                    passes++;
                    m.Unlock();
                }

                return passes;
            }, CancellationToken.None, TaskCreationOptions.LongRunning, TaskScheduler.Default))
            .ToArray();

        var results = await Task.WhenAll(tasks);

        foreach (var passes in results)
        {
            Assert.True(passes > 0);
        }
    }
}