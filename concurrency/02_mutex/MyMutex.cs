using System.Threading;
using a;

namespace DefaultNamespace;

public class MyMutex
{
    private const uint Free = 0;
    private const uint Held = 1;
    private const uint Contended = 2;

    private uint _state = Free;

    public bool TryLock()
    {
        return Interlocked.CompareExchange(ref _state, Held, Free) == 0;
    }

    public void Lock()
    {

        while (!TryLock())
        {
            if (Interlocked.CompareExchange(ref _state, Contended, Held) != Free)
                Futex.Wait(ref _state, Contended);
        }
    }

    public void Unlock()
    {
        var prev = Interlocked.Exchange(ref _state, 0);
        if (prev == Free) throw new Exception();
        else if (prev == Contended) Futex.Wake(ref _state);
    }
}