<script lang="ts">
	import { getContext } from 'svelte';
	import { READINGS_CONTEXT_KEY, type ReadingsContextValue } from '$lib/contexts/readings';

	const readingsContext = getContext<ReadingsContextValue>(READINGS_CONTEXT_KEY);
	if (!readingsContext) {
		throw new Error('AllReadings must be used inside ReadingsProvider');
	}

	const { query, refetch } = readingsContext;

	/** Sorted entries from readings record, always by key A–Z */
	const sortedReadings = $derived(
		query.data && typeof query.data === 'object' && !Array.isArray(query.data)
			? Object.entries(query.data).sort(([a], [b]) =>
					a.localeCompare(b, undefined, { sensitivity: 'base' })
				)
			: []
	);
</script>

<div class="rounded-xl border border-slate-200 bg-white shadow-sm dark:border-slate-700 dark:bg-slate-800/50">
	<div class="border-b border-slate-200 px-4 py-3 dark:border-slate-700">
		<h2 class="text-lg font-semibold text-slate-800 dark:text-slate-100">Sensor readings</h2>
		<p class="text-sm text-slate-500 dark:text-slate-400">Live values, sorted by key</p>
	</div>

	{#if query.isLoading}
		<div class="flex items-center gap-3 px-4 py-8 text-slate-500 dark:text-slate-400">
			<span class="inline-block h-5 w-5 animate-spin rounded-full border-2 border-slate-300 border-t-slate-600 dark:border-slate-600 dark:border-t-slate-400"></span>
			<span>Loading readings…</span>
		</div>
	{:else if query.isError}
		<div class="rounded-lg bg-red-50 px-4 py-3 text-red-700 dark:bg-red-900/20 dark:text-red-300">
			<span class="font-medium">Error:</span> {query.error?.message}
		</div>
	{:else if sortedReadings.length === 0}
		<p class="px-4 py-6 text-slate-500 dark:text-slate-400">No readings available.</p>
	{:else}
		<div class="max-h-[320px] overflow-y-auto overflow-x-auto">
			<table class="w-full min-w-[280px] text-left text-sm">
				<thead class="sticky top-0 z-10 bg-slate-50/95 dark:bg-slate-800/95 backdrop-blur-sm">
					<tr class="border-b border-slate-200 dark:border-slate-700">
						<th class="px-3 py-2 font-semibold text-slate-700 dark:text-slate-300">Reading</th>
						<th class="px-3 py-2 font-semibold text-slate-700 dark:text-slate-300">Value</th>
					</tr>
				</thead>
				<tbody>
					{#each sortedReadings as [key, value]}
						<tr class="border-b border-slate-100 transition-colors hover:bg-slate-50 dark:border-slate-700/80 dark:hover:bg-slate-800/50">
							<td class="px-3 py-2 font-medium text-slate-800 dark:text-slate-200">{key}</td>
							<td class="px-3 py-2">
								{#if value !== null && value !== undefined && typeof value === 'object' && !Array.isArray(value)}
									<pre class="inline-block max-w-[200px] overflow-x-auto rounded bg-slate-100 px-2 py-0.5 text-xs text-slate-700 dark:bg-slate-700 dark:text-slate-300">{JSON.stringify(value)}</pre>
								{:else}
									<span class="text-slate-600 dark:text-slate-300">{typeof value === 'number' ? value : String(value)}</span>
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>