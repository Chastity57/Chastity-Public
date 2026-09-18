/* 全栈开发1 · 前端逻辑
   页面上的三个选项完全对应终端里的 1 注册 / 2 登录 / 3 退出，
   点击后由桥接服务（server/main.go）转成后端的标准输入。 */

/* 用 file:// 直接打开页面时，接口指向本机桥接服务；由桥接服务托管时用同源 */
const API = location.protocol === 'file:' ? 'http://127.0.0.1:8080' : '';

const card      = document.getElementById('card');
const views     = { menu: 'view-menu', register: 'view-register', login: 'view-login' };
const result    = document.getElementById('result');
const resultMsg = document.getElementById('resultMsg');
const resultMeta= document.getElementById('resultMeta');
const consoleOut= document.getElementById('consoleOut');
const conn      = document.getElementById('conn');
const connText  = document.getElementById('connText');
const greet     = document.getElementById('greet');
const greetName = document.getElementById('greetName');
const shell     = document.querySelector('main.shell');

let current = 'menu';
let lastUsername = '';
let transcript = [];
let greetLastFocus = null;

/* ----------------------------------------------------------- 小工具 */

function setView(name) {
  current = name;
  card.dataset.view = name;
  for (const [key, id] of Object.entries(views)) {
    document.getElementById(id).classList.toggle('is-active', key === name);
  }
  if (name !== 'menu') {
    const form = document.querySelector(`#${views[name]} form`);
    const first = form.querySelector('input[name="username"]');
    // 进入表单后把用户名沿用上一次输入，省一次打字
    if (first && lastUsername && !first.value) first.value = lastUsername;
    // 等切换动画开始后再聚焦，避免浏览器把画面滚来滚去
    setTimeout(() => (first || form.querySelector('input')).focus({ preventScroll: true }), 60);
  }
}

function showResult(tone, message, meta) {
  result.dataset.tone = tone;
  resultMsg.textContent = message;
  resultMeta.textContent = meta || '';
}

function pushTranscript(lines) {
  if (!lines || !lines.length) return;
  if (transcript.length === 0) consoleOut.textContent = '';
  transcript = transcript.concat(lines);
  consoleOut.textContent = transcript.join('\n');
  consoleOut.scrollTop = consoleOut.scrollHeight;
}

function setBusy(form, busy) {
  const button = form.querySelector('button[type="submit"]');
  if (!button) return;
  if (busy) {
    button.dataset.label = button.textContent;
    button.textContent = '提交中…';
    button.setAttribute('aria-busy', 'true');
  } else {
    if (button.dataset.label) button.textContent = button.dataset.label;
    button.removeAttribute('aria-busy');
  }
}

async function callAPI(path, body) {
  let res;
  try {
    res = await fetch(API + path, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body || {}),
    });
  } catch (err) {
    const e = new Error('连接不上桥接服务');
    e.offline = true;
    throw e;
  }
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    const e = new Error(data.message || `请求失败（HTTP ${res.status}）`);
    throw e;
  }
  return data;
}

function toneOf(code) {
  if (code === 'register_success' || code === 'login_success' || code === 'exit_ok') return 'ok';
  if (code === 'duplicate_username') return 'warn';
  return 'err';
}

function metaOf(data) {
  const lines = data.output || [];
  return lines.length ? `后端输出：${lines.join(' ｜ ')}` : '';
}

/* ----------------------------------------------------------- 请求动作 */

async function submitForm(form, path, stageName) {
  const username = form.username.value.trim();
  const password = form.password.value;

  if (!username || !password) {
    showResult('err', '请填完用户名和密码', '后端用 fmt.Scan 逐项读取，两项都不能为空');
    return;
  }
  if (/\s/.test(username) || /\s/.test(password)) {
    showResult('err', '不能包含空格或换行', '后端按空白分词读取输入，带空格会被拆开');
    return;
  }
  lastUsername = username;
  setBusy(form, true);
  showResult('idle', '正在与后端通信…', `已向终端程序发送 ${stageName}`);

  try {
    const data = await callAPI(path, { username, password });
    pushTranscript(data.output);
    showResult(toneOf(data.code), data.message, metaOf(data));
    if (data.code === 'login_success') openGreeting(username);
  } catch (err) {
    if (err.offline) {
      setConn('offline');
      showResult('err', '连接不上桥接服务', '请先运行 启动.ps1（或 cd server && go run main.go）再试');
    } else {
      showResult('err', err.message, '');
    }
  } finally {
    setBusy(form, false);
  }
}

async function doExit() {
  showResult('idle', '正在退出后端程序…', '对应终端里的 3,退出程序');
  try {
    const data = await callAPI('/api/exit');
    pushTranscript(data.output);
    showResult('ok', data.message || '已退出后端程序', '注意：后端把用户存在内存里，程序退出后已注册账号会全部丢失');
    setConn('offline');
  } catch (err) {
    if (err.offline) {
      setConn('offline');
      showResult('err', '连接不上桥接服务', '请先运行 启动.ps1');
    } else {
      showResult('err', err.message, '');
    }
  }
}

function setConn(state) {
  conn.dataset.state = state;
  connText.textContent = state === 'online' ? '后端已就绪'
    : state === 'offline' ? '后端已退出' : '正在检测后端…';
}

/* ------------------------------------------------- 登录成功后的问候卡片 */

/* 打开：先把动画整体复位（no-anim），再用下一帧放开，
   这样「你」写出 →「好」写出 → 划一笔 → 用户名淡入 会完整跑一遍 */
function openGreeting(username) {
  greetName.textContent = username;
  greetLastFocus = document.activeElement;

  greet.classList.add('no-anim');
  greet.classList.remove('is-open');
  void greet.offsetWidth;                 // 强制回流，保证过渡从第一帧开始
  greet.classList.remove('no-anim');
  greet.classList.add('is-open');

  if (shell) shell.inert = true;          // 卡片弹出时，后面内容不再接受键盘焦点
  // 等卡片落位、书写动画开始后再移入焦点，避免打断开场
  setTimeout(() => {
    const closeBtn = greet.querySelector('.greet__close');
    if (closeBtn) closeBtn.focus({ preventScroll: true });
  }, 340);
}

function closeGreeting() {
  if (!greet.classList.contains('is-open')) return;
  greet.classList.remove('is-open');
  if (shell) shell.inert = false;
  if (greetLastFocus && document.contains(greetLastFocus)) {
    greetLastFocus.focus({ preventScroll: true });
  }
}

async function checkStatus() {
  try {
    const res = await fetch(API + '/api/status');
    const data = await res.json();
    setConn(data.alive ? 'online' : 'offline');
  } catch {
    setConn('offline');
  }
}

/* ----------------------------------------------------------- 事件绑定 */

document.querySelectorAll('[data-go]').forEach((button) => {
  button.addEventListener('click', () => {
    const target = button.dataset.go;
    if (target === 'exit') {
      doExit();
    } else {
      setView(target);
    }
  });
});

document.querySelectorAll('[data-back]').forEach((button) => {
  button.addEventListener('click', () => setView('menu'));
});

document.querySelectorAll('[data-greet-close]').forEach((el) => {
  el.addEventListener('click', closeGreeting);
});

document.getElementById('form-register').addEventListener('submit', (e) => {
  e.preventDefault();
  submitForm(e.currentTarget, '/api/register', '注册（1）');
});

document.getElementById('form-login').addEventListener('submit', (e) => {
  e.preventDefault();
  submitForm(e.currentTarget, '/api/login', '登录（2）');
});

/* 键盘：菜单界面按 1/2/3 等同点击；表单里按 Esc 返回菜单 */
document.addEventListener('keydown', (e) => {
  // 问候卡片打开时，Esc 优先关掉卡片
  if (greet.classList.contains('is-open') && e.key === 'Escape') {
    e.preventDefault();
    closeGreeting();
    return;
  }
  if (e.target.matches('input, textarea')) {
    if (e.key === 'Escape') setView('menu');
    return;
  }
  if (current !== 'menu' || e.metaKey || e.ctrlKey || e.altKey) return;
  if (e.key === '1') setView('register');
  if (e.key === '2') setView('login');
  if (e.key === '3') doExit();
});

checkStatus();

/* 允许用 ?view=register / ?view=login 直接打开某个表单（便于把链接发给别人） */
const initial = new URLSearchParams(location.search).get('view');
if (initial === 'register' || initial === 'login') setView(initial);
