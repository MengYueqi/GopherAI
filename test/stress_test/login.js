import http from 'k6/http';
import { check, sleep } from 'k6';

import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";

export function handleSummary(data) {
  return {
    "test/stress_test/result/login.html": htmlReport(data),
  };
}

export const options = {
  scenarios: {
    login_stress: {
      executor: 'ramping-arrival-rate',
      startRate: Number(__ENV.START_RATE) || 50,
      timeUnit: '1s',
      preAllocatedVUs: Number(__ENV.PRE_VUS) || 50,
      maxVUs: Number(__ENV.MAX_VUS) || 5000,
      stages: [
        // { duration: '30s', target: Number(__ENV.TARGET_RATE_1) || 100 },
        // { duration: '30s', target: Number(__ENV.TARGET_RATE_2) || 300 },
        // { duration: '30s', target: Number(__ENV.TARGET_RATE_3) || 600 },
        // { duration: '30s', target: Number(__ENV.TARGET_RATE_4) || 900 },
        { duration: '120s', target: Number(__ENV.TARGET_RATE_5) || 5000 },
        // { duration: '30s', target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<1000'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:9090';
const LOGIN_PATH = '/api/v1/user/login';

export default function () {
  const url = `${BASE_URL}${LOGIN_PATH}`;
  const payload = JSON.stringify({
    username: __ENV.LOGIN_USERNAME || '34890980136',
    password: __ENV.LOGIN_PASSWORD || '123456',
  });
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const res = http.post(url, payload, params);
  check(res, {
    'status is 200': (r) => r.status === 200,
  });
  sleep(1);
}
