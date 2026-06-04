/**
 * Script k6 — Test de charge Kura Platform
 * Usage : k6 run scripts/k6-load-test.js
 *
 * Scénario : simulation d'un utilisateur DevOps utilisant la plateforme
 * - Authentification (login + refresh token)
 * - Consultation des états Terraform
 * - Consultation des pipelines CI/CD
 * - Consultation des métriques
 */

import http from 'k6/http'
import { check, sleep } from 'k6'
import { Rate, Trend } from 'k6/metrics'

// ── Configuration ─────────────────────────────────────────────────────────────

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8000'
const TEST_USER_EMAIL = __ENV.TEST_EMAIL || 'test-k6@kura.io'
const TEST_USER_PASSWORD = __ENV.TEST_PASSWORD || 'K6testPassword123!'

// ── Métriques custom ──────────────────────────────────────────────────────────

const authErrors = new Rate('auth_errors')
const apiErrors  = new Rate('api_errors')
const loginDuration = new Trend('login_duration_ms', true)

// ── Profil de charge ──────────────────────────────────────────────────────────

export const options = {
  stages: [
    { duration: '30s', target: 5  },  // Montée progressive → 5 VUs
    { duration: '2m',  target: 10 },  // Charge nominale → 10 VUs
    { duration: '1m',  target: 20 },  // Pic de charge → 20 VUs
    { duration: '30s', target: 0  },  // Descente
  ],
  thresholds: {
    // SLOs définis dans A2.1
    http_req_duration: ['p(95)<500'],  // p95 < 500ms (objectif A2.1 : <200ms hors auth)
    http_req_failed:   ['rate<0.05'],  // Taux d'erreur < 5%
    auth_errors:       ['rate<0.02'],  // Erreurs auth < 2%
  },
}

// ── Setup : créer l'utilisateur de test ───────────────────────────────────────

export function setup() {
  const registerRes = http.post(
    `${BASE_URL}/api/v1/auth/register`,
    JSON.stringify({
      email: TEST_USER_EMAIL,
      password: TEST_USER_PASSWORD,
      username: 'k6-test-user',
      first_name: 'K6',
      last_name: 'Test',
    }),
    { headers: { 'Content-Type': 'application/json' } }
  )
  // 201 = créé, 400 = déjà existant (les deux sont OK)
  if (registerRes.status !== 201 && registerRes.status !== 400) {
    console.warn(`Setup register: ${registerRes.status} ${registerRes.body}`)
  }
  return {}
}

// ── Scénario principal ────────────────────────────────────────────────────────

export default function () {
  const headers = { 'Content-Type': 'application/json' }

  // 1. Login
  const t0 = Date.now()
  const loginRes = http.post(
    `${BASE_URL}/api/v1/auth/login`,
    JSON.stringify({ email: TEST_USER_EMAIL, password: TEST_USER_PASSWORD }),
    { headers }
  )
  loginDuration.add(Date.now() - t0)

  const loginOk = check(loginRes, {
    'login 200': (r) => r.status === 200,
    'token présent': (r) => {
      try { return !!JSON.parse(r.body).token } catch { return false }
    },
  })
  authErrors.add(!loginOk)

  if (!loginOk) {
    sleep(1)
    return
  }

  const token = JSON.parse(loginRes.body).token
  const authHeaders = { ...headers, Authorization: `Bearer ${token}` }

  sleep(0.5)

  // 2. Health check
  const healthRes = http.get(`${BASE_URL.replace(':8000', ':8080')}/health`)
  check(healthRes, { 'health 200': (r) => r.status === 200 })

  sleep(0.5)

  // 3. États Terraform
  const tfRes = http.get(
    `${BASE_URL}/api/v1/terraform/states?project_id=default`,
    { headers: authHeaders }
  )
  const tfOk = check(tfRes, {
    'terraform states 200 ou 404': (r) => r.status === 200 || r.status === 404,
  })
  apiErrors.add(!tfOk)

  sleep(0.5)

  // 4. Pipelines
  const pipeRes = http.get(
    `${BASE_URL}/api/v1/pipeline/runs`,
    { headers: authHeaders }
  )
  const pipeOk = check(pipeRes, {
    'pipeline runs 200': (r) => r.status === 200,
  })
  apiErrors.add(!pipeOk)

  sleep(0.5)

  // 5. Métriques
  const metricsRes = http.get(
    `${BASE_URL}/api/v1/metrics/overview`,
    { headers: authHeaders }
  )
  check(metricsRes, {
    'metrics overview 200': (r) => r.status === 200,
  })

  sleep(1)
}

// ── Teardown : supprimer l'utilisateur de test ────────────────────────────────

export function teardown() {
  // Pas de suppression automatique — l'utilisateur k6 peut rester pour les runs suivants
  console.log('Test terminé. Utilisateur k6 conservé pour les prochains runs.')
}
