import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const baseURL = __ENV.BASE_URL || 'http://api:8080';
const timeoutSeconds = Number(__ENV.PROCESSING_TIMEOUT_SECONDS || 20);
const readinessTimeoutSeconds = Number(__ENV.READINESS_TIMEOUT_SECONDS || 45);
const profile = __ENV.K6_PROFILE || 'smoke';
const stressPeak = Number(__ENV.K6_STRESS_PEAK || 1000);
const workerReplicas = Number(__ENV.WORKER_REPLICAS || 1);
const creationSuccess = new Rate('order_creation_success');
const processingSuccess = new Rate('order_processing_success');
const processingDuration = new Trend('order_processing_duration', true);
const httpThresholds = {
  http_req_failed: ['rate<0.01'],
  'http_req_duration{method:POST}': ['p(95)<500'],
};

const profiles = {
  smoke: {
    scenarios: { smoke: { executor: 'shared-iterations', vus: 1, iterations: 1, maxDuration: '30s' } },
    thresholds: {
      ...httpThresholds,
      order_creation_success: ['rate==1'],
      order_processing_success: ['rate==1'],
      order_processing_duration: ['p(95)<8000'],
    },
  },
  sustainable: {
    scenarios: {
      sustainable: {
        executor: 'constant-arrival-rate', rate: workerReplicas, timeUnit: '3s', duration: '1m',
        preAllocatedVUs: workerReplicas + 1, maxVUs: workerReplicas * 4,
      },
    },
    thresholds: {
      ...httpThresholds,
      order_creation_success: ['rate>0.99'],
      order_processing_success: ['rate>0.99'],
      order_processing_duration: ['p(95)<8000'],
    },
  },
  saturation: {
    scenarios: {
      saturation: {
        executor: 'constant-arrival-rate', rate: workerReplicas * 2, timeUnit: '1s', duration: '30s',
        preAllocatedVUs: workerReplicas * 4, maxVUs: workerReplicas * 16,
      },
    },
    thresholds: { ...httpThresholds, order_creation_success: ['rate>0.99'] },
  },
  stress: {
    scenarios: {
      stress: {
        executor: 'ramping-arrival-rate', exec: 'stress', startRate: Math.max(1, Math.round(stressPeak * 0.1)),
        timeUnit: '1s', preAllocatedVUs: 100, maxVUs: stressPeak * 2,
        stages: [
          { target: Math.max(1, Math.round(stressPeak * 0.25)), duration: '30s' },
          { target: Math.max(1, Math.round(stressPeak * 0.5)), duration: '30s' },
          { target: stressPeak, duration: '45s' },
          { target: stressPeak, duration: '30s' },
        ],
      },
    },
    thresholds: { ...httpThresholds, order_creation_success: ['rate>0.99'] },
  },
};

export const options = profiles[profile] || profiles.smoke;

export function setup() {
  const deadline = Date.now() + readinessTimeoutSeconds * 1000;
  while (Date.now() < deadline) {
    if (http.get(`${baseURL}/health`).status === 204 && warmup()) return;
    sleep(1);
  }
  throw new Error(`asynchronous order flow did not become ready within ${readinessTimeoutSeconds}s`);
}

export default function () {
  const response = createOrder();
  const id = response.json('id') || '';
  const created = check(response, { 'order is created': (result) => result.status === 201 && id !== '' });
  creationSuccess.add(created);
  if (!created) return;

  const started = Date.now();
  const processed = waitForProcessing(id);
  processingSuccess.add(processed);
  if (processed) processingDuration.add(Date.now() - started);
}

export function stress() {
  const response = createOrder();
  const created = check(response, { 'order is created': (result) => result.status === 201 });
  creationSuccess.add(created);
}

function waitForProcessing(id) {
  const deadline = Date.now() + timeoutSeconds * 1000;
  while (Date.now() < deadline) {
    const response = http.get(`${baseURL}/orders/${id}`);
    if (response.status === 200 && response.json('status') === 'PROCESSADO') return true;
    sleep(0.25);
  }
  return false;
}

function warmup() {
  const response = createOrder('warmup');
  const id = response.json('id') || '';
  return response.status === 201 && id !== '' && waitForProcessing(id);
}

function createOrder(prefix = 'order') {
  return http.post(`${baseURL}/orders`, JSON.stringify({
    product_name: `load-test-${prefix}-${Date.now()}`, quantity: 1,
  }), { headers: { 'Content-Type': 'application/json' } });
}
