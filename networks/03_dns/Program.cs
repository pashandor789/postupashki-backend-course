using System.Diagnostics;
using System.Net;
using System.Net.Sockets;
using System.Security.Cryptography;
using System.Text;

if (args.Length != 2 ||
    !IPAddress.TryParse(args[0], out var serverAddress) ||
    !int.TryParse(args[1], out var port) ||
    port is < 1 or > 65535)
{
    Console.Error.WriteLine("Error: Expected an IP address and a valid port");
    Environment.ExitCode = 1;
    return;
}

var server = new IPEndPoint(serverAddress, port);
var cache = new Dictionary<(string Name, string Type), CacheEntry>();

while (Console.ReadLine() is { } line)
{
    if (string.IsNullOrWhiteSpace(line))
        continue;

    var parts = line.Split(' ', StringSplitOptions.RemoveEmptyEntries);

    if (parts.Length != 2)
    {
        Console.Error.WriteLine("Error: Expected a domain name and a record type");
        continue;
    }

    var name = parts[0];
    var type = parts[1].ToUpperInvariant();
    var normalizedName = name.TrimEnd('.').ToLowerInvariant();
    var key = (normalizedName, type);

    Console.WriteLine($"query {name} {type}");

    try
    {
        GetTypeCode(type);
        EncodeName(name);

        if (cache.TryGetValue(key, out var cached))
        {
            if (Stopwatch.GetTimestamp() < cached.ExpiresAt)
            {
                Console.WriteLine("status NOERROR");
                foreach (var answer in cached.Answers)
                    Console.WriteLine($"answer {answer.Type} {answer.Value} {answer.Ttl}");
                Console.WriteLine("end");
                continue;
            }

            cache.Remove(key);
        }

        var id = (ushort)RandomNumberGenerator.GetInt32(0, 65536);
        var query = BuildQuery(id, name, type);
        byte[] response = [];
        var received = false;
        var stopwatch = Stopwatch.StartNew();

        using (var client = new UdpClient(serverAddress.AddressFamily))
        {
            await client.SendAsync(query, server);

            while (stopwatch.Elapsed < TimeSpan.FromSeconds(5))
            {
                var remaining = TimeSpan.FromSeconds(5) - stopwatch.Elapsed;
                var receiveTask = client.ReceiveAsync();
                var timeoutTask = Task.Delay(remaining);
                var completed = await Task.WhenAny(receiveTask, timeoutTask);

                if (completed == timeoutTask)
                    break;

                UdpReceiveResult result;

                try
                {
                    result = await receiveTask;
                }
                catch (SocketException)
                {
                    continue;
                }

                if (!result.RemoteEndPoint.Equals(server) ||
                    result.Buffer.Length < 12 ||
                    ReadUInt16(result.Buffer, 0) != id ||
                    (ReadUInt16(result.Buffer, 2) & 0x8000) == 0)
                    continue;

                response = result.Buffer;
                received = true;
                break;
            }
        }

        if (!received)
        {
            Console.WriteLine("status TIMEOUT");
            Console.WriteLine("end");
            Environment.ExitCode = 1;
            continue;
        }

        var flags = ReadUInt16(response, 2);
        var questionCount = ReadUInt16(response, 4);
        var answerCount = ReadUInt16(response, 6);
        var status = GetStatus(flags);
        var offset = 12;

        for (var i = 0; i < questionCount; i++)
        {
            var question = ReadName(response, offset);
            offset = question.NextOffset;
            Ensure(offset + 4 <= response.Length, "Incomplete DNS question");
            offset += 4;
        }

        var answers = new List<DnsAnswer>();
        uint minTtl = uint.MaxValue;
        var hasZeroTtl = false;

        for (var i = 0; i < answerCount; i++)
        {
            var answerName = ReadName(response, offset);
            offset = answerName.NextOffset;
            Ensure(offset + 10 <= response.Length, "Incomplete DNS answer");

            var answerType = ReadUInt16(response, offset);
            var answerClass = ReadUInt16(response, offset + 2);
            var rawTtl = ReadUInt32(response, offset + 4);
            var ttl = (rawTtl & 0x80000000) != 0 ? 0 : rawTtl;
            var rdLength = ReadUInt16(response, offset + 8);
            offset += 10;

            Ensure(offset + rdLength <= response.Length, "Invalid DNS record length");

            var answer = answerClass == 1
                ? ParseAnswer(response, answerType, ttl, offset, rdLength)
                : null;

            offset += rdLength;

            if (answerClass != 1)
                continue;

            minTtl = Math.Min(minTtl, ttl);
            if (ttl == 0)
                hasZeroTtl = true;

            if (answer != null)
                answers.Add(answer);
        }

        Console.WriteLine($"status {status}");

        foreach (var answer in answers)
            Console.WriteLine($"answer {answer.Type} {answer.Value} {answer.Ttl}");

        Console.WriteLine("end");

        if (status == "NOERROR" && answers.Count > 0 && !hasZeroTtl)
        {
            var expiresAt = Stopwatch.GetTimestamp()
                + (long)(minTtl * (double)Stopwatch.Frequency);

            cache[key] = new CacheEntry(answers, expiresAt);
        }
    }
    catch (Exception ex) when (
        ex is ArgumentException or
        InvalidDataException or
        IndexOutOfRangeException or
        OverflowException or
        FormatException)
    {
        Console.WriteLine("status ERROR");
        Console.WriteLine("end");
        Console.Error.WriteLine($"Error: {ex.Message}");
        Environment.ExitCode = 1;
    }
}

static byte[] BuildQuery(ushort id, string name, string type)
{
    using var stream = new MemoryStream();
    WriteUInt16(stream, id);
    WriteUInt16(stream, 0x0100);
    WriteUInt16(stream, 1);
    WriteUInt16(stream, 0);
    WriteUInt16(stream, 0);
    WriteUInt16(stream, 0);
    stream.Write(EncodeName(name));
    WriteUInt16(stream, GetTypeCode(type));
    WriteUInt16(stream, 1);
    return stream.ToArray();
}

static byte[] EncodeName(string name)
{
    if (string.IsNullOrWhiteSpace(name) || name.Any(c => c > 127))
        throw new ArgumentException("Invalid domain name");

    var normalized = name.EndsWith('.') ? name[..^1] : name;
    var labels = normalized.Split('.');
    var totalLength = 1;

    using var stream = new MemoryStream();

    foreach (var label in labels)
    {
        var bytes = Encoding.ASCII.GetBytes(label);

        if (bytes.Length is < 1 or > 63)
            throw new ArgumentException("Invalid DNS label length");

        totalLength += bytes.Length + 1;

        if (totalLength > 255)
            throw new ArgumentException("Domain name is too long");

        stream.WriteByte((byte)bytes.Length);
        stream.Write(bytes);
    }

    stream.WriteByte(0);
    return stream.ToArray();
}

static ushort GetTypeCode(string type)
{
    return type.ToUpperInvariant() switch
    {
        "A" => 1,
        "NS" => 2,
        "CNAME" => 5,
        "MX" => 15,
        "TXT" => 16,
        "AAAA" => 28,
        _ => throw new ArgumentException($"Unsupported type: {type}")
    };
}

static void WriteUInt16(Stream stream, ushort value)
{
    stream.WriteByte((byte)(value >> 8));
    stream.WriteByte((byte)value);
}

static void Ensure(bool condition, string message)
{
    if (!condition)
        throw new InvalidDataException(message);
}

static ushort ReadUInt16(byte[] data, int offset)
{
    Ensure(offset >= 0 && offset + 2 <= data.Length, "Incomplete DNS field");
    return (ushort)((data[offset] << 8) | data[offset + 1]);
}

static uint ReadUInt32(byte[] data, int offset)
{
    Ensure(offset >= 0 && offset + 4 <= data.Length, "Incomplete DNS field");

    return ((uint)data[offset] << 24)
         | ((uint)data[offset + 1] << 16)
         | ((uint)data[offset + 2] << 8)
         | data[offset + 3];
}

static (string Name, int NextOffset) ReadName(
    byte[] data, int offset, int fieldEnd = -1)
{
    if (fieldEnd < 0)
        fieldEnd = data.Length;

    Ensure(offset >= 0 && offset < fieldEnd && fieldEnd <= data.Length,
        "Invalid DNS name offset");

    var labels = new List<string>();
    var current = offset;
    var nextOffset = offset;
    var jumped = false;
    var jumps = 0;
    var nameLength = 1;

    while (true)
    {
        var limit = jumped ? data.Length : fieldEnd;
        Ensure(current < limit, "Incomplete DNS name");

        if (jumps++ > 128)
            throw new InvalidDataException("DNS name compression loop");

        var length = data[current];

        if (length == 0)
        {
            if (!jumped)
                nextOffset = current + 1;
            break;
        }

        if ((length & 0xC0) == 0xC0)
        {
            Ensure(current + 2 <= limit, "Invalid DNS pointer");

            if (!jumped)
                nextOffset = current + 2;

            var pointer = ((length & 0x3F) << 8) | data[current + 1];
            Ensure(pointer < data.Length, "Invalid DNS pointer");
            current = pointer;
            jumped = true;
            continue;
        }

        Ensure((length & 0xC0) == 0, "Invalid DNS label");
        Ensure(current + 1 + length <= limit, "Incomplete DNS label");

        nameLength += length + 1;
        Ensure(nameLength <= 255, "DNS name is too long");

        current++;
        labels.Add(Encoding.ASCII.GetString(data, current, length));
        current += length;
    }

    return (string.Join(".", labels) + ".", nextOffset);
}

static DnsAnswer? ParseAnswer(
    byte[] data, ushort type, uint ttl, int offset, int length)
{
    var end = offset + length;

    return type switch
    {
        1 when length == 4 =>
            new DnsAnswer("A", new IPAddress(data.AsSpan(offset, 4)).ToString(), ttl),

        28 when length == 16 =>
            new DnsAnswer("AAAA", new IPAddress(data.AsSpan(offset, 16)).ToString(), ttl),

        5 => ParseDomainRecord(data, "CNAME", ttl, offset, end),
        2 => ParseDomainRecord(data, "NS", ttl, offset, end),
        15 when length >= 3 => ParseMx(data, offset, end, ttl),
        16 => ParseTxt(data, offset, length, ttl),
        _ => null
    };
}

static DnsAnswer ParseDomainRecord(
    byte[] data, string type, uint ttl, int offset, int end)
{
    var name = ReadName(data, offset, end);
    Ensure(name.NextOffset == end, "Invalid DNS name record");
    return new DnsAnswer(type, name.Name, ttl);
}

static DnsAnswer ParseMx(byte[] data, int offset, int end, uint ttl)
{
    Ensure(offset + 2 <= end, "Invalid MX record");

    var priority = ReadUInt16(data, offset);
    var exchange = ReadName(data, offset + 2, end);
    Ensure(exchange.NextOffset == end, "Invalid MX record");

    return new DnsAnswer("MX", $"{priority} {exchange.Name}", ttl);
}

static DnsAnswer ParseTxt(byte[] data, int offset, int length, uint ttl)
{
    var end = offset + length;
    var current = offset;
    var builder = new StringBuilder();

    while (current < end)
    {
        var stringLength = data[current++];
        Ensure(current + stringLength <= end, "Invalid TXT record");
        builder.Append(Encoding.UTF8.GetString(data, current, stringLength));
        current += stringLength;
    }

    return new DnsAnswer("TXT", builder.ToString(), ttl);
}

static string GetStatus(ushort flags)
{
    return (flags & 0x000F) switch
    {
        0 => "NOERROR",
        1 => "FORMERR",
        2 => "SERVFAIL",
        3 => "NXDOMAIN",
        5 => "REFUSED",
        var code => $"RCODE{code}"
    };
}

record DnsAnswer(string Type, string Value, uint Ttl);
record CacheEntry(List<DnsAnswer> Answers, long ExpiresAt);