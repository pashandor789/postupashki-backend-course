namespace SpinLock.Tests;
using DefaultNamespace;
using Xunit;
using DefaultNamespace;
using Xunit;

public class SpinLockTests
{
    private static IEnumerable<ILocker> GetLockers()
    {
        yield return new MySpinlock();
        yield return new MyTtasSpinlock();
    }

    [Fact]
    public void LockUnlock()
    {
        foreach (var l in GetLockers())
        {
            l.Lock();
            l.Unlock();

            l.Lock();
            l.Unlock();
        }
    }

    [Fact]
    public async Task MutualExclusion()
    {
        const int goroutines = 8;
        const int iterations = 20_000;

        foreach (var l in GetLockers())
        {
            int counter = 0;

            var tasks = Enumerable.Range(0, goroutines)
                .Select(_ => Task.Run(() =>
                {
                    for (int j = 0; j < iterations; j++)
                    {
                        l.Lock();
                        counter++;
                        l.Unlock();
                    }
                }))
                .ToArray();

            await Task.WhenAll(tasks);

            Assert.Equal(goroutines * iterations, counter);
        }
    }

    [Fact]
    public async Task OnlyOneInside()
    {
        foreach (var l in GetLockers())
        {
            int inside = 0;
            bool bad = false;

            var tasks = Enumerable.Range(0, 8)
                .Select(_ => Task.Run(() =>
                {
                    for (int j = 0; j < 2000; j++)
                    {
                        l.Lock();

                        inside++;

                        if (inside != 1)
                            bad = true;

                        inside--;

                        l.Unlock();
                    }
                }))
                .ToArray();

            await Task.WhenAll(tasks);

            Assert.False(bad);
        }
    }

    [Fact]
    public async Task TryLock()
    {
        foreach (var l in GetLockers())
        {
            Assert.True(l.TryLock());

            var task = Task.Run(() => l.TryLock());

            var completed = await Task.WhenAny(
                task,
                Task.Delay(TimeSpan.FromSeconds(1))
            );

            Assert.Same(task, completed);
            Assert.False(await task);

            l.Unlock();

            Assert.True(l.TryLock());

            l.Unlock();
        }
    }

    [Fact]
    public async Task LockWaitsForUnlock()
    {
        foreach (var l in GetLockers())
        {
            l.Lock();

            var second = Task.Run(() =>
            {
                l.Lock();
                l.Unlock();
            });

            await Task.Delay(50);

            Assert.False(second.IsCompleted);

            l.Unlock();

            var completed = await Task.WhenAny(
                second,
                Task.Delay(TimeSpan.FromSeconds(2))
            );

            Assert.Same(second, completed);
        }
    }

    [Fact]
    public void UnlockWithoutLockThrows()
    {
        foreach (var l in GetLockers())
        {
            Assert.Throws<Exception>(() => l.Unlock());
        }
    }
}