using System.Text;

try
{
    var input = Console.In.ReadToEnd();
    var hex = new StringBuilder();

    foreach (var c in input)
    {
        if (Uri.IsHexDigit(c))
            hex.Append(c);
        else if (!char.IsWhiteSpace(c) && c != ':')
            throw new PacketException("Invalid hex");
    }

    Ensure(hex.Length % 2 == 0, "Invalid hex");
    var data = Convert.FromHexString(hex.ToString());
    var output = new StringBuilder();

    Ensure(data.Length >= 14, "Incomplete Ethernet");

    output.AppendLine($"eth.dst {Mac(data, 0)}");
    output.AppendLine($"eth.src {Mac(data, 6)}");

    ushort ethertype = ReadUInt16(data, 12);
    output.AppendLine($"eth.ethertype 0x{ethertype:x4}");

    if (ethertype == 0x0800)
    {
        const int ip = 14;
        Ensure(data.Length >= ip + 20, "Incomplete IP");

        int version = data[ip] >> 4;
        int ihlBytes = (data[ip] & 0x0f) * 4;

        Ensure(version == 4, "Invalid IP version");
        Ensure(ihlBytes >= 20, "Invalid IP header length");
        Ensure(data.Length >= ip + ihlBytes, "Incomplete IP");

        ushort totalLength = ReadUInt16(data, ip + 2);
        Ensure(totalLength >= ihlBytes, "Invalid IP total length");
        Ensure(data.Length >= ip + totalLength, "Incomplete IP packet");

        ushort id = ReadUInt16(data, ip + 4);
        ushort flagsAndOffset = ReadUInt16(data, ip + 6);
        int ttl = data[ip + 8];
        int protocol = data[ip + 9];

        output.AppendLine($"ip.version {version}");
        output.AppendLine($"ip.ihl_bytes {ihlBytes}");
        output.AppendLine($"ip.total_length {totalLength}");
        output.AppendLine($"ip.id 0x{id:x4}");
        output.AppendLine($"ip.ttl {ttl}");
        output.AppendLine($"ip.protocol {protocol}");

        bool df = (flagsAndOffset & 0x4000) != 0;
        bool mf = (flagsAndOffset & 0x2000) != 0;
        var flags = new List<string>();

        if (df)
            flags.Add("DF");
        if (mf)
            flags.Add("MF");

        output.AppendLine($"ip.flags {(flags.Count == 0 ? "none" : string.Join(",", flags))}");
        output.AppendLine($"ip.frag_offset {(flagsAndOffset & 0x1fff) * 8}");
        output.AppendLine($"ip.src {IpAddress(data, ip + 12)}");
        output.AppendLine($"ip.dst {IpAddress(data, ip + 16)}");
        output.AppendLine($"ip.checksum_valid {ChecksumValid(data, ip, ihlBytes).ToString().ToLower()}");

        int transport = ip + ihlBytes;
        int ipPayloadLength = totalLength - ihlBytes;
        bool firstFragment = (flagsAndOffset & 0x1fff) == 0;

        if (protocol == 6 && firstFragment)
        {
            Ensure(ipPayloadLength >= 20, "Incomplete TCP");

            ushort srcPort = ReadUInt16(data, transport);
            ushort dstPort = ReadUInt16(data, transport + 2);
            uint seq = ReadUInt32(data, transport + 4);
            uint ack = ReadUInt32(data, transport + 8);
            int dataOffsetBytes = (data[transport + 12] >> 4) * 4;
            ushort tcpFlags = (ushort)(ReadUInt16(data, transport + 12) & 0x01ff);
            ushort window = ReadUInt16(data, transport + 14);

            Ensure(dataOffsetBytes >= 20, "Invalid TCP header length");
            Ensure(dataOffsetBytes <= ipPayloadLength, "Incomplete TCP");

            output.AppendLine($"tcp.src_port {srcPort}");
            output.AppendLine($"tcp.dst_port {dstPort}");
            output.AppendLine($"tcp.seq {seq}");
            output.AppendLine($"tcp.ack {ack}");
            output.AppendLine($"tcp.data_offset_bytes {dataOffsetBytes}");
            output.AppendLine($"tcp.flags {TcpFlags(tcpFlags)}");
            output.AppendLine($"tcp.window {window}");
            output.AppendLine($"payload.length {ipPayloadLength - dataOffsetBytes}");
        }
        else if (protocol == 17 && firstFragment)
        {
            Ensure(ipPayloadLength >= 8, "Incomplete UDP");

            ushort srcPort = ReadUInt16(data, transport);
            ushort dstPort = ReadUInt16(data, transport + 2);
            ushort udpLength = ReadUInt16(data, transport + 4);

            Ensure(udpLength >= 8, "Invalid UDP length");
            Ensure(udpLength <= ipPayloadLength, "Invalid UDP length");

            output.AppendLine($"udp.src_port {srcPort}");
            output.AppendLine($"udp.dst_port {dstPort}");
            output.AppendLine($"udp.length {udpLength}");
            output.AppendLine($"payload.length {udpLength - 8}");
        }
        else
        {
            output.AppendLine($"payload.length {ipPayloadLength}");
        }
    }

    Console.Write(output.ToString());
}
catch (PacketException ex)
{
    Console.Error.WriteLine($"Error: {ex.Message}");
    Environment.ExitCode = 1;
}

static void Ensure(bool condition, string error)
{
    if (!condition)
        throw new PacketException(error);
}

static ushort ReadUInt16(byte[] data, int offset)
{
    return (ushort)((data[offset] << 8) | data[offset + 1]);
}

static uint ReadUInt32(byte[] data, int offset)
{
    return ((uint)data[offset] << 24)
         | ((uint)data[offset + 1] << 16)
         | ((uint)data[offset + 2] << 8)
         | data[offset + 3];
}

static string Mac(byte[] data, int offset)
{
    return string.Join(":", data.Skip(offset).Take(6).Select(x => x.ToString("x2")));
}

static string IpAddress(byte[] data, int offset)
{
    return string.Join(".", data.Skip(offset).Take(4));
}

static bool ChecksumValid(byte[] data, int offset, int length)
{
    uint sum = 0;

    for (int i = 0; i < length; i += 2)
    {
        sum += ReadUInt16(data, offset + i);
        sum = (sum & 0xffff) + (sum >> 16);
    }

    while (sum >> 16 != 0)
        sum = (sum & 0xffff) + (sum >> 16);

    return sum == 0xffff;
}

static string TcpFlags(ushort flags)
{
    string[] names = ["FIN", "SYN", "RST", "PSH", "ACK", "URG"];
    var result = new List<string>();

    for (int i = 0; i < names.Length; i++)
    {
        if ((flags & (1 << i)) != 0)
            result.Add(names[i]);
    }

    return result.Count == 0 ? "none" : string.Join(",", result);
}

class PacketException(string message) : Exception(message);