<script lang="ts">
	import { getContext } from 'svelte';
	import { READINGS_CONTEXT_KEY, type ReadingsContextValue } from '$lib/contexts/readings';

	const readingsContext = getContext<ReadingsContextValue>(READINGS_CONTEXT_KEY);
	if (!readingsContext) {
		throw new Error('BatteryPercent must be used inside ReadingsProvider');
	}

	const { query } = readingsContext;

	/** battery_percent from readings, 0–100; undefined if missing or invalid */
	const batteryPercent = $derived.by(() => {
		const data = query.data;
		if (data == null || typeof data !== 'object' || Array.isArray(data)) return undefined;
		const raw = (data as Record<string, unknown>)['battery_percent'];
		if (typeof raw !== 'number') return undefined;
		return Math.min(100, Math.max(0, raw));
	});

	/** Hue for fill: 120 = green (full), 0 = red (empty) */
	const fillHue = $derived(batteryPercent != null ? (batteryPercent / 100) * 120 : 0);
</script>

<div class="battery">
	<div class="battery-body">
		<div
			class="battery-fill"
			style="width: {batteryPercent != null ? batteryPercent : 0}%; background-color: hsl({fillHue}, 70%, 45%);"
			aria-hidden="true"
		></div>
	</div>
	<div class="battery-tip" aria-hidden="true"></div>
	<span class="battery-label">
		{batteryPercent != null ? `${Math.round(batteryPercent)}%` : '—'}
	</span>
</div>

<style>
	.battery {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
	}

	.battery-body {
		position: relative;
		width: 4rem;
		height: 1.5rem;
		border: 2px solid var(--battery-border, #374151);
		border-radius: 4px;
		overflow: hidden;
		background: var(--battery-bg, #1f2937);
	}

	.battery-fill {
		position: absolute;
		left: 0;
		top: 0;
		bottom: 0;
		border-radius: 2px;
		transition: width 0.3s ease, background-color 0.3s ease;
	}

	.battery-tip {
		width: 0.25rem;
		height: 0.75rem;
		background: var(--battery-border, #374151);
		border-radius: 0 2px 2px 0;
	}

	.battery-label {
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--battery-text, #374151);
		min-width: 2.5rem;
	}

	:global(.dark) .battery-body {
		--battery-border: #6b7280;
		--battery-bg: #374151;
	}

	:global(.dark) .battery-tip {
		--battery-border: #6b7280;
	}

	:global(.dark) .battery-label {
		--battery-text: #e5e7eb;
	}
</style>
