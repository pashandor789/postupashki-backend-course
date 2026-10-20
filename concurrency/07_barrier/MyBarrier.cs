using System.Threading;
using a;

namespace DefaultNamespace;

public class MyBarrier
{
    private readonly int _participants;
    private int _count;
    private int _generation;

    public MyBarrier(int n)
    {
        if (n <= 0)
            throw new ArgumentException();

        _participants = n;
    }

    public void Wait()
    {
        int generation = Volatile.Read(ref _generation);

        int count = Interlocked.Increment(ref _count);

        if (count == _participants)
        {
            Interlocked.Exchange(ref _count, 0);
            Interlocked.Increment(ref _generation);
            Futex.WakeAll(ref _generation);
            return;
        }

        while (Volatile.Read(ref _generation) == generation)
            Futex.Wait(ref _generation, generation);
    }
}