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

    var parts = line.Split(default(char[]), StringSplitOptions.RemoveEmptyEntries);

    if (parts.Length != 2)
    {
        Console.WriteLine($"query {string.Join(' ', parts)}");
        Console.WriteLine("status ERROR");
        Console.WriteLine("end");
        Console.Error.WriteLine("Error: Expected a domain name and a record type");
        Environment.ExitCode = 1;
        continue;
    }

    var name = parts[0];
    var type = parts[1];
    var key = (Name: name.TrimEnd('.').ToLowerInvariant(), Type: type.ToUpperInvariant());

    Console.WriteLine($"query {name} {type}");

    try
    {
        var typeCode = GetTypeCode(type);
        var encodedName = EncodeName(name);

        if (cache.TryGetValue(key, out var cached))
        {
            if (Stopwatch.GetTimestamp() < cached.ExpiresAt)
            {
                Console.WriteLine("status NOERROR");
                foreach (var cachedAnswer in cached.Answers)
                    Console.WriteLine($"answer {cachedAnswer.Type} {cachedAnswer.Value} {cachedAnswer.Ttl}");
                Console.WriteLine("end");
                continue;
            }

            cache.Remove(key);
        }

        var id = (ushort)RandomNumberGenerator.GetInt32(0, 65536);
        var query = BuildQuery(id, encodedName, typeCode);
        var expectedName = ReadName(encodedName, 0).Name;

        (byte[] Data, int Offset)? reply;

        using (var client = new UdpClient(serverAddress.AddressFamily))
        {
            await client.SendAsync(query, server);
            reply = await ReceiveResponseAsync(client, server, id, expectedName, typeCode);
        }

        if (reply is null)
        {
            Console.WriteLine("status TIMEOUT");
            Console.WriteLine("end");
            Environment.ExitCode = 1;
            continue;
        }

        var (response, offset) = reply.Value;
        var flags = ReadUInt16(response, 2);
        var answerCount = ReadUInt16(response, 6);
        var status = GetStatus(flags);
        var truncated = (flags & 0x0200) != 0;

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

        if (status == "NOERROR" && !truncated && answers.Count > 0 && !hasZeroTtl)
        {
            var expiresAt = Stopwatch.GetTimestamp()
                + (long)(minTtl * (double)Stopwatch.Frequency);

            cache[key] = new CacheEntry(answers, expiresAt);
        }
    }
    catch (DnsException ex)
    {
        Console.WriteLine("status ERROR");
        Console.WriteLine("end");
        Console.Error.WriteLine($"Error: {ex.Message}");
        Environment.ExitCode = 1;
    }
}

static async Task<(byte[] Data, int Offset)?> ReceiveResponseAsync(
    UdpClient client, IPEndPoint server, ushort id, string expectedName, ushort typeCode)
{
    using var cts = new CancellationTokenSource(TimeSpan.FromSeconds(5));

    while (!cts.IsCancellationRequested)
    {
        UdpReceiveResult result;

        try
        {
            result = await client.ReceiveAsync(cts.Token);
        }
        catch (OperationCanceledException)
        {
            return null;
        }
        catch (SocketException)
        {
            continue;
        }

        var data = result.Buffer;

        if (!result.RemoteEndPoint.Equals(server) ||
            data.Length < 12 ||
            ReadUInt16(data, 0) != id ||
            (ReadUInt16(data, 2) & 0x8000) == 0)
            continue;

        if (MatchQuestion(data, expectedName, typeCode) is { } answersOffset)
            return (data, answersOffset);
    }

    return null;
}

static int? MatchQuestion(byte[] data, string expectedName, ushort typeCode)
{
    try
    {
        var count = ReadUInt16(data, 4);

        if (count == 0)
            return (ReadUInt16(data, 2) & 0x000F) != 0 ? 12 : null;

        if (count != 1)
            return null;

        var question = ReadName(data, 12);
        var end = question.NextOffset + 4;

        if (end > data.Length ||
            !string.Equals(question.Name, expectedName, StringComparison.OrdinalIgnoreCase) ||
            ReadUInt16(data, question.NextOffset) != typeCode ||
            ReadUInt16(data, question.NextOffset + 2) != 1)
            return null;

        return end;
    }
    catch (DnsException)
    {
        return null;
    }
}

static byte[] BuildQuery(ushort id, byte[] encodedName, ushort typeCode)
{
    using var stream = new MemoryStream();
    WriteUInt16(stream, id);
    WriteUInt16(stream, 0x0100);
    WriteUInt16(stream, 1);
    WriteUInt16(stream, 0);
    WriteUInt16(stream, 0);
    WriteUInt16(stream, 0);
    stream.Write(encodedName);
    WriteUInt16(stream, typeCode);
    WriteUInt16(stream, 1);
    return stream.ToArray();
}

static byte[] EncodeName(string name)
{
    if (string.IsNullOrWhiteSpace(name) || name.Any(c => c > 127))
        throw new DnsException("Invalid domain name");

    if (name == ".")
        return [0];

    var normalized = name.EndsWith('.') ? name[..^1] : name;
    var labels = normalized.Split('.');
    var totalLength = 1;

    using var stream = new MemoryStream();

    foreach (var label in labels)
    {
        var bytes = Encoding.ASCII.GetBytes(label);

        if (bytes.Length is < 1 or > 63)
            throw new DnsException("Invalid DNS label length");

        totalLength += bytes.Length + 1;

        if (totalLength > 255)
            throw new DnsException("Domain name is too long");

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
        _ => throw new DnsException($"Unsupported type: {type}")
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
        throw new DnsException(message);
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

static string EscapeLabel(byte[] data, int start, int length)
{
    var builder = new StringBuilder(length);

    for (var i = start; i < start + length; i++)
    {
        var b = data[i];

        if (b is (byte)'.' or (byte)'\\')
            builder.Append('\\').Append((char)b);
        else if (b <= 0x20 || b >= 0x7F)
            builder.Append('\\').Append(b.ToString("D3"));
        else
            builder.Append((char)b);
    }

    return builder.ToString();
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
            throw new DnsException("DNS name compression loop");

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
        labels.Add(EscapeLabel(data, current, length));
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
    using var bytes = new MemoryStream();

    while (current < end)
    {
        var stringLength = data[current++];
        Ensure(current + stringLength <= end, "Invalid TXT record");
        bytes.Write(data, current, stringLength);
        current += stringLength;
    }

    var text = Encoding.UTF8.GetString(bytes.ToArray());
    var builder = new StringBuilder(text.Length);

    foreach (var c in text)
    {
        if (char.IsControl(c) || c is '\u2028' or '\u2029')
            builder.Append('\\').Append(((int)c).ToString("D3"));
        else
            builder.Append(c);
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

class DnsException(string message) : Exception(message);