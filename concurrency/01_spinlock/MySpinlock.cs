using SpinLock.Tests;

namespace DefaultNamespace;
using System.Threading;

public class MySpinlock: ILocker
{
    private int _state = 0;

    public bool TryLock()
    {
        return Interlocked.CompareExchange(ref _state, 1, 0) == 0;
    }

    public void Lock()
    {
        while (!TryLock()){ }

    }

    public void Unlock()
    {
        if (Interlocked.Exchange(ref _state, 0) == 0) throw new Exception();
    }
}