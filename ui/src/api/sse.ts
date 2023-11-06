import { InfoPayload, InvoicesPayload, PaymentsPayload } from "../types/events";

const eventSourceURL = `${import.meta.env.VITE_API_URL}/api/events?stream=events`

type EventPayload = InfoPayload | InvoicesPayload | PaymentsPayload

export enum EventName {
	Info = "info",
	Invoices = "invoices",
	Payments = "payments"
}

export class SSE {

	private stream: EventSource
	// In seconds
	private delay = 3
	private listening: string[] = []

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

	listen<T extends EventPayload>(eventName: EventName, onEvent: (payload: T) => void): void {
		// Do not listen for the same event type more than once
		if (this.listening.includes(eventName)) {
			return
		}

		this.stream.addEventListener(eventName, (event) => {
			const payload: T = JSON.parse(event.data)
			onEvent(payload)
		})

		this.listening.push(eventName)
	}

	close(): void {
		this.stream.close()
	}
}