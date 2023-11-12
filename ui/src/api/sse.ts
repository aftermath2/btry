import { InfoPayload, InvoicesPayload, PaymentsPayload } from "../types/events";

const eventSourceURL = `${import.meta.env.VITE_API_URL}/api/events?stream=events`

export interface Events {
	"info": InfoPayload
	"invoices": InvoicesPayload
	"payments": PaymentsPayload
}

export class SSE {

	private stream: EventSource
	// In seconds
	private delay = 3

	constructor() {
		this.stream = new EventSource(eventSourceURL)
		this.stream.addEventListener("error", () => this.reconnect())
		this.stream.addEventListener("open", () => this.delay = 1)
	}

	private reconnect(): void {
		setTimeout(() => {
			this.stream = new EventSource(eventSourceURL)
			this.stream.addEventListener("error", () => this.reconnect())
			this.stream.addEventListener("open", () => this.delay = 1)
		}, this.delay * 1000)

		// Increase frequency and cap at one minute 
		this.delay *= 2
		if (this.delay > 60) {
			this.delay = 60
		}
	}

	Subscribe<T extends keyof Events>(event: T, onEvent: (payload: Events[T]) => void): void {
		this.stream.addEventListener(event, (e) => {
			const payload: Events[T] = JSON.parse(e.data)
			onEvent(payload)
		})
	}

	Close(): void {
		const events: Array<keyof Events> = ["info", "invoices", "payments"]
		for (const event of events) {
			this.stream.removeEventListener(event, () => { })
		}
	}
}
