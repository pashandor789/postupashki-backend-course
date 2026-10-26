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
    private static extern bool WaitOnAddress(
        ref int Address,
        ref int CompareAddress,
        UIntPtr AddressSize,
        uint dwMilliseconds);

    [DllImport("api-ms-win-core-synch-l1-2-0.dll",
        ExactSpelling = true)]
    private static extern void WakeByAddressSingle(
        ref uint Address);

    [DllImport("api-ms-win-core-synch-l1-2-0.dll",
        ExactSpelling = true)]
    private static extern void WakeByAddressAll(
        ref int Address);

    public static void Wait(ref uint address, uint expected)
    {
        uint compare = expected;

        WaitOnAddress(
            ref address,
            ref compare,
            (UIntPtr)sizeof(uint),
            uint.MaxValue);
    }

    public static void Wait(ref int address, int expected)
    {
        int compare = expected;

        WaitOnAddress(
            ref address,
            ref compare,
            (UIntPtr)sizeof(int),
            uint.MaxValue);
    }

    public static void Wake(ref uint address)
    {
        WakeByAddressSingle(ref address);
    }

    public static void WakeAll(ref int address)
    {
        WakeByAddressAll(ref address);
    }
}