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
        ExactSpelling = true,
        EntryPoint = "WakeByAddressSingle")]
    private static extern void WakeByAddressSingleUInt(ref uint Address);

    [DllImport("api-ms-win-core-synch-l1-2-0.dll",
        ExactSpelling = true,
        EntryPoint = "WakeByAddressAll")]
    private static extern void WakeByAddressAllUInt(ref uint Address);

    [DllImport("api-ms-win-core-synch-l1-2-0.dll",
        ExactSpelling = true,
        EntryPoint = "WakeByAddressAll")]
    private static extern void WakeByAddressAllInt(ref int Address);


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
        WakeByAddressSingleUInt(ref address);
    }

    public static void WakeAll(ref uint address)
    {
        WakeByAddressAllUInt(ref address);
    }

    public static void WakeAll(ref int address)
    {
        WakeByAddressAllInt(ref address);
    }
}