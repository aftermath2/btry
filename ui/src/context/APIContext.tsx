import { createContext, useContext, JSX } from "solid-js";

import { API } from "../api/api";

const APIContext = createContext<API>()

export function APIProvider(props: { children: JSX.Element }) {
	const api = new API()

	return (
		<APIContext.Provider value={api}>
			{props.children}
		</APIContext.Provider>
	)
}

export function useAPIContext(): API {
	const context = useContext(APIContext)
	if (!context) {
		throw new Error("useAPIContext: couldn't find APIContext")
	}
	return context
}
