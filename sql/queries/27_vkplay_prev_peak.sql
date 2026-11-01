-- 27. Пик онлайна против пика предыдущего эфира того же стримера.
-- Только площадка 'VK Play Live'; prev_peak — пик предыдущего по started_at эфира того
-- же стримера, delta — их разность; первый эфир каждого стримера не выводится.
--
-- Колонки: streamer, followers, started_at, title, peak_viewers, prev_peak, delta
-- Порядок: followers по убыванию, внутри стримера started_at по возрастанию
-- Строк:   592

WITH stream_lags AS (
    SELECT
        s.nickname AS streamer,
        s.followers,
        st.started_at,
        st.title,
        st.peak_viewers,
        LAG(st.peak_viewers) OVER (
            PARTITION BY s.streamer_id
            ORDER BY st.started_at
        ) AS prev_peak
    FROM streamers s
    INNER JOIN platforms pl ON s.platform_id = pl.platform_id
    INNER JOIN streams st ON s.streamer_id = st.streamer_id
    WHERE pl.name = 'VK Play Live'
)
SELECT
    streamer,
    followers,
    started_at,
    title,
    peak_viewers,
    prev_peak,
    peak_viewers - prev_peak AS delta
FROM stream_lags
WHERE prev_peak IS NOT NULL
ORDER BY followers DESC, started_at ASC;