#!/usr/bin/env node

const BACKEND_BASE_URL = process.env.BACKEND_BASE_URL || 'http://localhost:8080';
const REQUEST_TIMEOUT_MS = Number.parseInt(process.env.SMOKE_TIMEOUT_MS || '60000', 10);
const SMOKE_PASSWORD = process.env.SMOKE_PASSWORD || 'Smoke123456!';

function assert(condition, message) {
  if (!condition) {
    throw new Error(message);
  }
}

function buildUrl(pathname) {
  return new URL(pathname, BACKEND_BASE_URL).toString();
}

async function request(pathname, options = {}) {
  const {
    method = 'GET',
    token,
    body,
    expectedStatus = [200],
    headers = {},
  } = options;

  const url = buildUrl(pathname);
  const requestHeaders = { ...headers };
  if (token) {
    requestHeaders.Authorization = `Bearer ${token}`;
  }

  const init = {
    method,
    headers: requestHeaders,
  };

  if (body !== undefined) {
    if (body instanceof FormData) {
      init.body = body;
    } else if (
      typeof body === 'string' ||
      body instanceof Blob ||
      body instanceof URLSearchParams
    ) {
      init.body = body;
    } else {
      init.headers['Content-Type'] = 'application/json';
      init.body = JSON.stringify(body);
    }
  }

  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS);

  try {
    const res = await fetch(url, { ...init, signal: controller.signal });
    const raw = await res.text();

    let json = null;
    try {
      json = raw ? JSON.parse(raw) : null;
    } catch {
      json = null;
    }

    if (!expectedStatus.includes(res.status)) {
      throw new Error(
        `请求失败 ${method} ${pathname}，HTTP ${res.status}\n响应: ${raw.slice(0, 1200)}`
      );
    }

    return { status: res.status, json, raw };
  } finally {
    clearTimeout(timer);
  }
}

function expectApiSuccess(step, response) {
  assert(response.json && typeof response.json === 'object', `${step}: 响应不是 JSON`);
  assert(response.json.code === 0, `${step}: 业务返回失败 -> ${JSON.stringify(response.json)}`);
  return response.json.data;
}

async function main() {
  const suffix = `${Date.now()}${Math.floor(Math.random() * 10000)}`;
  const username = `smoke_${suffix}`;
  const email = `${username}@example.com`;

  let token = '';
  let lessonId = '';
  let documentId = '';

  console.log(`[smoke] backend: ${BACKEND_BASE_URL}`);

  try {
    const health = await request('/health');
    assert(health.json?.status === 'ok', `健康检查失败: ${health.raw}`);
    console.log('[smoke] ✓ health');

    const registerResp = await request('/api/v1/auth/register', {
      method: 'POST',
      body: {
        username,
        email,
        password: SMOKE_PASSWORD,
        full_name: 'Smoke Test User',
      },
    });
    expectApiSuccess('注册', registerResp);
    console.log('[smoke] ✓ register');

    const loginResp = await request('/api/v1/auth/login', {
      method: 'POST',
      body: {
        username,
        password: SMOKE_PASSWORD,
      },
    });
    const loginData = expectApiSuccess('登录', loginResp);
    token = loginData?.access_token || '';
    assert(token, '登录成功但未拿到 access_token');
    console.log('[smoke] ✓ login');

    const generateResp = await request('/api/v1/generate', {
      method: 'POST',
      token,
      body: {
        subject: '数学',
        grade: '七年级',
        topic: '一元一次方程',
        duration: 45,
        objectives: ['理解一元一次方程的定义并能完成基础求解'],
        keywords: ['方程', '等式性质'],
        style: '启发式',
        difficulty: 'medium',
      },
    });
    const generateData = expectApiSuccess('生成教案', generateResp);
    assert(generateData?.id, '生成教案返回缺少 id');
    assert(generateData?.status, '生成教案返回缺少 status');
    console.log(`[smoke] ✓ generate (status=${generateData.status})`);

    const createLessonResp = await request('/api/v1/lessons', {
      method: 'POST',
      token,
      expectedStatus: [201],
      body: {
        title: 'Smoke Test Lesson',
        subject: '数学',
        grade: '七年级',
        duration: 45,
        objectives: '掌握一元一次方程概念',
        content: '一元一次方程基础讲解',
        activities: '课堂练习 10 分钟',
        assessment: '随堂小测',
        resources: '教材、练习册',
        tags: ['smoke', 'automation'],
      },
    });
    const lessonData = expectApiSuccess('保存教案', createLessonResp);
    lessonId = lessonData?.id || '';
    assert(lessonId, '保存教案成功但未拿到 lesson id');
    console.log('[smoke] ✓ save lesson');

    const form = new FormData();
    form.append(
      'file',
      new Blob(
        [
          '# Smoke 文档\n\n这是 smoke 测试文档，用于验证知识文档上传链路是否可用。\n',
        ],
        { type: 'text/markdown' }
      ),
      'smoke-test.md'
    );
    form.append('title', 'Smoke Knowledge Doc');
    form.append('subject', '数学');
    form.append('grade', '7');

    const uploadResp = await request('/api/v1/knowledge/documents', {
      method: 'POST',
      token,
      body: form,
    });
    const uploadData = expectApiSuccess('上传知识文档', uploadResp);
    documentId = uploadData?.id || '';
    assert(documentId, '上传知识文档成功但未拿到 document id');
    console.log('[smoke] ✓ upload knowledge document');

    console.log('[smoke] ✅ all core checks passed');
  } finally {
    if (token && documentId) {
      try {
        await request(`/api/v1/knowledge/documents/${documentId}`, {
          method: 'DELETE',
          token,
        });
        console.log('[smoke] cleanup document done');
      } catch (error) {
        console.warn('[smoke] cleanup document failed:', error.message);
      }
    }

    if (token && lessonId) {
      try {
        await request(`/api/v1/lessons/${lessonId}`, {
          method: 'DELETE',
          token,
        });
        console.log('[smoke] cleanup lesson done');
      } catch (error) {
        console.warn('[smoke] cleanup lesson failed:', error.message);
      }
    }
  }
}

main().catch((error) => {
  console.error('[smoke] ❌ failed');
  console.error(error.message);
  process.exit(1);
});
