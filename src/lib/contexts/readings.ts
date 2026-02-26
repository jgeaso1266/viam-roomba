import type { QueryObserverResult } from '@tanstack/svelte-query';

/**
 * Readings are returned from the sensor as a record (key -> value).
 * This type is the same shape as getReadings() result.
 */
export type ReadingsRecord = Record<string, unknown>;

export interface ReadingsContextValue {
	/** Reactive query result; use .data for the current readings record, .isLoading, .error */
	query: QueryObserverResult<ReadingsRecord>;
	/** Trigger a refetch of the latest readings */
	refetch: () => void;
}

export const READINGS_CONTEXT_KEY = Symbol('readings');
