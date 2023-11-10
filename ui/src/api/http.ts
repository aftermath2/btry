import { getAuthFromStorage } from "../context/AuthContext";
import { ErrResponse, LNURLErrorResponse } from "../types/api";

export interface Req {
	url: string,
	keepalive?: boolean,
	headers?: { [key: string]: string },
	signal?: AbortSignal,
}

export interface ReqWithBody extends Req {
	body?: BodyInit
}

enum Status {
	NoContent = 204,
	BadRequest = 400,
	NotFound = 404,
	TooManyRequests = 429,
	InternalServerError = 500
}

export class HTTP {

	static async delete(req: Req): Promise<void> {
		setAuthorizationHeader(req);
		const res = await fetch(req.url, {
			method: "DELETE",
			keepalive: req.keepalive,
			signal: req.signal,
			headers: {
				"Accept": "application/json",
				...(req.headers),
			},
			credentials: "same-origin",
		})
		if (!res.ok) {
			panic(await parseErr(res))
		}
	}

	static async get<T>(req: Req): Promise<T> {
		setAuthorizationHeader(req)
		const res = await fetch(req.url, {
			method: "GET",
			keepalive: req.keepalive,
			signal: req.signal,
			headers: {
				"Accept": "application/json",
				...(req.headers),
			},
		})
		if (!res.ok) {
			panic(await parseErr(res))
		}
		if (res.status === Status.NoContent) {
			return new Promise(() => { });
		}

		const json = await res.json()
		return <T>json
	}

	static async post<T>(req: ReqWithBody): Promise<T> {
		setAuthorizationHeader(req)
		const res = await fetch(req.url, {
			method: "POST",
			body: req.body,
			keepalive: req.keepalive,
			signal: req.signal,
			headers: {
				"Accept": "application/json",
				"Content-Type": "application/json; charset=UTF-8",
				...(req.headers),
			},
		})
		if (!res.ok) {
			panic(await parseErr(res))
		}

		const json = await res.json()
		return <T>json
	}

	static async put(req: ReqWithBody): Promise<Response> {
		setAuthorizationHeader(req)
		const res = await fetch(req.url, {
			method: "PUT",
			body: req.body,
			keepalive: req.keepalive,
			signal: req.signal,
			headers: {
				"Accept": "application/json",
				"Content-Type": "application/json; charset=UTF-8",
				...(req.headers),
			},
		})
		if (!res.ok) {
			panic(await parseErr(res))
		}

		return res
	}

	/**
	 * paralell executes multiple promises in parallel.
	 * Calling too many promises simultaneously may overload the device's memory.
	 * @param requests array of promises
	 * @returns an array of the responses received, only if all of the requests succeeded
	 */
	static async paralell(requests: PromiseLike<unknown>[]): Promise<unknown[]> {
		return await Promise.all(requests)
	}

}

const parseErr = async (res: Response): Promise<string> => {
	if (res.status === Status.TooManyRequests) {
		return "Too many requests"
	}
	if (res.status === Status.NotFound) {
		return "404 - Not found"
	}
	const err: ErrResponse | LNURLErrorResponse = await res.json()
	return JSON.stringify({ url: res.url, message: err }, null, 4)
}

export function panic(message: string) {
	throw new Error(message)
}

const setAuthorizationHeader = async (req: Req | ReqWithBody): Promise<void> => {
	const auth = getAuthFromStorage()

	if (auth?.publicKey === "") {
		return
	}

	const value = `Bearer ${auth?.publicKey}`

	if (req.headers) {
		req.headers["Authorization"] = value
		return
	}

	req.headers = { "Authorization": value }
}