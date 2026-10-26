namespace SpinLock.Tests;

public interface ILocker
{
    void Lock();
    void Unlock();
    bool TryLock();
}