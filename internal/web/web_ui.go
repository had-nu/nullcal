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
#cal{flex:1.5;border-right:1px solid var(--border);display:flex;flex-direction:column;overflow:hidden;min-height:0}
.cal-row{display:grid;grid-template-columns:44px repeat(7,1fr)}
.cal-col-hdr{padding:4px 6px;font-size:12px;font-weight:bold;letter-spacing:.05em;border-bottom:1px solid var(--border);border-right:1px solid var(--border);white-space:nowrap;overflow:hidden;text-align:center}
.cal-col-hdr.today{color:var(--accent);background:#1c1c1c}
.cal-allday-cell{border-right:1px solid var(--border);border-bottom:1px solid var(--border);min-height:24px;display:flex;flex-direction:column;gap:1px;padding:2px;background:var(--bg2)}
.cal-col{border-right:1px solid var(--border);position:relative;height:1152px} /* 24 * 48px */
.cal-col:last-child{border-right:none}
#cal-scroll{flex:1;overflow-y:auto;overflow-x:hidden;background:var(--bg)}
.cal-time-col{border-right:1px solid var(--border);position:relative;height:1152px}
.cal-time-lbl{position:absolute;width:100%;text-align:right;padding-right:4px;font-size:10px;color:var(--dim);transform:translateY(-50%)}
.cal-grid-line{position:absolute;width:100%;height:1px;background:var(--border);opacity:0.5}

.abs-ev{position:absolute;width:calc(100% - 4px);left:2px;border-radius:2px;overflow:hidden;font-size:11px;padding:2px 4px;z-index:10;border-left:2px solid #4285f4;background:rgba(66,133,244,.15);color:#ccc;box-shadow:0 1px 2px rgba(0,0,0,.3)}
.abs-ti{position:absolute;width:calc(100% - 4px);left:2px;border-radius:2px;overflow:hidden;font-size:11.5px;padding:2px 4px;z-index:11;background:var(--bg3);border:1px solid var(--border);color:var(--fg);box-shadow:0 1px 2px rgba(0,0,0,.3);cursor:pointer;display:flex;align-items:flex-start;gap:4px}
.abs-ti:hover{border-color:var(--accent)}
.abs-ti.selected{background:var(--sel-bg);color:var(--sel-fg) !important;border-color:var(--sel-fg)}
.abs-ti .pfx{opacity:0.6;flex-shrink:0}

.rb{font-size:11px;color:#666;padding:1px 4px;border-left:2px solid #444;background:rgba(255,255,255,.02);margin-top:1px}
.rb-lbl{font-weight:bold;color:#777}

/* ── TODO ── */
#todo{width:220px;min-width:150px;max-width:400px;display:flex;flex-direction:column;overflow:hidden;min-height:0}
#todo.hidden{display:none}
#todo-resizer{width:4px;cursor:col-resize;background:var(--bg2);border-right:1px solid var(--border);border-left:1px solid var(--border);transition:background .15s;flex-shrink:0}
#todo-resizer:hover, #todo-resizer.active{background:var(--border)}
#todo-resizer.hidden{display:none}

/* ── DETAILS (INSPECTOR) ── */
#details{flex:0 0 320px;border-left:1px solid var(--border);display:flex;flex-direction:column;overflow:hidden;background:var(--bg2)}
#details.hidden{display:none}
#details-body{padding:14px;display:flex;flex-direction:column;gap:12px;overflow-y:auto;flex:1}
#details-body label{font-size:11.5px;color:var(--dim);font-weight:bold;letter-spacing:.05em;margin-bottom:-4px}
#details-body input, #details-body textarea{
  background:transparent;border:1px solid var(--border);color:var(--fg);
  font-family:var(--font);font-size:13px;padding:6px 8px;outline:none;
  transition:border-color .15s;width:100%;
}
#details-body textarea{resize:vertical;min-height:120px;flex:1}
#details-body input:focus, #details-body textarea:focus{border-color:var(--accent)}
#details-meta{display:flex;flex-direction:column;gap:10px}
.details-row{display:flex;gap:10px;width:100%}

/* ── RESIZER ── */
#resizer{height:6px;background:var(--bg2);border-bottom:1px solid var(--border);cursor:row-resize;flex-shrink:0;transition:background .15s}
#resizer:hover, #resizer.active{background:var(--border)}

/* ── KANBAN ── */
#kan{height:300px;flex:0 0 auto;display:flex;flex-direction:column;overflow:hidden}
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
.ti.selected{background:var(--sel-bg);color:var(--sel-fg) !important}
.ti.t-normal{color:var(--fg)}
.ti.t-duesoon{color:var(--duesoon)}
.ti.t-overdue{color:var(--overdue)}
.ti.t-done{color:var(--done);text-decoration:line-through}
.ti .pfx{flex-shrink:0;width:22px;color:inherit;opacity:.6}
.ti .lbl{overflow:hidden;text-overflow:ellipsis;flex:1}
.ti .tag{font-size:12px;background:#222;color:var(--dim);padding:0 4px;flex-shrink:0}

/* ── STATUS + HELP ── */
#statusbar{flex-shrink:0;padding:3px 14px;font-size:13px;background:#141414;border-top:1px solid var(--border);display:flex;justify-content:space-between;align-items:center}
#statusbar .sl{color:var(--fg)}
#statusbar .sl em{color:var(--accent);font-style:normal}
#statusbar .sr{display:flex;align-items:center;gap:12px;color:var(--dim)}
.icon-link{display:inline-flex;align-items:center;color:var(--dim);text-decoration:none;opacity:.55;transition:opacity .15s}
.icon-link:hover{opacity:1;color:var(--fg)}
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
        <div id="cal-hdr" class="cal-row"></div>
        <div id="cal-allday" class="cal-row"></div>
        <div id="cal-scroll">
          <div id="cal-body" class="cal-row"></div>
        </div>
      </div>

      <!-- TODO -->
      <div id="todo">
        <div class="pane-hdr">TO-DO LIST</div>
        <div id="todo-list" style="padding:6px 10px;display:flex;flex-direction:column;gap:3px;overflow-y:auto;flex:1"></div>
      </div>

      <!-- TODO RESIZER -->
      <div id="todo-resizer"></div>

      <!-- DETAILS (INSPECTOR) -->
      <div id="details" class="hidden">
        <div class="pane-hdr" style="display:flex;justify-content:space-between">
          <span>DETAILS</span>
          <button style="background:transparent;border:none;color:var(--dim);cursor:pointer;font-family:inherit" onclick="selectTask(null)">[esc] close</button>
        </div>
        <div id="details-body">
          <label>TITLE</label>
          <input id="dt-title" onblur="saveDetails()">
          
          <div id="details-meta">
            <div class="details-row">
              <div style="flex:1">
                <label>DUE TO</label>
                <input id="dt-due" type="date" style="margin-top:4px" onblur="saveDetails()">
              </div>
              <div style="flex:1">
                <label>TIME</label>
                <input id="dt-time" type="time" style="margin-top:4px" onblur="saveDetails()">
              </div>
            </div>
            <div class="details-row">
              <div style="flex:1">
                <label>TAG</label>
                <input id="dt-tag" style="margin-top:4px" placeholder="..." onblur="saveDetails()">
              </div>
              <div style="flex:1">
                <label>RECURRENCE</label>
                <select id="dt-recur" style="margin-top:4px;width:100%;background:transparent;border:1px solid var(--border);color:var(--fg);font-family:var(--font);font-size:13px;padding:6px 8px;outline:none;cursor:pointer" onchange="saveDetails()">
                  <option value="">None</option>
                  <option value="daily">Daily</option>
                  <option value="weekly">Weekly</option>
                  <option value="monthly">Monthly</option>
                </select>
              </div>
              <div style="flex:0.5">
                <label>POMODOROS</label>
                <input id="dt-pomodoros" type="number" min="1" max="20" style="margin-top:4px;text-align:center" onblur="saveDetails()">
              </div>
            </div>
          </div>

          <label>DESCRIPTION</label>
          <textarea id="dt-desc" placeholder="Add task context or notes here..." onblur="saveDetails()"></textarea>
        </div>
      </div>

    </div>

    <!-- RESIZER -->
    <div id="resizer"></div>

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
    <span class="sr">
      <span id="status-right">nullcal v0.1.0</span>
      <a class="icon-link" href="https://github.com/had-nu/lazy.go" target="_blank" title="Built with lazy.go">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <rect x="2" y="2" width="20" height="20" rx="3" stroke="currentColor" stroke-width="2"/>
          <text x="12" y="16" text-anchor="middle" font-size="10" font-family="monospace" fill="currentColor" font-weight="bold">go</text>
        </svg>
      </a>
      <a class="icon-link" href="https://github.com/had-nu/nullcal" target="_blank" title="View on GitHub">
        <svg width="14" height="14" viewBox="0 0 98 96" xmlns="http://www.w3.org/2000/svg">
          <path fill-rule="evenodd" clip-rule="evenodd" d="M48.854 0C21.839 0 0 22 0 49.217c0 21.756 13.993 40.172 33.405 46.69 2.427.49 3.316-1.059 3.316-2.362 0-1.141-.08-5.052-.08-9.127-13.59 2.934-16.42-5.867-16.42-5.867-2.184-5.704-5.42-7.17-5.42-7.17-4.448-3.015.324-3.015.324-3.015 4.934.326 7.523 5.052 7.523 5.052 4.367 7.496 11.404 5.378 14.235 4.074.404-3.178 1.699-5.378 3.074-6.6-10.839-1.141-22.243-5.378-22.243-24.283 0-5.378 1.94-9.778 5.014-13.2-.485-1.222-2.184-6.275.486-13.038 0 0 4.125-1.304 13.426 5.052a46.97 46.97 0 0 1 12.214-1.63c4.125 0 8.33.571 12.213 1.63 9.302-6.356 13.427-5.052 13.427-5.052 2.67 6.763.97 11.816.485 13.038 3.155 3.422 5.015 7.822 5.015 13.2 0 18.905-11.404 23.06-22.324 24.283 1.78 1.548 3.316 4.481 3.316 9.126 0 6.6-.08 11.897-.08 13.526 0 1.304.89 2.853 3.316 2.364 19.412-6.52 33.405-24.935 33.405-46.691C97.707 22 75.788 0 48.854 0z" fill="currentColor"/>
        </svg>
      </a>
    </span>
  </div>
  <div id="help">
    <span><kbd>n</kbd> new</span>
    <span><kbd>e</kbd> edit</span>
    <span><kbd>j/k</kbd> up/down</span>
    <span><kbd>t</kbd> today</span>
    <span><kbd>b</kbd> backlog</span>
    <span><kbd>d</kbd> doing</span>
    <span><kbd>x</kbd> done</span>
    <span><kbd>m</kbd> details</span>
    <span><kbd>del</kbd> delete</span>
    <span><kbd>|</kbd> todo split</span>
    <span><kbd>-</kbd> kanban split</span>
    <span><kbd>z</kbd> zen mode</span>
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
    <div style="display:flex;gap:8px">
      <div style="flex:1">
        <label>Project Tag</label>
        <input id="f-tag" placeholder="wardex / ..." maxlength="30">
      </div>
      <div style="flex:0.6">
        <label>Pomodoros</label>
        <input id="f-pomodoros" type="number" min="1" max="20" value="1" style="text-align:center">
      </div>
    </div>
    <div style="display:flex;gap:8px">
      <div style="flex:1">
        <label>Due Date</label>
        <input id="f-due" type="date">
      </div>
      <div style="flex:1">
        <label>Time</label>
        <input id="f-time" type="time">
      </div>
      <div style="flex:1">
        <label>Recurrence</label>
        <select id="f-recur" style="display:block;width:100%;background:var(--bg);border:1px solid var(--border);color:var(--fg);font-family:var(--font);font-size:14px;padding:4px 8px;outline:none;cursor:pointer">
          <option value="">None</option>
          <option value="daily">Daily</option>
          <option value="weekly">Weekly</option>
          <option value="monthly">Monthly</option>
        </select>
      </div>
    </div>
    <label style="display:flex;align-items:center;gap:8px;cursor:pointer;margin-top:4px">
      <input type="checkbox" id="f-backlog" style="accent-color:var(--accent);width:14px;height:14px">
      <span style="font-size:13px;color:var(--dim)">Add directly to Kanban backlog</span>
    </label>
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
let state = { tasks: [], routine_blocks: [], calendar_events: [] };
let selectedId = null;
let editId = null;       // null = create mode
let showTodo = true;
let showKan  = true;
let zenMode  = false;
let showAllDone = false;
let weekOffset = 0;      // weeks from current

let pomoTimerId = null;
let pomoEnd = 0;
let pomoTaskId = null;

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

const SHORT_MONTHS = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'];

function fmtDate(d) {
  const dd = String(d.getDate()).padStart(2,'0');
  const month = SHORT_MONTHS[d.getMonth()];
  const yy = d.getFullYear();
  return dd+' '+month+' '+yy;
}

function fmtTime(d) {
  let h = d.getHours();
  const m = String(d.getMinutes()).padStart(2,'0');
  return String(h).padStart(2,'0') + ':' + m;
}

// ── RENDER ─────────────────────────────────────────────────────────────────
const SHORT_DAYS = ['SUN','MON','TUE','WED','THU','FRI','SAT'];

function render() {
  renderHeader();
  renderCalendar();
  renderTodo();
  renderKanban();
  renderStatus();
  renderDetails();
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
  const hdrRow = document.getElementById('cal-hdr');
  const allRow = document.getElementById('cal-allday');
  const bdyRow = document.getElementById('cal-body');
  
  hdrRow.innerHTML = '<div class="cal-col-hdr" style="font-size:10px;color:var(--dim);border-right:1px solid var(--border)">GMT</div>';
  allRow.innerHTML = '<div class="cal-allday-cell" style="font-size:10px;color:var(--dim);justify-content:center;align-items:center">ALL-DAY</div>';
  
  // Build Time Sidebar
  let timeHtml = '<div class="cal-time-col">';
  for (let i = 0; i < 24; i++) {
    timeHtml += '<div class="cal-time-lbl" style="top:'+(i*48)+'px">'+String(i).padStart(2,'0')+':00</div>';
  }
  timeHtml += '</div>';
  bdyRow.innerHTML = timeHtml;

  const base = mondayOf(addDays(new Date(), weekOffset * 7));
  const today = new Date(); today.setHours(0,0,0,0);

  for (let i = 0; i < 7; i++) {
    const day = addDays(base, i);
    const isToday = sameDay(day, today);
    const dayStart = day.getTime();
    const dayEnd = dayStart + 86400000;

    // Header Col
    const hdr = document.createElement('div');
    hdr.className = 'cal-col-hdr' + (isToday ? ' today' : '');
    hdr.textContent = SHORT_DAYS[day.getDay()] + ' ' + String(day.getDate()).padStart(2,'0');
    hdrRow.appendChild(hdr);

    // All-Day Col
    const allC = document.createElement('div');
    allC.className = 'cal-allday-cell';
    allRow.appendChild(allC);

    // Body Col
    const bdyC = document.createElement('div');
    bdyC.className = 'cal-col';
    bdyC.dataset.date = day.getTime(); // for drag and drop
    
    // Calendar Col Drop Zone
    bdyC.ondragover = (e) => { e.preventDefault(); e.dataTransfer.dropEffect = 'move'; };
    bdyC.ondrop = (e) => {
      e.preventDefault();
      const id = e.dataTransfer.getData('text/plain');
      const t = state.tasks.find(x => x.id === id);
      if (!t) return;
      
      const rect = bdyC.getBoundingClientRect();
      const y = e.clientY - rect.top;
      let hours = Math.round(y / 24) * 0.5; // snap to 30 mins
      if (hours < 0) hours = 0;
      if (hours >= 24) hours = 23.5;
      
      const newD = new Date(parseInt(bdyC.dataset.date, 10));
      newD.setHours(Math.floor(hours), (hours % 1) * 60, 0, 0);
      
      // Keep ISO string local format YYYY-MM-DDTHH:MM:SS
      const isoY = newD.getFullYear();
      const isoM = String(newD.getMonth()+1).padStart(2,'0');
      const isoD = String(newD.getDate()).padStart(2,'0');
      const isoh = String(newD.getHours()).padStart(2,'0');
      const isom = String(newD.getMinutes()).padStart(2,'0');
      const newDue = isoY+'-'+isoM+'-'+isoD+'T'+isoh+':'+isom+':00';
      
      send({ type:'update', task: { ...t, due_at: newDue } });
    };

    // Draw background grid lines
    for (let h = 0; h < 24; h++) {
      const gl = document.createElement('div');
      gl.className = 'cal-grid-line';
      gl.style.top = (h*48) + 'px';
      bdyC.appendChild(gl);
    }

    // 1. Routine Blocks
    const rbs = (state.routine_blocks||[]).filter(rb => rb.weekday === day.getDay());
    rbs.forEach(rb => {
      // parse start_time HH:MM
      const [sh, sm] = rb.start_time.split(':').map(Number);
      const [eh, em] = rb.end_time.split(':').map(Number);
      const top = (sh + sm/60) * 48;
      const h   = ((eh + em/60) - (sh + sm/60)) * 48;
      
      const el = document.createElement('div');
      el.className = 'abs-ev rb';
      el.style.top = top + 'px';
      el.style.minHeight = Math.max(h, 15) + 'px';
      el.innerHTML = '<span class="rb-lbl">'+esc(rb.label)+'</span>';
      bdyC.appendChild(el);
    });

    // 2. Local Tasks (Filter out done tasks so they don't clog the calendar)
    const dayTasks = (state.tasks||[]).filter(t => t.due_at && t.status !== 'done' && sameDay(new Date(t.due_at), day));
    dayTasks.forEach(t => {
      const dT = new Date(t.due_at);
      if (t.due_at.length <= 10 || (dT.getHours()===0 && dT.getMinutes()===0)) {
        // all-day task
        allC.appendChild(makeTaskEl(t, { prefix: t.status==='done'?'[x]':'[ ]', compact:true, noabs:true }));
      } else {
        const top = (dT.getHours() + dT.getMinutes()/60) * 48;
        const el = makeTaskEl(t, { prefix: t.status==='done'?'[x]':'[ ]', compact:true });
        el.className = el.className.replace('ti','abs-ti'); // swap class
        el.style.top = top + 'px';
        const pomodoros = t.pomodoros && t.pomodoros > 0 ? t.pomodoros : 1;
        el.style.minHeight = Math.max(pomodoros * 24, 24) + 'px'; // 1 pomodoro = 30m = 24px
        el.draggable = true;
        el.ondragstart = (e) => {
          e.dataTransfer.setData('text/plain', t.id);
          e.dataTransfer.effectAllowed = 'move';
        };
        
        // Resize handle
        const handle = document.createElement('div');
        handle.style.position = 'absolute';
        handle.style.bottom = '0';
        handle.style.left = '0';
        handle.style.right = '0';
        handle.style.height = '6px';
        handle.style.cursor = 'ns-resize';
        handle.onmousedown = (e) => {
          e.stopPropagation();
          startTaskResize(e, t.id, el);
        };
        el.appendChild(handle);
        
        bdyC.appendChild(el);
      }
    });

    // 3. GCal Events
    const dayEvents = (state.calendar_events||[]).filter(ev => {
      if (!ev.start_at) return false;
      const dStart = new Date(ev.start_at);
      const dEnd   = new Date(ev.end_at || ev.start_at);
      if (dStart.getTime() === dEnd.getTime()) {
         return dStart.getTime() >= dayStart && dStart.getTime() < dayEnd;
      }
      return dStart.getTime() < dayEnd && dEnd.getTime() > dayStart;
    });

    dayEvents.forEach(ev => {
      const dStart = new Date(ev.start_at);
      const dEnd   = new Date(ev.end_at || ev.start_at);
      const allDay = dStart.getHours()===0 && dStart.getMinutes()===0 &&
                     dEnd.getHours()===0   && dEnd.getMinutes()===0;
                     
      if (allDay) {
        const el = document.createElement('div');
        el.className = 'abs-ev';
        el.style.position = 'relative'; // reset to flex flow in allday col
        el.style.width = '100%';
        el.innerHTML = esc(ev.title);
        el.title = ev.description || ev.title;
        allC.appendChild(el);
      } else {
        // limit bounds to current day (starts prev day -> 00:00, ends next day -> 24:00)
        let sh = dStart.getTime() < dayStart ? 0 : dStart.getHours() + dStart.getMinutes()/60;
        let eh = dEnd.getTime() > dayEnd ? 24 : dEnd.getHours() + dEnd.getMinutes()/60;
        const top = sh * 48;
        const h = (eh - sh) * 48;
        
        const el = document.createElement('div');
        el.className = 'abs-ev';
        el.style.top = top + 'px';
        el.style.minHeight = Math.max(h, 20) + 'px';
        el.innerHTML = esc(ev.title);
        el.title = ev.description || ev.title;
        bdyC.appendChild(el);
      }
    });

    bdyRow.appendChild(bdyC);
  }
}

function renderTodo() {
  const list = document.getElementById('todo-list');
  list.innerHTML = '';
  // Only show tasks with status 'todo' — backlog tasks live in kanban only.
  const active = (state.tasks||[]).filter(t => t.status === 'todo');
  active.forEach(t => list.appendChild(makeTaskEl(t, { prefix: '[ ]' })));
}

function renderKanban() {
  // backlog column: tasks with status 'backlog' only (not 'todo')
  ['backlog','doing','done'].forEach(status => {
    const el = document.getElementById('kan-'+status);
    el.innerHTML = '';
    let items = (state.tasks||[]).filter(t => t.status === status);
    
    if (status === 'done') {
      const total = items.length;
      if (!showAllDone && total > 10) {
        // Assume the last 10 in the list are the most relevant (default server sorting is due_at ASC)
        items = items.slice(total - 10);
      }
      items.forEach(t => el.appendChild(makeTaskEl(t, {
        prefix: '[x]',
        showTag: true
      })));
      
      if (total > 10) {
        const btn = document.createElement('div');
        btn.style = "text-align:center;font-size:11px;color:var(--dim);cursor:pointer;padding:8px 4px;margin-top:auto;border-top:1px dashed var(--border)";
        btn.textContent = showAllDone ? "Show less" : "Show all " + total + " completed tasks...";
        btn.onclick = () => { showAllDone = !showAllDone; renderKanban(); };
        el.appendChild(btn);
      }
    } else {
      items.forEach(t => el.appendChild(makeTaskEl(t, {
        prefix: '[ ]',
        showTag: true
      })));
    }
  });
}

function makeTaskEl(t, opts) {
  const el = document.createElement('div');
  // The 'ti' class is for general task items. 'abs-task' is added by renderCalendar for timed events.
  // If opts.noabs is true, it means it's an all-day task in the calendar or a task in todo/kanban,
  // so it should not get the 'abs-task' class by default here.
  let baseClass = opts.noabs ? 'ti ' : 'ti ';
  el.className = baseClass + dueClass(t) + (t.id===selectedId ? ' selected' : '');
  el.dataset.id = t.id;

  const pfx = document.createElement('span');
  pfx.className = 'pfx';
  pfx.textContent = t.id===selectedId ? '>' : (opts.prefix||'[ ]');
  if (opts.prefix === '[x]') pfx.style.opacity = '0.5';

  const lbl = document.createElement('span');
  lbl.className = 'lbl';
  const recIcon = (t.recurrence && t.recurrence !== 'none') ? '<span style="font-size:10px;margin-right:2px;vertical-align:middle;color:var(--accent)">↻</span>' : '';
  const pomIcon = (t.pomodoros && t.pomodoros > 1) ? '<span style="font-size:10px;margin-right:4px;vertical-align:middle;color:var(--overdue)">🍅' + t.pomodoros + '</span>' : '';
  lbl.innerHTML = pomIcon + recIcon + esc(t.title);

  el.appendChild(pfx);
  el.appendChild(lbl);

  if (t.due_at) {
    const d = new Date(t.due_at);
    const hasTime = d.getHours() !== 0 || d.getMinutes() !== 0;
    if (hasTime) {
      const badge = document.createElement('span');
      badge.className = 'time-badge';
      badge.textContent = fmtTime(d);
      el.appendChild(badge);
    }
  }

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

  if (pomoTimerId) {
    const left = Math.max(0, pomoEnd - Date.now());
    const m = Math.floor(left / 60000);
    const s = Math.floor((left % 60000) / 1000);
    const t = state.tasks.find(x => x.id === pomoTaskId);
    document.getElementById('status-left').innerHTML =
      '<em style="color:var(--overdue)">[ 🍅 '+String(m).padStart(2,'0')+':'+String(s).padStart(2,'0')+' ]</em> &nbsp; '+esc(t ? t.title : '');
  } else {
    document.getElementById('status-left').innerHTML =
      '<em>[CAL'+splits+']</em> &nbsp; '+n+' tasks';
  }
}

function updateToolbar() {
  document.getElementById('btn-todo').classList.toggle('active', showTodo);
  document.getElementById('btn-kan').classList.toggle('active', showKan);
  // Zen Mode overrides
  if (zenMode) {
    document.getElementById('todo').classList.add('hidden');
    document.getElementById('todo-resizer').classList.add('hidden');
    document.getElementById('kan').classList.add('hidden');
    document.getElementById('resizer').style.display = 'none';
  } else {
    document.getElementById('todo').classList.toggle('hidden', !showTodo);
    document.getElementById('todo-resizer').classList.toggle('hidden', !showTodo);
    document.getElementById('kan').classList.toggle('hidden', !showKan);
    document.getElementById('resizer').style.display = showKan ? 'block' : 'none';
  }
}

// ── SELECTION ──────────────────────────────────────────────────────────────
function selectTask(id) {
  selectedId = selectedId === id ? null : id;
  render();

  // If selecting a task, populate details and focus description if it's empty
  if (selectedId) {
    const t = selectedTask();
    if (t) {
      document.getElementById('dt-title').value = t.title || '';
      
      let dueD = '';
      let dueT = '';
      if (t.due_at) {
        dueD = t.due_at.slice(0, 10);
        if (t.due_at.length > 10) {
          const d = new Date(t.due_at);
          dueT = String(d.getHours()).padStart(2,'0') + ':' + String(d.getMinutes()).padStart(2,'0');
          if (dueT === '00:00') dueT = '';
        }
      }
      document.getElementById('dt-due').value = dueD;
      document.getElementById('dt-time').value = dueT;

      document.getElementById('dt-tag').value = t.project_tag || '';
      document.getElementById('dt-recur').value = t.recurrence || '';
      document.getElementById('dt-pomodoros').value = t.pomodoros || 1;
      document.getElementById('dt-desc').value = t.description || '';
      
      // Removed the document.getElementById('details').classList.remove('hidden') auto-popup
      // It must be triggered explicitly with 'm'
    }
  }
}

function selectedTask() {
  if (!selectedId) return null;
  return state.tasks.find(x => x.id === selectedId) || null;
}

// ── DETAILS (INSPECTOR) ────────────────────────────────────────────────────
function renderDetails() {
  const dEl = document.getElementById('details');
  const t = selectedTask();
  if (!t || zenMode) {
    dEl.classList.add('hidden');
    return;
  }
  dEl.classList.remove('hidden');
  
  // We only update the visual state attributes here without touching values
  // so we don't overwrite user's typing focus during real-time broadcasts.
  // The actual values are populated during selectTask().
}

function saveDetails() {
  if (!selectedId) return;
  const t = state.tasks.find(x => x.id === selectedId);
  if (!t) return;

  let nTitle = document.getElementById('dt-title').value.trim();
  const nDesc  = document.getElementById('dt-desc').value.trim();
  let nTag     = document.getElementById('dt-tag').value.trim();
  const nDueD  = document.getElementById('dt-due').value.trim();
  const nDueT  = document.getElementById('dt-time').value.trim();
  const nRecur = document.getElementById('dt-recur').value;
  const nPomo  = parseInt(document.getElementById('dt-pomodoros').value, 10) || 1;

  // Check for slash command in title
  const match = nTitle.match(/ \/(\w+)$/);
  if (match) {
    nTag = match[1];
    nTitle = nTitle.replace(match[0], '');
    document.getElementById('dt-title').value = nTitle;
    document.getElementById('dt-tag').value = nTag;
  }

  let nDue = null;
  if (nDueD && /^\d{4}-\d{2}-\d{2}$/.test(nDueD)) {
    nDue = nDueT ? nDueD + 'T' + nDueT + ':00' : nDueD;
  }

  // Restore task state so optimistic rendering isn't weird if it fails
  if (!nTitle) {
    document.getElementById('dt-title').value = t.title;
    return;
  }

  if (nTitle!==t.title || nDesc!==(t.description||'') || nTag!==(t.project_tag||'') || nDue!==(t.due_at||null) || nRecur!==(t.recurrence||'') || nPomo!==(t.pomodoros||1)) {
    const fresh = { ...t, title:nTitle, description:nDesc, project_tag:nTag, due_at:nDue, recurrence:nRecur, pomodoros:nPomo };
    send({ type:'update', task:fresh });
  }
}

// ── WEEK NAV ───────────────────────────────────────────────────────────────
function shiftWeek(n) {
  weekOffset += n;
  render();
}

// ── LAYOUT TOGGLES ─────────────────────────────────────────────────────────
function toggleTodo() { showTodo = !showTodo; if(!zenMode) render(); }
function toggleKan()  { showKan  = !showKan;  if(!zenMode) render(); }
function toggleZen()  { zenMode = !zenMode; render(); }

// ── MODAL ──────────────────────────────────────────────────────────────────
function openCreate() {
  editId = null;
  document.getElementById('modal-title').textContent = 'NEW TASK';
  document.getElementById('f-title').value    = '';
  document.getElementById('f-desc').value     = '';
  document.getElementById('f-tag').value      = '';
  document.getElementById('f-due').value      = '';
  document.getElementById('f-time').value     = '';
  document.getElementById('f-recur').value    = '';
  document.getElementById('f-pomodoros').value= 1;
  document.getElementById('f-backlog').checked = false;
  document.getElementById('modal-err').textContent = '';
  document.getElementById('modal-overlay').classList.add('open');
  setTimeout(() => document.getElementById('f-title').focus(), 50);
}

function openEdit(id) {
  const t = (state.tasks||[]).find(x => x.id===id);
  if (!t) return;
  editId = id;
  document.getElementById('modal-title').textContent = 'EDIT TASK';
  document.getElementById('f-title').value    = t.title;
  document.getElementById('f-desc').value     = t.description || '';
  document.getElementById('f-tag').value      = t.project_tag || '';
  const dueDate = t.due_at ? t.due_at.slice(0,10) : '';
  document.getElementById('f-due').value      = dueDate;
  // Restore time portion if present
  if (t.due_at && t.due_at.length > 10) {
    const d = new Date(t.due_at);
    const hhmm = String(d.getHours()).padStart(2,'0') + ':' + String(d.getMinutes()).padStart(2,'0');
    document.getElementById('f-time').value = (hhmm === '00:00') ? '' : hhmm;
  } else {
    document.getElementById('f-time').value = '';
  }
  document.getElementById('f-recur').value     = t.recurrence || '';
  document.getElementById('f-pomodoros').value = t.pomodoros || 1;
  document.getElementById('f-backlog').checked = t.status === 'backlog';
  document.getElementById('modal-err').textContent = '';
  document.getElementById('modal-overlay').classList.add('open');
  setTimeout(() => document.getElementById('f-title').focus(), 50);
}

function closeModal() {
  document.getElementById('modal-overlay').classList.remove('open');
}

function submitModal() {
  let title   = document.getElementById('f-title').value.trim();
  const dueDate = document.getElementById('f-due').value.trim();
  const dueTime = document.getElementById('f-time').value.trim();
  const toBacklog = document.getElementById('f-backlog').checked;
  const recur   = document.getElementById('f-recur').value;
  const pomo    = parseInt(document.getElementById('f-pomodoros').value, 10) || 1;
  let tag       = document.getElementById('f-tag').value.trim();
  const errEl   = document.getElementById('modal-err');

  // Check for slash command in title
  const match = title.match(/ \/(\w+)$/);
  if (match) {
    tag = match[1];
    title = title.replace(match[0], '');
  }

  if (!title) { errEl.textContent = '! title is required'; return; }
  if (dueDate && !/^\d{4}-\d{2}-\d{2}$/.test(dueDate)) {
    errEl.textContent = '! due date must be YYYY-MM-DD'; return;
  }

  // Combine date + time into ISO string if both provided.
  let dueAt = null;
  if (dueDate) {
    dueAt = dueTime ? dueDate + 'T' + dueTime + ':00' : dueDate;
  }

  const task = {
    id:          editId || '',
    title,
    description: document.getElementById('f-desc').value.trim(),
    project_tag: tag,
    due_at:      dueAt,
    recurrence:  recur,
    pomodoros:   pomo,
    // On create: 'backlog' if checkbox; 'todo' otherwise.
    // On edit: preserve existing status (backend decides).
    status:      editId ? undefined : (toBacklog ? 'backlog' : 'todo'),
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
  // Ignore global keybinds if we are currently typing in an input/textarea
  if (['INPUT', 'TEXTAREA'].includes(document.activeElement.tagName)) {
    if (e.key === 'Escape') {
      document.activeElement.blur();
      // If we are in the details pane, also deselect task
      if (!document.getElementById('modal-overlay').classList.contains('open')) {
        selectTask(null);
      }
    }
    // Except for enter inside inputs globally
    if (e.key === 'Enter' && document.activeElement.tagName === 'INPUT' && !document.getElementById('modal-overlay').classList.contains('open')) {
      document.activeElement.blur();
    }
    return;
  }

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
    case 'p': { if (selectedId) togglePomodoro(selectedId); break; }
    case 'n': openCreate(); break;
    case 'e': { const t = selectedTask(); if(t) openEdit(t.id); break; }
    case '\\': {
      openCreate();
      document.getElementById('f-title').value = "Focus /pomodoro";
      const now = new Date();
      const end = new Date(now.getTime() + 25 * 60000);
      document.getElementById('f-due').value = end.toISOString().slice(0, 10);
      document.getElementById('f-time').value = String(end.getHours()).padStart(2,'0') + ':' + String(end.getMinutes()).padStart(2,'0');
      // small timeout to select the text so user can just type a new title if they want
      setTimeout(() => document.getElementById('f-title').select(), 50);
      break;
    }
    case 'j': {
      const els = Array.from(document.querySelectorAll('.ti'));
      if(els.length===0) break;
      const idx = els.findIndex(el => el.dataset.id === selectedId);
      if(idx===-1) selectTask(els[0].dataset.id);
      else if(idx < els.length-1) selectTask(els[idx+1].dataset.id);
      break;
    }
    case 'k': {
      const els = Array.from(document.querySelectorAll('.ti'));
      if(els.length===0) break;
      const idx = els.findIndex(el => el.dataset.id === selectedId);
      if(idx===-1) selectTask(els[els.length-1].dataset.id);
      else if(idx > 0) selectTask(els[idx-1].dataset.id);
      break;
    }
    case 'x': {
      const t = selectedTask();
      if (t) send({ type:'setStatus', id:t.id, status: t.status==='done' ? 'todo' : 'done' });
      break;
    }
    case 'd': {
      const t = selectedTask();
      if (t) send({ type:'setStatus', id:t.id, status: t.status==='doing' ? 'todo' : 'doing' });
      break;
    }
    case 'b': {
      const t = selectedTask();
      if (t) send({ type:'setStatus', id:t.id, status: t.status==='backlog' ? 'todo' : 'backlog' });
      break;
    }
    case 'm': {
      document.getElementById('details').classList.toggle('hidden');
      break;
    }
    case 'Enter': {
      const t = selectedTask();
      if (t) {
        const next = {todo:'backlog',backlog:'doing',doing:'done',done:'todo'}[t.status]||'todo';
        send({ type:'setStatus', id:t.id, status:next });
      }
      break;
    }
    case 'Delete': case 'D': {
      const t = selectedTask(); if(t) openConfirm(t.id); break;
    }
    case 't': weekOffset = 0; render(); break;
    case '|': toggleTodo(); break;
    case '-': toggleKan();  break;
    case 'z': zenMode = !zenMode; render(); break;
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

// ── KANBAN RESIZER ─────────────────────────────────────────────────────────
let isResizing = false;

document.getElementById('resizer').addEventListener('mousedown', function(e) {
  isResizing = true;
  document.body.style.cursor = 'row-resize';
  document.body.style.userSelect = 'none';
  document.getElementById('resizer').classList.add('active');
});

// ── TODO RESIZER ───────────────────────────────────────────────────────────
let isTodoResizing = false;

document.getElementById('todo-resizer').addEventListener('mousedown', function(e) {
  isTodoResizing = true;
  document.body.style.cursor = 'col-resize';
  document.body.style.userSelect = 'none';
  document.getElementById('todo-resizer').classList.add('active');
});

// ── TASK RESIZER ───────────────────────────────────────────────────────────
let resizeTaskId = null;
let resizeTaskEl = null;
let resizeStartY = 0;
let resizeStartPomo = 0;

function startTaskResize(e, id, el) {
  const t = (state.tasks||[]).find(x => x.id === id);
  if (!t) return;
  resizeTaskId = id;
  resizeTaskEl = el;
  resizeStartY = e.clientY;
  resizeStartPomo = t.pomodoros || 1;
  document.body.style.cursor = 'ns-resize';
  document.body.style.userSelect = 'none';
}

// ── GLOBAL RESIZE LISTENER ─────────────────────────────────────────────────
document.addEventListener('mousemove', function(e) {
  if (isTodoResizing) {
    const todoRect = document.getElementById('cal').getBoundingClientRect();
    // Since #todo is between #cal and details, the new width is mouse.X - left of #todo
    // Actually, #todo is placed immediately after #cal.
    const newW = e.clientX - todoRect.right;
    if (newW >= 150 && newW <= 600) {
      document.getElementById('todo').style.width = newW + 'px';
      document.getElementById('todo').style.flex = 'none';
    }
  }

  if (isResizing) {
    const bodyEl = document.getElementById('body');
    const bodyRect = bodyEl.getBoundingClientRect();
    let newHeight = bodyRect.bottom - e.clientY;
    const minHeight = 100;
    const maxHeight = bodyRect.height - 150;
    if (newHeight < minHeight) newHeight = minHeight;
    if (newHeight > maxHeight) newHeight = maxHeight;
    document.getElementById('kan').style.height = newHeight + 'px';
  }

  if (resizeTaskId && resizeTaskEl) {
    const diffY = e.clientY - resizeStartY;
    // 1 pomodoro = 24px height
    const diffPomo = Math.round(diffY / 24);
    let newPomo = resizeStartPomo + diffPomo;
    if (newPomo < 1) newPomo = 1;
    resizeTaskEl.style.minHeight = (newPomo * 24) + 'px';
  }
});

document.addEventListener('mouseup', function() {
  if (isResizing) {
    isResizing = false;
    document.body.style.cursor = 'default';
    document.body.style.userSelect = '';
    document.getElementById('resizer').classList.remove('active');
  }
  if (isTodoResizing) {
    isTodoResizing = false;
    document.body.style.cursor = 'default';
    document.body.style.userSelect = '';
    document.getElementById('todo-resizer').classList.remove('active');
  }

  if (resizeTaskId) {
    const t = (state.tasks||[]).find(x => x.id === resizeTaskId);
    if (t) {
       const h = parseInt(resizeTaskEl.style.minHeight, 10) || 24;
       const finalPomo = Math.max(1, Math.round(h / 24));
       if (finalPomo !== (t.pomodoros || 1)) {
         send({ type: 'update', task: { ...t, pomodoros: finalPomo } });
       }
    }
    resizeTaskId = null;
    resizeTaskEl = null;
    if (!isResizing && !isTodoResizing) {
      document.body.style.cursor = 'default';
      document.body.style.userSelect = '';
    }
  }
});

// ── UTILS ──────────────────────────────────────────────────────────────────
function esc(s) {
  return String(s)
    .replace(/&/g,'&amp;')
    .replace(/</g,'&lt;')
    .replace(/>/g,'&gt;');
}

// ── POMODORO ───────────────────────────────────────────────────────────────
function togglePomodoro(id) {
  if (pomoTimerId && pomoTaskId === id) {
    clearInterval(pomoTimerId);
    pomoTimerId = null;
    pomoTaskId = null;
    renderStatus();
    return;
  }
  if (pomoTimerId) clearInterval(pomoTimerId);
  pomoTaskId = id;
  pomoEnd = Date.now() + 25 * 60000;
  pomoTimerId = setInterval(pomoTick, 1000);
  pomoTick();
}

function pomoTick() {
  const left = pomoEnd - Date.now();
  if (left <= 0) {
    clearInterval(pomoTimerId);
    pomoTimerId = null;
    const t = state.tasks.find(x => x.id === pomoTaskId);
    pomoTaskId = null;
    playBeep();
    renderStatus();
    if (t) {
      let p = t.pomodoros || 1;
      if (p > 1) {
        send({ type:'update', task: { ...t, pomodoros: p - 1 }});
      } else {
        send({ type:'setStatus', id: t.id, status: 'done' });
      }
    }
    return;
  }
  renderStatus();
}

function playBeep() {
  try {
    const ctx = new (window.AudioContext || window.webkitAudioContext)();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();
    osc.connect(gain);
    gain.connect(ctx.destination);
    osc.type = 'sine';
    osc.frequency.value = 800;
    gain.gain.setValueAtTime(0, ctx.currentTime);
    gain.gain.linearRampToValueAtTime(1, ctx.currentTime + 0.05);
    gain.gain.linearRampToValueAtTime(0, ctx.currentTime + 0.5);
    osc.start();
    osc.stop(ctx.currentTime + 0.5);
  } catch(e) {}
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
