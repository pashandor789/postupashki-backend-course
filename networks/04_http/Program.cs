using System.Globalization;
using System.Text;

using var memory = new MemoryStream();
await Console.OpenStandardInput().CopyToAsync(memory);
var data = memory.ToArray();

string output;

try
{
    output = Parse(data);
}
catch (HttpParseException ex)
{
    output = $"error {ex.Reason}\n";
    Environment.ExitCode = 1;
}

using var stdout = Console.OpenStandardOutput();
stdout.Write(Encoding.Latin1.GetBytes(output));
stdout.Flush();

static string Parse(byte[] data)
{
    var sb = new StringBuilder();
    var startLineEnd = Find(data, [13, 10]);

    if (startLineEnd < 0)
        throw new HttpParseException("bad_start_line");

    var firstLf = Array.IndexOf(data, (byte)10);
    if (firstLf != startLineEnd + 1)
        throw new HttpParseException("bad_start_line");

    var firstLine = Encoding.Latin1.GetString(data[..startLineEnd]);

    if (firstLine.StartsWith("HTTP/", StringComparison.Ordinal))
    {
        sb.Append("type response\n");
        ParseResponseLine(firstLine, sb);
    }
    else
    {
        sb.Append("type request\n");
        ParseRequestLine(firstLine, sb);
    }

    var headerEnd = Find(data, [13, 10, 13, 10], startLineEnd);
    if (headerEnd < 0)
        throw new HttpParseException("bad_header");

    var headerBytes = data[(startLineEnd + 2)..(headerEnd + 2)];
    var headers = ParseHeaders(SplitLines(headerBytes));

    foreach (var (name, value) in headers)
        sb.Append($"header {name} {value}\n");

    var bodyStart = headerEnd + 4;
    var transferEncodings = headers
        .Where(h => h.Name == "transfer-encoding")
        .SelectMany(h => h.Value.Split(','))
        .Select(s => s.Trim(' ', '\t'))
        .ToList();

    var contentLengths = headers
        .Where(h => h.Name == "content-length")
        .SelectMany(h => h.Value.Split(','))
        .Select(s => s.Trim(' ', '\t'))
        .ToList();

    if (transferEncodings.Any(s => s.Length == 0))
        throw new HttpParseException("bad_header");

    if (transferEncodings.Count > 0 && contentLengths.Count > 0)
        throw new HttpParseException("bad_header");

    var isChunked = transferEncodings.Count > 0 &&
        transferEncodings[^1].Equals("chunked", StringComparison.OrdinalIgnoreCase);

    if (transferEncodings.Any(s => s.Equals("chunked", StringComparison.OrdinalIgnoreCase)) &&
        !isChunked)
        throw new HttpParseException("bad_header");

    byte[] body;

    if (isChunked)
    {
        body = ReadChunked(data, bodyStart);
    }
    else if (contentLengths.Count > 0)
    {
        long? expectedLength = null;

        foreach (var text in contentLengths)
        {
            if (text.Length == 0 || !text.All(c => c >= '0' && c <= '9') ||
                !long.TryParse(text, NumberStyles.None, CultureInfo.InvariantCulture, out var length))
                throw new HttpParseException("bad_header");

            if (expectedLength.HasValue && expectedLength.Value != length)
                throw new HttpParseException("bad_header");

            expectedLength = length;
        }

        var available = data.Length - bodyStart;
        if (expectedLength!.Value > available)
            throw new HttpParseException("incomplete_body");

        body = data[bodyStart..(bodyStart + (int)expectedLength.Value)];
    }
    else
    {
        body = [];
    }

    sb.Append($"body.length {body.Length}\n");

    if (body.Length > 0 && body.All(b => b >= 0x20 && b <= 0x7E))
        sb.Append($"body.text {Encoding.ASCII.GetString(body)}\n");

    return sb.ToString();
}

static byte[] ReadChunked(byte[] data, int pos)
{
    using var result = new MemoryStream();

    while (true)
    {
        var lineEnd = Find(data, [13, 10], pos);
        if (lineEnd < 0)
            throw new HttpParseException("bad_chunk");

        var sizeText = Encoding.Latin1.GetString(data[pos..lineEnd]);
        var semi = sizeText.IndexOf(';');
        if (semi >= 0)
            sizeText = sizeText[..semi].TrimEnd(' ', '\t');

        if (sizeText.Length == 0 || !sizeText.All(Uri.IsHexDigit) ||
            !ulong.TryParse(sizeText, NumberStyles.AllowHexSpecifier,
                CultureInfo.InvariantCulture, out var size))
            throw new HttpParseException("bad_chunk");

        pos = lineEnd + 2;

        if (size == 0)
        {
            SkipTrailer(data, pos);
            return result.ToArray();
        }

        var remaining = data.Length - pos;
        if (remaining < 2 || size > (ulong)(remaining - 2))
            throw new HttpParseException("bad_chunk");

        var chunkLength = (int)size;
        var end = pos + chunkLength;

        if (data[end] != 13 || data[end + 1] != 10)
            throw new HttpParseException("bad_chunk");

        result.Write(data, pos, chunkLength);
        pos = end + 2;
    }
}

static void SkipTrailer(byte[] data, int pos)
{
    while (true)
    {
        var lineEnd = Find(data, [13, 10], pos);
        if (lineEnd < 0)
            throw new HttpParseException("bad_chunk");

        if (lineEnd == pos)
            return;

        pos = lineEnd + 2;
    }
}

static bool IsTokenChar(char c) =>
    (c >= 'a' && c <= 'z') ||
    (c >= 'A' && c <= 'Z') ||
    (c >= '0' && c <= '9') ||
    "!#$%&'*+-.^_`|~".Contains(c);

static List<(string Name, string Value)> ParseHeaders(List<byte[]> lines)
{
    var headers = new List<(string Name, string Value)>();

    foreach (var line in lines)
    {
        var colon = Array.IndexOf(line, (byte)':');

        if (colon <= 0)
            throw new HttpParseException("bad_header");

        var name = Encoding.Latin1.GetString(line[..colon]);

        if (!name.All(IsTokenChar))
            throw new HttpParseException("bad_header");

        for (var i = colon + 1; i < line.Length; i++)
        {
            var b = line[i];

            if ((b < 0x20 && b != 0x09) || b == 0x7F)
                throw new HttpParseException("bad_header");
        }

        var start = colon + 1;

        while (start < line.Length && (line[start] == 32 || line[start] == 9))
            start++;

        var end = line.Length;

        while (end > start && (line[end - 1] == 32 || line[end - 1] == 9))
            end--;

        var value = Encoding.Latin1.GetString(line[start..end]);
        headers.Add((name.ToLowerInvariant(), value));
    }

    return headers;
}

static void ParseRequestLine(string line, StringBuilder sb)
{
    var parts = line.Split(' ');

    if (parts.Length != 3 ||
        parts[0].Length == 0 ||
        !parts[0].All(IsTokenChar) ||
        parts[1].Length == 0 ||
        parts[1].Any(c => c <= 0x20 || c >= 0x7F) ||
        parts[2] != "HTTP/1.1")
        throw new HttpParseException("bad_start_line");

    sb.Append($"method {parts[0]}\n");
    sb.Append($"target {parts[1]}\n");
    sb.Append($"version {parts[2]}\n");
}

static void ParseResponseLine(string line, StringBuilder sb)
{
    if (!line.StartsWith("HTTP/1.1 ", StringComparison.Ordinal))
        throw new HttpParseException("bad_start_line");

    var rest = line[9..];

    if (rest.Length < 3 ||
        !rest[..3].All(c => c >= '0' && c <= '9') ||
        (rest.Length > 3 && rest[3] != ' '))
        throw new HttpParseException("bad_start_line");

    var status = rest[..3];
    var reason = rest.Length > 4 ? rest[4..] : "";
    if (reason.Any(c => (c < 0x20 && c != '\t') || c == 0x7F))
        throw new HttpParseException("bad_start_line");

    sb.Append("version HTTP/1.1\n");
    sb.Append($"status {status}\n");
    sb.Append($"reason {reason}\n");
}

static List<byte[]> SplitLines(byte[] data)
{
    var result = new List<byte[]>();
    var start = 0;

    while (start < data.Length)
    {
        var end = Find(data, [13, 10], start);

        if (end < 0)
            throw new HttpParseException("bad_header");

        result.Add(data[start..end]);
        start = end + 2;
    }

    return result;
}

static int Find(byte[] data, byte[] pattern, int start = 0)
{
    if (start < 0 || start > data.Length)
        return -1;

    var index = data.AsSpan(start).IndexOf(pattern);
    return index < 0 ? -1 : start + index;
}

class HttpParseException(string reason) : Exception(reason)
{
    public string Reason { get; } = reason;
}