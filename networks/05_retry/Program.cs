using System.Globalization;
using System.Net;
using System.Net.Http.Headers;

if (args.Length == 0)
{
    Console.Error.WriteLine("Error: URL is required");
    Environment.ExitCode = 1;
    return;
}

var url = args[0];
var method = "GET";
var maxAttempts = 5;
string? idempotencyKey = null;

for (var i = 1; i < args.Length; i++)
{
    switch (args[i])
    {
        case "--method":
            if (++i >= args.Length)
            {
                Console.Error.WriteLine("Error: --method requires a value");
                Environment.ExitCode = 1;
                return;
            }

            method = args[i];
            break;

        case "--max-attempts":
            if (++i >= args.Length ||
                !int.TryParse(args[i], out maxAttempts) ||
                maxAttempts < 1)
            {
                Console.Error.WriteLine("Error: --max-attempts must be a positive integer");
                Environment.ExitCode = 1;
                return;
            }

            break;

        case "--idempotency-key":
            if (++i >= args.Length)
            {
                Console.Error.WriteLine("Error: --idempotency-key requires a value");
                Environment.ExitCode = 1;
                return;
            }

            idempotencyKey = args[i];
            break;

        default:
            Console.Error.WriteLine($"Error: Unknown argument: {args[i]}");
            Environment.ExitCode = 1;
            return;
    }
}

if (!Uri.TryCreate(url, UriKind.Absolute, out var uri) ||
    uri.Scheme != Uri.UriSchemeHttp)
{
    Console.Error.WriteLine("Error: URL must be an absolute HTTP URL");
    Environment.ExitCode = 1;
    return;
}

if (method.Length == 0 || !method.All(IsTokenChar))
{
    Console.Error.WriteLine("Error: Invalid HTTP method");
    Environment.ExitCode = 1;
    return;
}

var canRetryMethod = IsIdempotentMethod(method) || idempotencyKey != null;
var success = false;
var attempts = 0;

for (var attempt = 1; attempt <= maxAttempts; attempt++)
{
    attempts = attempt;

    try
    {
        using var handler = new HttpClientHandler
        {
            AllowAutoRedirect = false,
            UseProxy = false
        };

        using var client = new HttpClient(handler)
        {
            Timeout = TimeSpan.FromSeconds(5)
        };

        client.DefaultRequestHeaders.ConnectionClose = true;

        using var request = new HttpRequestMessage(new HttpMethod(method), uri);

        if (idempotencyKey != null)
            request.Headers.Add("Idempotency-Key", idempotencyKey);

        using var response = await client.SendAsync(
            request,
            HttpCompletionOption.ResponseContentRead);

        Console.WriteLine($"attempt {attempt} status {(int)response.StatusCode}");

        if ((int)response.StatusCode is >= 200 and <= 399)
        {
            success = true;
            break;
        }

        if (!ShouldRetry((int)response.StatusCode) ||
            !canRetryMethod ||
            attempt == maxAttempts)
            break;

        var delay = GetDelay(attempt + 1, response.Headers.RetryAfter);
        Console.WriteLine($"sleep_ms {delay}");
        await Task.Delay(delay);
    }
    catch (Exception ex) when (ex is HttpRequestException or TaskCanceledException)
    {
        Console.WriteLine($"attempt {attempt} error {GetErrorText(ex)}");

        if (!canRetryMethod || attempt == maxAttempts)
            break;

        var delay = GetDelay(attempt + 1, null);
        Console.WriteLine($"sleep_ms {delay}");
        await Task.Delay(delay);
    }
}

Console.WriteLine($"result {(success ? "success" : "failure")} attempts {attempts}");
Environment.ExitCode = success ? 0 : 1;

static bool IsIdempotentMethod(string method) =>
    method is "GET" or "HEAD" or "PUT" or "DELETE" or "OPTIONS" or "TRACE";

static bool IsTokenChar(char c) =>
    (c >= 'a' && c <= 'z') ||
    (c >= 'A' && c <= 'Z') ||
    (c >= '0' && c <= '9') ||
    "!#$%&'*+-.^_`|~".Contains(c);

static bool ShouldRetry(int statusCode) =>
    statusCode is 429 or 500 or 502 or 503 or 504;

static int GetDelay(int attempt, RetryConditionHeaderValue? retryAfter)
{
    if (retryAfter?.Delta is { } delta)
        return ClampDelay(delta.TotalMilliseconds);

    if (retryAfter?.Date is { } date)
        return ClampDelay((date - DateTimeOffset.UtcNow).TotalMilliseconds);

    var interval = Math.Min(2000, 200 * Math.Pow(2, attempt - 1));
    return Random.Shared.Next(0, (int)interval + 1);
}

static int ClampDelay(double milliseconds)
{
    if (double.IsNaN(milliseconds) || milliseconds <= 0)
        return 0;

    if (milliseconds >= int.MaxValue)
        return int.MaxValue;

    return (int)Math.Ceiling(milliseconds);
}

static string GetErrorText(Exception ex) =>ex is TaskCanceledException ? "timeout" : ex.Message;