<script lang="ts">
	import { DEFAULT_PART_ID } from '$lib/consts';
	import type { DialConf } from '@viamrobotics/sdk';
	import { ViamProvider } from '@viamrobotics/svelte-sdk';
	import { getCookie } from 'typescript-cookie';
	import AllReadings from './AllReadings.svelte';
	import BatteryPercent from './BatteryPercent.svelte';
	import ReadingsProvider from './ReadingsProvider.svelte';

	let { host, credentials, machineId } = $props<{
		host: string;
		credentials: any;
		machineId?: string | null;
	}>();

	const dialConfigs: Record<string, DialConf> = $derived({
		[DEFAULT_PART_ID]: {
			host: host,
			credentials: credentials,
			signalingAddress: getCookie('is_local')
				? `http://localhost:8081` // Hard-coded to the custom viam-server bind port
				: 'https://app.viam.com',
			disableSessions: false
		}
	});

	$effect(() => {
		console.log(
			`Connecting to ${host} with credentials: ${JSON.stringify(credentials, undefined, 4)}`
		);
	});
</script>

<ViamProvider {dialConfigs}>
	<ReadingsProvider partID={DEFAULT_PART_ID} name="sensor">
		<div>
			<p><strong>Host:</strong> {host}</p>
			<p class="mt-2 flex items-center gap-2">
				<strong>Battery:</strong> <BatteryPercent />
			</p>
		</div>
		<AllReadings />
	</ReadingsProvider>
</ViamProvider>
