using System.Runtime.InteropServices;

namespace a;

public static class Futex
{
    [DllImport("api-ms-win-core-synch-l1-2-0.dll",
        ExactSpelling = true)]
    private static extern bool WaitOnAddress(
        ref uint Address,
        ref uint CompareAddress,
        UIntPtr AddressSize,
        uint dwMilliseconds);

    [DllImport("api-ms-win-core-synch-l1-2-0.dll",
        ExactSpelling = true)]
    private static extern void WakeByAddressSingle(
        ref uint Address);

    public static void Wait(ref uint address, uint expected)
    {
        uint compare = expected;

        WaitOnAddress(
            ref address,
            ref compare,
            (UIntPtr)sizeof(uint),
            uint.MaxValue);
    }

    public static void Wake(ref uint address)
    {
        WakeByAddressSingle(ref address);
    }
}