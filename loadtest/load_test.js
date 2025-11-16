import http from 'k6/http';
import { check } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
  scenarios: {
    create_pr_load: {
      executor: 'constant-arrival-rate',
      rate: 20,
      timeUnit: '1s',
      duration: '5m',
      preAllocatedVUs: 30,
      maxVUs: 50,
      exec: 'createPR',
    },
    
    read_operations: {
      executor: 'constant-arrival-rate',
      rate: 50,
      timeUnit: '1s',
      duration: '5m',
      preAllocatedVUs: 20,
      maxVUs: 40,
      exec: 'readOperations',
    },
    
    spike_test: {
      executor: 'ramping-arrival-rate',
      startRate: 5,
      timeUnit: '1s',
      stages: [
        { duration: '1m', target: 5 },
        { duration: '30s', target: 30 },
        { duration: '2m', target: 30 },
        { duration: '1m', target: 5 },
      ],
      preAllocatedVUs: 30,
      maxVUs: 50,
      exec: 'spikeTest',
    },
  },
  
  thresholds: {
    http_req_duration: ['p(95)<1000', 'p(99)<2000'],
    http_req_failed: ['rate<0.02'],
    errors: ['rate<0.05'],
  },
};

const BASE_URL = 'http://localhost:8080';

const users = [
  '550e8400-e29b-41d4-a716-446655440001',
  '550e8400-e29b-41d4-a716-446655440002',
  '550e8400-e29b-41d4-a716-446655440003',
  '550e8400-e29b-41d4-a716-446655440004',
];

function generateUUID() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0;
    const v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

export function createPR() {
  const prID = generateUUID();
  const authorID = users[Math.floor(Math.random() * users.length)];
  
  const payload = JSON.stringify({
    pull_request_id: prID,
    pull_request_name: `Feature ${prID.substring(0, 8)}`,
    author_id: authorID,
  });

  const res = http.post(`${BASE_URL}/pullRequest/create`, payload, {
    headers: { 'Content-Type': 'application/json' },
    timeout: '10s',
  });

  const success = check(res, {
    'create PR is 201': (r) => r.status === 201,
  });

  errorRate.add(!success);
}

export function readOperations() {
  const operations = [
    () => http.get(`${BASE_URL}/health`),
    () => http.get(`${BASE_URL}/team/get?team_name=backend`),
    () => http.get(`${BASE_URL}/users/getReview?user_id=${users[0]}`),
    () => http.get(`${BASE_URL}/stats`),
  ];
  
  const op = operations[Math.floor(Math.random() * operations.length)];
  const res = op();
  
  check(res, {
    'read operation is 200': (r) => r.status === 200,
  });
}

export function spikeTest() {
  const prID = generateUUID();
  const authorID = users[Math.floor(Math.random() * users.length)];
  
  const payload = JSON.stringify({
    pull_request_id: prID,
    pull_request_name: `Spike PR ${prID.substring(0, 8)}`,
    author_id: authorID,
  });

  const res = http.post(`${BASE_URL}/pullRequest/create`, payload, {
    headers: { 'Content-Type': 'application/json' },
    timeout: '10s',
  });

  check(res, {
    'spike PR is 201': (r) => r.status === 201,
  });
}
