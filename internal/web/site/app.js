async function readJSON(url) {
  const response = await fetch(url, { headers: { Accept: 'application/json' } });
  if (!response.ok) throw new Error(`HTTP ${response.status}`);
  return response.json();
}

function renderStations(values) {
  const root = document.querySelector('#stations');
  document.querySelector('#station-count').textContent = values.length;
  if (!values.length) { root.innerHTML = '<p class="muted">还没有活跃站点。</p>'; return; }
  root.innerHTML = values.map((station) => `
    <div class="station"><div><strong>${station.name}</strong><small>${station.code} · ${station.habitat || '未注明栖息地'}</small></div><span class="status">${station.timezone}</span></div>`).join('');
}

async function refresh() {
  const health = document.querySelector('#health');
  try {
    await readJSON('/healthz');
    health.textContent = '服务在线';
    renderStations(await readJSON('/api/stations'));
  } catch (error) {
    health.textContent = '无法连接';
    document.querySelector('#stations').innerHTML = `<p class="muted">${error.message}</p>`;
  }
}

document.querySelector('#refresh').addEventListener('click', refresh);
refresh();
