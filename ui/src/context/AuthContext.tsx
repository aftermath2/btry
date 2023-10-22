import { createContext, useContext, JSX, Signal } from "solid-js";

import { createStorageSignal } from "../utils/storage";
import { Auth } from "../types/auth";

const storageKey = "auth"
const AuthContext = createContext<Signal<Auth>>()

export function AuthProvider(props: { children: JSX.Element }) {
	const [auth, setAuth] = createStorageSignal<Auth>(
		storageKey,
		{ privateKey: "", publicKey: "", nickname: "" },
	)

	return (
		<AuthContext.Provider value={[auth, setAuth]}>
			{props.children}
		</AuthContext.Provider>
	)
}

export function useAuthContext(): Signal<Auth> {
	const context = useContext(AuthContext)
	if (!context) {
		throw new Error("useAuthContext: cannot find AuthContext")
	}
	return context
}

export function getAuthFromStorage(): Auth | undefined {
	const item = localStorage.getItem(storageKey)
	const auth = item ? JSON.parse(item) as Auth : undefined
	return auth
}