import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API = `${BASE_URL}/api/v1`;
const USER_COUNT = 30;   // matches the seeder
const PRODUCT_COUNT = 100;

export const options = {
  scenarios: {
    // Anonymous shoppers: list -> detail
    browse: {
      executor: 'ramping-vus',
      exec: 'browse',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 50 },
        { duration: '1m', target: 50 },
        { duration: '15s', target: 0 },
      ],
    },
    // Logged-in buyers hammering checkout (exercises the row lock)
    checkout: {
      executor: 'constant-vus',
      exec: 'checkout',
      vus: 20,
      duration: '1m45s',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    'http_req_duration{scenario:browse}': ['p(95)<300'],
    'http_req_duration{scenario:checkout}': ['p(95)<500'],
  },
};

// Log in all 30 seeded users once; VUs reuse the tokens. Logging in on every
// iteration would turn this into a bcrypt benchmark instead of a checkout one.
export function setup() {
  const tokens = [];
  for (let i = 1; i <= USER_COUNT; i++) {
    const res = http.post(
      `${API}/auth/login`,
      JSON.stringify({ email: `user${i}@example.com`, password: 'password123' }),
      { headers: { 'Content-Type': 'application/json' } },
    );
    check(res, { 'login 200': (r) => r.status === 200 });
    tokens.push(res.json('token'));
  }
  return { tokens };
}

export function browse() {
  const page = Math.floor(Math.random() * 5) + 1;
  const list = http.get(`${API}/products?page=${page}&limit=20`, { tags: { name: 'list' } });
  check(list, { 'list 200': (r) => r.status === 200 });

  const id = Math.floor(Math.random() * PRODUCT_COUNT) + 1;
  const detail = http.get(`${API}/products/${id}`, { tags: { name: 'detail' } });
  check(detail, { 'detail 200': (r) => r.status === 200 });

  sleep(1);
}

export function checkout(data) {
  const token = data.tokens[(__VU - 1) % data.tokens.length];
  const res = http.post(
    `${API}/checkout`,
    JSON.stringify({
      product_id: Math.floor(Math.random() * PRODUCT_COUNT) + 1,
      quantity: Math.floor(Math.random() * 3) + 1,
    }),
    {
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      tags: { name: 'checkout' },
    },
  );
  check(res, { 'checkout 201': (r) => r.status === 201 });

  sleep(0.5);
}
