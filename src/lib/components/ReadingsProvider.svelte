<script lang="ts">
	import { createResourceClient, createResourceQuery } from '@viamrobotics/svelte-sdk';
	import { SensorClient } from '@viamrobotics/sdk';
	import { setContext } from 'svelte';
	import { READINGS_CONTEXT_KEY, type ReadingsContextValue } from '$lib/contexts/readings';

	interface Props {
		partID: string;
		name: string;
		refetchInterval?: number;
		children?: import('svelte').Snippet;
	}

	let { partID, name, refetchInterval = 1000, children }: Props = $props();

	const sensorClient = createResourceClient(SensorClient, () => partID, () => name);
	const query = createResourceQuery(sensorClient, 'getReadings', () => ({
		refetchInterval
	}));

	const contextValue: ReadingsContextValue = {
		query,
		refetch: () => {
			query.refetch();
		}
	};
	setContext(READINGS_CONTEXT_KEY, contextValue);
</script>

{#if children}
	{@render children()}
{/if}
