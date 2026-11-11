-- 26. Нарастающая сумма донатов внутри каждого стрима Вжлинка за 2021 год.
-- Стримы Вжлинка, начавшиеся в 2021 году по UTC, строка на каждый донат; running_total —
-- нарастающая сумма донатов своего стрима по sent_at_sec.
--
-- Колонки: stream_id, title, sent_at_sec, donor, amount, running_total
-- Порядок: stream_id по возрастанию, внутри стрима sent_at_sec по возрастанию
-- Строк:   192

SELECT
    st.stream_id,
    st.title,
    sd.sent_at_sec,
    sd.nickname AS donor,
    sd.amount,
    SUM(sd.amount) OVER (
        PARTITION BY st.stream_id
        ORDER BY sd.sent_at_sec
        ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
    ) AS running_total
FROM streamers s
INNER JOIN streams st ON s.streamer_id = st.streamer_id
INNER JOIN stream_donations sd ON sd.stream_id = st.stream_id
WHERE (s.nickname = 'Вжлинк')
    AND EXTRACT(YEAR FROM st.started_at AT TIME ZONE 'UTC') = 2021
ORDER BY st.stream_id ASC, sd.sent_at_sec ASC;