import { Signal, createSignal } from "solid-js";

/**
 * createStorageSignal returns a signal that stores the key/value in the storage specified.
 */
export function createStorageSignal<T>(
	key: string,
	defaultValue: T,
	storage = localStorage
): Signal<T> {
	const storedValue = storage.getItem(key)
	const initialValue = storedValue ? JSON.parse(storedValue) as T : defaultValue;

	const [value, setValue] = createSignal<T>(initialValue);
	const setValueAndStore = ((v) => {
		const val = setValue(v);
		storage.setItem(key, JSON.stringify(val));
		return val;
	}) as typeof setValue;

	return [value, setValueAndStore];
}