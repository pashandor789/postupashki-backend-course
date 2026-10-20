using System.Threading;
using a;

namespace DefaultNamespace;

public class MyOnce
{
    private const uint NotStarted = 0;
    private const uint Running = 1;
    private const uint DoneState = 2;
    private uint _state = NotStarted;

    public void Do(Action f)
    {
        if (Volatile.Read(ref _state) == DoneState)
            return;

        if (Interlocked.CompareExchange(ref _state, Running, NotStarted) == NotStarted)
        {
            try
            {
                f();
            }
            finally
            {
                Volatile.Write(ref _state, DoneState);
                Futex.WakeAll(ref _state);
            }
            return;
        }

        while (true)
        {
            var state = Volatile.Read(ref _state);
            if (state == DoneState)
                return;
            Futex.Wait(ref _state, Running);
        }
    }
    public bool Done()
    {
        return Volatile.Read(ref _state) == DoneState;
    }
}