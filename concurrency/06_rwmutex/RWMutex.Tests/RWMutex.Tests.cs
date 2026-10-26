using System.Threading;
using DefaultNamespace;
using Xunit;

public class RwMutexTests
{
    private static Task Finished(Action action)
    {
        return Task.Run(action);
    }

    [Fact]
    public async Task WriterExcludesEveryone()
    {
        var rw = new MyRWMutex();
        rw.Lock();

        var reader = Finished(() =>
        {
            rw.RLock();
            rw.RUnlock();
        });

        var writer = Finished(() =>
        {
            rw.Lock();
            rw.Unlock();
        });

        await Task.Delay(50);

        Assert.False(reader.IsCompleted, "читатель прошёл, пока держит писатель");
        Assert.False(writer.IsCompleted, "второй писатель прошёл, пока держит первый");

        rw.Unlock();

        var readerCompleted = await Task.WhenAny(reader, Task.Delay(TimeSpan.FromSeconds(2)));
        Assert.Same(reader, readerCompleted);

        var writerCompleted = await Task.WhenAny(writer, Task.Delay(TimeSpan.FromSeconds(2)));
        Assert.Same(writer, writerCompleted);
    }

    [Fact]
    public async Task ReadersGoTogether()
    {
        var rw = new MyRWMutex();
        const int readers = 8;

        int inside = 0;
        int peak = 0;

        var tasks = Enumerable.Range(0, readers).Select(_ => Task.Run(() =>
        {
            rw.RLock();

            int now = Interlocked.Increment(ref inside);

            while (true)
            {
                int old = Volatile.Read(ref peak);

                if (now <= old || Interlocked.CompareExchange(ref peak, now, old) == old)
                    break;
            }

            Thread.Sleep(30);
            Interlocked.Decrement(ref inside);

            rw.RUnlock();
        })).ToArray();

        var allTasks = Task.WhenAll(tasks);
        var completed = await Task.WhenAny(allTasks, Task.Delay(TimeSpan.FromSeconds(5)));

        Assert.Same(allTasks, completed);
        Assert.True(peak >= 2, $"читатели шли по одному (пик {peak}) — смысл RWMutex теряется");
    }

    [Fact]
    public async Task WriterWaitsForReaders()
    {
        var rw = new MyRWMutex();
        rw.RLock();
        rw.RLock();

        var writer = Finished(() =>
        {
            rw.Lock();
            rw.Unlock();
        });

        await Task.Delay(50);
        Assert.False(writer.IsCompleted, "писатель прошёл при живых читателях");

        rw.RUnlock();

        await Task.Delay(50);
        Assert.False(writer.IsCompleted, "писатель прошёл, пока остался один читатель");

        rw.RUnlock();

        var completed = await Task.WhenAny(writer, Task.Delay(TimeSpan.FromSeconds(2)));
        Assert.Same(writer, completed);
    }

    [Fact]
    public async Task NoTornState()
    {
        var rw = new MyRWMutex();
        int shared = 0;
        int bad = 0;

        var writers = Enumerable.Range(0, 4).Select(_ => Task.Run(() =>
        {
            for (int j = 0; j < 5000; j++)
            {
                rw.Lock();
                shared++;
                shared++;
                rw.Unlock();
            }
        })).ToArray();

        var readers = Enumerable.Range(0, 8).Select(_ => Task.Run(() =>
        {
            for (int j = 0; j < 5000; j++)
            {
                rw.RLock();

                if (shared % 2 != 0)
                    Interlocked.Exchange(ref bad, 1);

                rw.RUnlock();
            }
        })).ToArray();

        var allTasks = Task.WhenAll(writers.Concat(readers));
        var completed = await Task.WhenAny(allTasks, Task.Delay(TimeSpan.FromSeconds(30)));

        Assert.Same(allTasks, completed);
        Assert.Equal(0, Volatile.Read(ref bad));
        Assert.Equal(4 * 5000 * 2, shared);
    }

    [Fact]
    public void UnlockWithoutLockPanics()
    {
        var rw = new MyRWMutex();
        Assert.Throws<Exception>(() => rw.Unlock());
    }

    [Fact]
    public void RUnlockWithoutRLockPanics()
    {
        var rw = new MyRWMutex();
        Assert.Throws<Exception>(() => rw.RUnlock());
    }
}