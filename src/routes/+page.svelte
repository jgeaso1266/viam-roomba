<script lang="ts">
	import { getHostAndCredentials, type HostAndCredentials } from '$lib/auth';
	import { browser } from '$app/environment';
	import Main from '$lib/components/Main.svelte';

	const emptyCreds: HostAndCredentials = {
		host: '',
		credentials: { type: 'api-key', payload: '', authEntity: '' },
		machineId: null
	};
	let hostAndCreds = $state<HostAndCredentials>(emptyCreds);

	$effect(() => {
		if (browser) {
			hostAndCreds = getHostAndCredentials();
		}
	});
</script>

{#if browser && hostAndCreds.host}
	<Main
		host={hostAndCreds.host}
		credentials={hostAndCreds.credentials}
		machineId={hostAndCreds.machineId}
	/>
{/if}