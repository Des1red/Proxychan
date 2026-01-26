async function fetchConnections() {
	try {
		const res = await fetch('/connections/by-ip');
		if (!res.ok) return;

		const groups = await res.json();
		const container = document.getElementById('content');
		container.innerHTML = '';

		if (groups.length === 0) {
			container.textContent = 'No active connections';
			return;
		}

		for (const g of groups) {
			const details = document.createElement('details');
			details.open = true;

			const summary = document.createElement('summary');
			summary.textContent = `${g.source_ip} (${g.count} connections)`;
			details.appendChild(summary);

			for (const c of g.conns) {
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
	} catch (_) {
		// silent fail
	}
}

fetchConnections();
setInterval(fetchConnections, 2000);
