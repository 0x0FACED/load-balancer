import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

export let rateLimited = new Rate('rate_limited');

export let options = {
  stages: [
    { duration: '5s', target: 50 },
    { duration: '30s', target: 50 },
    { duration: '5s', target: 0 },
  ],
};

export default function () {
  const clientId = Math.floor(Math.random() * 10) + 1; // 10 уникальных клиентов
  const headers = {
    'X-Client-ID': clientId.toString(),
  };

  const res = http.get('http://localhost:8080/ping', { headers });

  check(res, {
    'status is 200 or 429': (r) => r.status === 200 || r.status === 429,
  });

  rateLimited.add(res.status === 429);
  sleep(0.1);
}
