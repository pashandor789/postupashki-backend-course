using System.Threading;
using a;

namespace DefaultNamespace;

public class MyRWMutex
{
    private const uint WriterBit = 1u << 31;
    private const uint ReaderMask = ~WriterBit;
    private uint _state;

    public void RLock()
    {
        while (true)
        {
            uint state = Volatile.Read(ref _state);

            if ((state & WriterBit) != 0)
            {
                Futex.Wait(ref _state, state);
                continue;
            }

            uint newState = state + 1;

            if (Interlocked.CompareExchange(ref _state, newState, state) == state)
                return;
        }
    }

    public void RUnlock()
    {
        while (true)
        {
            uint state = Volatile.Read(ref _state);

            if ((state & WriterBit) != 0)
                throw new Exception();

            uint readers = state & ReaderMask;

            if (readers == 0)
                throw new Exception();

            uint newState = state - 1;

            if (Interlocked.CompareExchange(ref _state, newState, state) == state)
            {
                if (readers == 1)
                    Futex.Wake(ref _state);

                return;
            }
        }
    }

    public void Lock()
    {
        while (true)
        {
            uint state = Volatile.Read(ref _state);

            if (state == 0)
            {
                if (Interlocked.CompareExchange(ref _state, WriterBit, 0) == 0)
                    return;

                continue;
            }

            Futex.Wait(ref _state, state);
        }
    }

    public void Unlock()
    {
        if (Interlocked.CompareExchange(ref _state, 0, WriterBit) != WriterBit)
            throw new Exception();

        Futex.Wake(ref _state);
    }
}