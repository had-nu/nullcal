package web

// indexHTML is the single-file browser client served at /.
// It opens a WebSocket to /ws, renders state updates from the server,
// and sends action messages for every user interaction.
const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>nullcal</title>
<link rel="icon" type="image/svg+xml" href="/favicon.svg">
<link href="https://fonts.googleapis.com/css2?family=Share+Tech+Mono&display=swap" rel="stylesheet">
<style>
*,*::before,*::after{box-sizing:border-box;margin:0;padding:0}
:root{
  --bg:#111;--bg2:#1a1a1a;--bg3:#1e1e1e;
  --fg:#e0e0e0;--dim:#555;--accent:#fff;--border:#2a2a2a;
  --overdue:#ff5555;--duesoon:#ffb86c;--done:#50fa7b;
  --sel-bg:#e0e0e0;--sel-fg:#111;
  --font:'Share Tech Mono',monospace;
}
html,body{height:100%;background:var(--bg);color:var(--fg);font-family:var(--font);font-size:15px;line-height:1.4;overflow:hidden}

/* ── LAYOUT ── */
#app{display:flex;flex-direction:column;height:100vh}
#header{flex-shrink:0;display:flex;align-items:baseline;gap:16px;padding:8px 14px 6px;border-bottom:1px solid var(--border);background:var(--bg2)}
#header .logo{color:var(--accent);font-weight:bold;font-size:16px;letter-spacing:.15em}
#header .meta{color:var(--accent);background:var(--bg3);padding:1px 10px;font-size:14px;letter-spacing:.06em}
#header .conn{margin-left:auto;font-size:13px;color:var(--dim)}
#header .conn.ok{color:#50fa7b}
#header .conn.err{color:var(--overdue)}

#body{flex:1;display:flex;flex-direction:column;overflow:hidden;min-height:0}
#top{flex:1;display:flex;overflow:hidden;min-height:0;border-bottom:1px solid var(--border)}

/* ── TOOLBAR ── */
#toolbar{flex-shrink:0;display:flex;gap:6px;padding:6px 14px;border-bottom:1px solid var(--border);background:var(--bg2)}
#toolbar button{
  background:transparent;border:1px solid var(--border);color:var(--dim);
  font-family:var(--font);font-size:13px;padding:2px 10px;cursor:pointer;
  letter-spacing:.06em;transition:color .15s,border-color .15s;
}
#toolbar button:hover{color:var(--accent);border-color:var(--accent)}
#toolbar button.active{color:var(--accent);border-color:var(--accent)}
#toolbar .sep{flex:1}
#toolbar .week-nav{display:flex;align-items:center;gap:8px;color:var(--dim);font-size:13px}
#toolbar .week-nav button{padding:2px 8px}

/* ── CALENDAR ── */
#cal{flex:0 0 60%;border-right:1px solid var(--border);display:flex;flex-direction:column;overflow:hidden;min-height:0}
#cal-grid{display:grid;grid-template-columns:repeat(7,1fr);flex:1;overflow:hidden}
.cal-col{border-right:1px solid var(--border);display:flex;flex-direction:column;overflow:hidden;cursor:default}
.cal-col:last-child{border-right:none}
.cal-col-hdr{padding:4px 6px;font-size:13px;font-weight:bold;letter-spacing:.05em;border-bottom:1px solid var(--border);white-space:nowrap;overflow:hidden;flex-shrink:0}
.cal-col-hdr.today{color:var(--accent);background:#1c1c1c}
.cal-col-body{padding:5px 4px;display:flex;flex-direction:column;gap:2px;overflow-y:auto;flex:1}
.rb{font-size:13px;color:#666;padding:1px 0;flex-shrink:0}
.rb-lbl{color:#777;font-weight:bold}
.task-gap{height:3px}

/* ── TODO ── */
#todo{flex:0 0 40%;display:flex;flex-direction:column;overflow:hidden;min-height:0}
#todo.hidden{display:none}

/* ── KANBAN ── */
#kan{flex:0 0 200px;display:flex;flex-direction:column;overflow:hidden;border-top:1px solid var(--border)}
#kan.hidden{display:none}
#kan-grid{display:grid;grid-template-columns:repeat(3,1fr);flex:1;overflow:hidden;min-height:0}
.kan-col{border-right:1px solid var(--border);display:flex;flex-direction:column;overflow:hidden;min-height:0}
.kan-col:last-child{border-right:none}
.kan-col-body{padding:4px 8px;display:flex;flex-direction:column;gap:3px;overflow-y:auto;flex:1;min-height:0}

/* ── PANE HEADER ── */
.pane-hdr{padding:4px 10px;font-size:13px;font-weight:bold;letter-spacing:.1em;border-bottom:1px solid var(--border);flex-shrink:0}

/* ── TASK ITEMS ── */
.ti{
  font-size:13.5px;padding:2px 5px;cursor:pointer;
  display:flex;align-items:baseline;gap:4px;
  border-radius:1px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;
  transition:background .1s;user-select:none;flex-shrink:0;
}
.ti:hover{background:#1f1f1f}
.ti.selected{background:var(--sel-bg);color:var(--sel-fg)}
.ti.t-normal{color:var(--fg)}
.ti.t-duesoon{color:var(--duesoon)}
.ti.t-overdue{color:var(--overdue)}
.ti.t-done{color:var(--done);text-decoration:line-through}
.ti .pfx{flex-shrink:0;width:12px;color:inherit;opacity:.6}
.ti .lbl{overflow:hidden;text-overflow:ellipsis;flex:1}
.ti .tag{font-size:12px;background:#222;color:var(--dim);padding:0 4px;flex-shrink:0}

/* ── STATUS + HELP ── */
#statusbar{flex-shrink:0;padding:3px 14px;font-size:13px;background:#141414;border-top:1px solid var(--border);display:flex;justify-content:space-between}
#statusbar .sl{color:var(--fg)}
#statusbar .sl em{color:var(--accent);font-style:normal}
#help{flex-shrink:0;padding:3px 14px 5px;font-size:12.5px;color:var(--dim);background:#141414;display:flex;flex-wrap:wrap;gap:0 14px}
#help kbd{color:#888;font-family:var(--font);font-size:12.5px}

/* ── MODAL ── */
#modal-overlay{
  display:none;position:fixed;inset:0;background:rgba(0,0,0,.7);
  align-items:center;justify-content:center;z-index:100;
}
#modal-overlay.open{display:flex}
#modal{
  background:var(--bg2);border:1px solid var(--fg);padding:20px 24px;
  width:420px;max-width:95vw;font-family:var(--font);
}
#modal h2{font-size:15px;color:var(--accent);margin-bottom:16px;letter-spacing:.1em}
#modal label{display:block;font-size:13px;color:var(--fg);font-weight:bold;margin-bottom:4px;margin-top:12px;letter-spacing:.05em}
#modal label:first-of-type{margin-top:0}
#modal input{
  display:block;width:100%;background:var(--bg);border:1px solid var(--border);
  color:var(--fg);font-family:var(--font);font-size:14px;padding:4px 8px;
  outline:none;transition:border-color .15s;
}
#modal input:focus{border-color:var(--accent)}
#modal .err{color:var(--overdue);font-size:13px;margin-top:10px;min-height:16px}
#modal .actions{display:flex;gap:8px;margin-top:16px}
#modal .actions button{
  background:transparent;border:1px solid var(--border);color:var(--dim);
  font-family:var(--font);font-size:13px;padding:4px 16px;cursor:pointer;
  letter-spacing:.06em;transition:color .15s,border-color .15s;
}
#modal .actions button.primary{border-color:var(--accent);color:var(--accent)}
#modal .actions button:hover{color:var(--fg);border-color:var(--fg)}

/* ── CONFIRM ── */
#confirm-overlay{
  display:none;position:fixed;inset:0;background:rgba(0,0,0,.7);
  align-items:center;justify-content:center;z-index:100;
}
#confirm-overlay.open{display:flex}
#confirm{
  background:var(--bg2);border:1px solid var(--fg);padding:20px 24px;
  width:280px;font-family:var(--font);
}
#confirm p{font-size:14px;color:var(--fg);margin-bottom:16px}
#confirm .actions{display:flex;gap:8px}
#confirm .actions button{
  background:transparent;border:1px solid var(--border);color:var(--dim);
  font-family:var(--font);font-size:13px;padding:4px 16px;cursor:pointer;
}
#confirm .actions button.danger{border-color:var(--overdue);color:var(--overdue)}
#confirm .actions button:hover{color:var(--fg);border-color:var(--fg)}

/* scrollbar */
::-webkit-scrollbar{width:4px;height:4px}
::-webkit-scrollbar-track{background:transparent}
::-webkit-scrollbar-thumb{background:var(--border);border-radius:2px}
</style>
</head>
<body>
<div id="app">

  <!-- HEADER -->
  <div id="header">
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="18" height="18" fill="none">
      <rect x="8" y="8" width="84" height="84" rx="3" stroke="#e0e0e0" stroke-width="6"/>
      <line x1="8" y1="29" x2="92" y2="29" stroke="#e0e0e0" stroke-width="6"/>
      <line x1="19" y1="87" x2="81" y2="13" stroke="#e0e0e0" stroke-width="9" stroke-linecap="square"/>
    </svg>
    <span class="logo">NULLCAL</span>
    <span class="meta" id="hdr-meta"></span>
    <span class="conn" id="conn-status">● connecting</span>
  </div>

  <!-- TOOLBAR -->
  <div id="toolbar">
    <button onclick="openCreate()">[ n ] new</button>
    <button id="btn-todo" onclick="toggleTodo()">[ | ] todo</button>
    <button id="btn-kan"  onclick="toggleKan()">[ - ] kanban</button>
    <div class="sep"></div>
    <div class="week-nav">
      <button onclick="shiftWeek(-1)">◀ h</button>
      <span id="week-label"></span>
      <button onclick="shiftWeek(1)">l ▶</button>
    </div>
  </div>

  <!-- BODY -->
  <div id="body">
    <div id="top">

      <!-- CALENDAR -->
      <div id="cal">
        <div id="cal-grid"></div>
      </div>

      <!-- TODO -->
      <div id="todo">
        <div class="pane-hdr">TO-DO LIST</div>
        <div id="todo-list" style="padding:6px 10px;display:flex;flex-direction:column;gap:3px;overflow-y:auto;flex:1"></div>
      </div>

    </div>

    <!-- KANBAN -->
    <div id="kan">
      <div id="kan-grid">
        <div class="kan-col">
          <div class="pane-hdr">BACKLOG</div>
          <div class="kan-col-body" id="kan-backlog"></div>
        </div>
        <div class="kan-col">
          <div class="pane-hdr">DOING</div>
          <div class="kan-col-body" id="kan-doing"></div>
        </div>
        <div class="kan-col">
          <div class="pane-hdr">DONE</div>
          <div class="kan-col-body" id="kan-done"></div>
        </div>
      </div>
    </div>
  </div>

  <!-- STATUS -->
  <div id="statusbar">
    <span class="sl" id="status-left"><em>[CAL]</em> &nbsp; 0 tasks</span>
    <span id="status-right">nullcal v0.1.0</span>
  </div>
  <div id="help">
    <span><kbd>n</kbd> new</span>
    <span><kbd>e</kbd> edit selected</span>
    <span><kbd>x</kbd> toggle done</span>
    <span><kbd>del</kbd> delete</span>
    <span><kbd>|</kbd> todo split</span>
    <span><kbd>-</kbd> kanban split</span>
    <span><kbd>h/l</kbd> week</span>
  </div>

</div>

<!-- TASK MODAL -->
<div id="modal-overlay">
  <div id="modal">
    <h2 id="modal-title">NEW TASK</h2>
    <label>Title</label>
    <input id="f-title" placeholder="Task title" maxlength="100">
    <label>Description</label>
    <input id="f-desc" placeholder="Optional" maxlength="256">
    <label>Project Tag</label>
    <input id="f-tag" placeholder="wardex / vexil / ..." maxlength="30">
    <label>Due Date</label>
    <input id="f-due" placeholder="YYYY-MM-DD" maxlength="10">
    <div class="err" id="modal-err"></div>
    <div class="actions">
      <button class="primary" onclick="submitModal()">save</button>
      <button onclick="closeModal()">cancel</button>
    </div>
  </div>
</div>

<!-- DELETE CONFIRM -->
<div id="confirm-overlay">
  <div id="confirm">
    <p id="confirm-msg">Delete this task?</p>
    <div class="actions">
      <button class="danger" id="confirm-yes">delete</button>
      <button onclick="closeConfirm()">cancel</button>
    </div>
  </div>
</div>

<script>
// ── STATE ──────────────────────────────────────────────────────────────────
let ws = null;
let state = { tasks: [], routine_blocks: [] };
let selectedId = null;
let editId = null;       // null = create mode
let showTodo = true;
let showKan  = true;
let weekOffset = 0;      // weeks from current

// ── WEBSOCKET ──────────────────────────────────────────────────────────────
function connect() {
  const proto = location.protocol === 'https:' ? 'wss' : 'ws';
  ws = new WebSocket(proto + '://' + location.host + '/ws');

  ws.onopen = () => {
    document.getElementById('conn-status').textContent = '● connected';
    document.getElementById('conn-status').className = 'conn ok';
  };

  ws.onmessage = (e) => {
    const msg = JSON.parse(e.data);
    if (msg.type === 'state') {
      state = msg;
      render();
    }
  };

  ws.onclose = ws.onerror = () => {
    document.getElementById('conn-status').textContent = '● disconnected';
    document.getElementById('conn-status').className = 'conn err';
    setTimeout(connect, 2000);
  };
}

function send(msg) {
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(msg));
  }
}

// ── TIME UTILS ─────────────────────────────────────────────────────────────
function mondayOf(d) {
  const t = new Date(d);
  const day = t.getDay(); // 0=Sun
  const diff = day === 0 ? -6 : 1 - day;
  t.setDate(t.getDate() + diff);
  t.setHours(0,0,0,0);
  return t;
}

function isoWeek(d) {
  const t = new Date(Date.UTC(d.getFullYear(), d.getMonth(), d.getDate()));
  t.setUTCDate(t.getUTCDate() + 4 - (t.getUTCDay() || 7));
  const jan1 = new Date(Date.UTC(t.getUTCFullYear(), 0, 1));
  return Math.ceil((((t - jan1) / 86400000) + 1) / 7);
}

function addDays(d, n) {
  const t = new Date(d); t.setDate(t.getDate() + n); return t;
}

function sameDay(a, b) {
  return a.getFullYear()===b.getFullYear() &&
         a.getMonth()===b.getMonth() &&
         a.getDate()===b.getDate();
}

function dueClass(task) {
  if (task.status === 'done') return 't-done';
  if (!task.due_at) return 't-normal';
  const h = (new Date(task.due_at) - Date.now()) / 3600000;
  if (h < 0)  return 't-overdue';
  if (h < 48) return 't-duesoon';
  return 't-normal';
}

function fmtDate(d) {
  const dd = String(d.getDate()).padStart(2,'0');
  const mm = String(d.getMonth()+1).padStart(2,'0');
  const yy = d.getFullYear();
  return dd+'. '+mm+'. '+yy;
}

// ── RENDER ─────────────────────────────────────────────────────────────────
const SHORT_DAYS = ['SUN','MON','TUE','WED','THU','FRI','SAT'];

function render() {
  renderHeader();
  renderCalendar();
  renderTodo();
  renderKanban();
  renderStatus();
  updateToolbar();
}

function renderHeader() {
  const now = new Date();
  const monday = mondayOf(addDays(now, weekOffset * 7));
  const wn = isoWeek(monday);
  document.getElementById('hdr-meta').textContent =
    fmtDate(now) + '    ' + SHORT_DAYS[now.getDay()] + '    week ' + String(wn).padStart(2,'0');
  document.getElementById('week-label').textContent =
    'week ' + String(wn).padStart(2,'0');
}

function renderCalendar() {
  const grid = document.getElementById('cal-grid');
  grid.innerHTML = '';

  const base = mondayOf(addDays(new Date(), weekOffset * 7));
  const today = new Date(); today.setHours(0,0,0,0);

  for (let i = 0; i < 7; i++) {
    const day = addDays(base, i);
    const isToday = sameDay(day, today);

    const col = document.createElement('div');
    col.className = 'cal-col';

    const hdr = document.createElement('div');
    hdr.className = 'cal-col-hdr' + (isToday ? ' today' : '');
    hdr.textContent = SHORT_DAYS[day.getDay()] + ' ' + String(day.getDate()).padStart(2,'0');
    col.appendChild(hdr);

    const body = document.createElement('div');
    body.className = 'cal-col-body';

    // Routine blocks
    const rbs = (state.routine_blocks||[]).filter(rb => rb.weekday === day.getDay());
    rbs.forEach(rb => {
      const el = document.createElement('div');
      el.className = 'rb';
      el.innerHTML = '<div class="rb-lbl">'+esc(rb.label)+'</div>' +
                     '<div>'+esc(rb.start_time)+'-'+esc(rb.end_time)+'</div>';
      body.appendChild(el);
    });

    if (rbs.length > 0) {
      const gap = document.createElement('div'); gap.className = 'task-gap';
      body.appendChild(gap);
    }

    // Tasks for this day
    const dayTasks = (state.tasks||[]).filter(t => {
      if (!t.due_at) return false;
      return sameDay(new Date(t.due_at), day);
    });

    dayTasks.forEach(t => {
      body.appendChild(makeTaskEl(t, { prefix: t.status==='done' ? 'x' : '-', compact: true }));
    });

    col.appendChild(body);
    grid.appendChild(col);
  }
}

function renderTodo() {
  const list = document.getElementById('todo-list');
  list.innerHTML = '';
  const active = (state.tasks||[]).filter(t => t.status !== 'done');
  active.forEach(t => list.appendChild(makeTaskEl(t, { prefix: '-' })));
}

function renderKanban() {
  ['backlog','doing','done'].forEach(status => {
    const el = document.getElementById('kan-'+status);
    el.innerHTML = '';
    (state.tasks||[])
      .filter(t => t.status === status)
      .forEach(t => el.appendChild(makeTaskEl(t, {
        prefix: status==='done' ? 'x' : '-',
        showTag: true
      })));
  });
}

function makeTaskEl(t, opts) {
  const el = document.createElement('div');
  el.className = 'ti ' + dueClass(t) + (t.id===selectedId ? ' selected' : '');
  el.dataset.id = t.id;

  const pfx = document.createElement('span');
  pfx.className = 'pfx';
  pfx.textContent = t.id===selectedId ? '>' : (opts.prefix||'-');

  const lbl = document.createElement('span');
  lbl.className = 'lbl';
  lbl.textContent = t.title;

  el.appendChild(pfx);
  el.appendChild(lbl);

  if (opts.showTag && t.project_tag) {
    const tag = document.createElement('span');
    tag.className = 'tag';
    tag.textContent = t.project_tag;
    el.appendChild(tag);
  }

  el.onclick = (e) => { e.stopPropagation(); selectTask(t.id); };
  el.ondblclick = (e) => { e.stopPropagation(); openEdit(t.id); };
  return el;
}

function renderStatus() {
  const n = (state.tasks||[]).length;
  let splits = '';
  if (showTodo)  splits += ' |TODO';
  if (showKan)   splits += ' -KAN';
  document.getElementById('status-left').innerHTML =
    '<em>[CAL'+splits+']</em> &nbsp; '+n+' tasks';
}

function updateToolbar() {
  document.getElementById('btn-todo').classList.toggle('active', showTodo);
  document.getElementById('btn-kan').classList.toggle('active', showKan);
  document.getElementById('todo').classList.toggle('hidden', !showTodo);
  document.getElementById('kan').classList.toggle('hidden', !showKan);
}

// ── SELECTION ──────────────────────────────────────────────────────────────
function selectTask(id) {
  selectedId = selectedId === id ? null : id;
  render();
}

function selectedTask() {
  return (state.tasks||[]).find(t => t.id === selectedId) || null;
}

// ── WEEK NAV ───────────────────────────────────────────────────────────────
function shiftWeek(d) {
  weekOffset += d;
  render();
}

// ── LAYOUT TOGGLES ─────────────────────────────────────────────────────────
function toggleTodo() { showTodo = !showTodo; render(); }
function toggleKan()  { showKan  = !showKan;  render(); }

// ── MODAL ──────────────────────────────────────────────────────────────────
function openCreate() {
  editId = null;
  document.getElementById('modal-title').textContent = 'NEW TASK';
  document.getElementById('f-title').value = '';
  document.getElementById('f-desc').value  = '';
  document.getElementById('f-tag').value   = '';
  document.getElementById('f-due').value   = '';
  document.getElementById('modal-err').textContent = '';
  document.getElementById('modal-overlay').classList.add('open');
  setTimeout(() => document.getElementById('f-title').focus(), 50);
}

function openEdit(id) {
  const t = (state.tasks||[]).find(x => x.id===id);
  if (!t) return;
  editId = id;
  document.getElementById('modal-title').textContent = 'EDIT TASK';
  document.getElementById('f-title').value = t.title;
  document.getElementById('f-desc').value  = t.description || '';
  document.getElementById('f-tag').value   = t.project_tag || '';
  document.getElementById('f-due').value   = t.due_at ? t.due_at.slice(0,10) : '';
  document.getElementById('modal-err').textContent = '';
  document.getElementById('modal-overlay').classList.add('open');
  setTimeout(() => document.getElementById('f-title').focus(), 50);
}

function closeModal() {
  document.getElementById('modal-overlay').classList.remove('open');
}

function submitModal() {
  const title = document.getElementById('f-title').value.trim();
  const due   = document.getElementById('f-due').value.trim();
  const errEl = document.getElementById('modal-err');

  if (!title) { errEl.textContent = '! title is required'; return; }
  if (due && !/^\d{4}-\d{2}-\d{2}$/.test(due)) {
    errEl.textContent = '! due date must be YYYY-MM-DD'; return;
  }

  const task = {
    id:          editId || '',
    title,
    description: document.getElementById('f-desc').value.trim(),
    project_tag: document.getElementById('f-tag').value.trim(),
    due_at:      due || null,
  };

  send({ type: editId ? 'update' : 'create', task });
  closeModal();
}

// ── DELETE ─────────────────────────────────────────────────────────────────
let pendingDeleteId = null;

function openConfirm(id) {
  const t = (state.tasks||[]).find(x => x.id===id);
  if (!t) return;
  pendingDeleteId = id;
  document.getElementById('confirm-msg').textContent = 'Delete "'+t.title+'"?';
  document.getElementById('confirm-overlay').classList.add('open');
}

function closeConfirm() {
  document.getElementById('confirm-overlay').classList.remove('open');
  pendingDeleteId = null;
}

document.getElementById('confirm-yes').onclick = () => {
  if (pendingDeleteId) {
    if (pendingDeleteId === selectedId) selectedId = null;
    send({ type: 'delete', id: pendingDeleteId });
  }
  closeConfirm();
};

// ── KEYBOARD ───────────────────────────────────────────────────────────────
document.addEventListener('keydown', e => {
  // Modal open — only esc
  if (document.getElementById('modal-overlay').classList.contains('open')) {
    if (e.key === 'Escape') closeModal();
    return;
  }
  if (document.getElementById('confirm-overlay').classList.contains('open')) {
    if (e.key === 'Escape') closeConfirm();
    return;
  }

  switch(e.key) {
    case 'n': openCreate(); break;
    case 'e': { const t = selectedTask(); if(t) openEdit(t.id); break; }
    case 'x': {
      const t = selectedTask();
      if (t) send({ type:'setStatus', id:t.id, status: t.status==='done' ? 'backlog' : 'done' });
      break;
    }
    case 'm': case 'Enter': {
      const t = selectedTask();
      if (t) {
        const next = {backlog:'doing',doing:'done',done:'backlog'}[t.status]||'backlog';
        send({ type:'setStatus', id:t.id, status:next });
      }
      break;
    }
    case 'Delete': case 'D': {
      const t = selectedTask(); if(t) openConfirm(t.id); break;
    }
    case '|': toggleTodo(); break;
    case '-': toggleKan();  break;
    case 'h': shiftWeek(-1); break;
    case 'l': shiftWeek(1);  break;
  }
});

// Close modal on overlay click
document.getElementById('modal-overlay').onclick = e => {
  if(e.target===document.getElementById('modal-overlay')) closeModal();
};
document.getElementById('modal').onkeydown = e => {
  if(e.key==='Enter') submitModal();
};

// ── UTILS ──────────────────────────────────────────────────────────────────
function esc(s) {
  return String(s)
    .replace(/&/g,'&amp;')
    .replace(/</g,'&lt;')
    .replace(/>/g,'&gt;');
}

// ── BOOT ───────────────────────────────────────────────────────────────────
connect();
render();
</script>
</body>
</html>`

// iconSVG is the web app's favicon and logo asset.
const iconSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100" fill="none">
  <!-- Outer frame — the terminal/calendar window -->
  <rect x="8" y="8" width="84" height="84" rx="3"
        stroke="#e0e0e0" stroke-width="6"/>

  <!-- Header rule — the top bar of every calendar column header -->
  <line x1="8" y1="29" x2="92" y2="29"
        stroke="#e0e0e0" stroke-width="6"/>

  <!-- Slash — the null. Cuts from bottom-left to top-right of the inner body -->
  <line x1="19" y1="87" x2="81" y2="13"
        stroke="#e0e0e0" stroke-width="9" stroke-linecap="square"/>
</svg>`
