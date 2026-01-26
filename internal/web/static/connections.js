const panelState = new Map();
let lastGroups = [];
let searchValue = '';

document.getElementById('search').addEventListener('input', (e) => {
	searchValue = e.target.value.toLowerCase();
	render();
});

async function fetchConnections() {
	try {
		const res = await fetch('/connections/by-ip');
		if (!res.ok) return;

		lastGroups = await res.json();
		render();
	} catch (_) {
		// silent
	}
}

function render() {
	const container = document.getElementById('content');
	container.innerHTML = '';

	if (lastGroups.length === 0) return;

	for (const g of lastGroups) {
		// filter group + connections
		const matchedConns = g.conns.filter(c => {
			const user = (c.username || '').toLowerCase();
			const dst = c.destination.toLowerCase();

			return (
				g.source_ip.toLowerCase().includes(searchValue) ||
				user.includes(searchValue) ||
				dst.includes(searchValue)
			);
		});

		if (matchedConns.length === 0) continue;

		const details = document.createElement('details');
		details.dataset.sourceIp = g.source_ip;
		details.open = panelState.get(g.source_ip) ?? false;

		details.addEventListener('toggle', () => {
			panelState.set(g.source_ip, details.open);
		});

		const summary = document.createElement('summary');
		summary.textContent = `${g.source_ip} (${matchedConns.length} connections)`;
		details.appendChild(summary);

		for (const c of matchedConns) {
			const div = document.createElement('div');
			div.className = 'conn';

			const ageSec = Math.floor(
				(Date.now() - new Date(c.started_at)) / 1000
			);

			const user = c.username || '-';

			div.textContent =
				`ID=${c.id} USER=${user} DST=${c.destination} AGE=${ageSec}s`;

			details.appendChild(div);
		}

		container.appendChild(details);
	}
}

// polling
fetchConnections();
setInterval(fetchConnections, 2000);

// controls
document.getElementById('openAll').addEventListener('click', () => {
	for (const d of document.querySelectorAll('#content details')) {
		d.open = true;
		panelState.set(d.dataset.sourceIp, true);
	}
});

document.getElementById('closeAll').addEventListener('click', () => {
	for (const d of document.querySelectorAll('#content details')) {
		d.open = false;
		panelState.set(d.dataset.sourceIp, false);
	}
});
