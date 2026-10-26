using System.Threading;
using a;

namespace DefaultNamespace;

public class MySemaphore
{
    private uint _slots;

    public MySemaphore(int n)
    {
        _slots = (uint)n;
    }

    public void Acquire()
    {
        while (true)
        {
            if (TryAcquire())
                return;
            
            Futex.Wait(ref _slots, 0);
        }
    }

    public bool TryAcquire()
    {
        while (true)
        {
            var slots = Volatile.Read(ref _slots);

            if (slots == 0)
                return false;

            if (Interlocked.CompareExchange(
                    ref _slots,
                    slots - 1,
                    slots) == slots)
            {
                return true;
            }
        }
    }

    public void Release()
    {
        Interlocked.Increment(ref _slots);
        Futex.Wake(ref _slots);
    }

    public int Available()
    {
        return (int)Volatile.Read(ref _slots);
    }
}