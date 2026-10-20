using System.Threading;
using a;

namespace DefaultNamespace;

public class MyWaitgroup
{
    private int _slots;
    public void Add(int jobs)
    {
        while (true)
        {
            int cur = Volatile.Read(ref _slots);
            int next = cur + jobs;
            if (next < 0)
                throw new Exception();
            if (Interlocked.CompareExchange(ref _slots, next,cur) == cur)
            {
                return;
            }
        }
    }

    public void Done()
    {
        int value;

        while (true)
        {
            int cur = Volatile.Read(ref _slots);
            if (cur == 0)
                throw new Exception();

            int next = cur - 1;
            if (Interlocked.CompareExchange(ref _slots, next, cur) == cur)
            {
                value = next;
                break;
            }
        }
        if (value == 0)
        {
            Futex.WakeAll(ref _slots);
        }
    }

    public void Wait()
    {
        while (true)
        {
            int count = Volatile.Read(ref _slots);
            if (count == 0)
                return;
            Futex.Wait(ref _slots, count);
        }
    }
}