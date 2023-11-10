import { Emitter, Listener, createEmitter } from "@solid-primitives/event-bus";
import { InfoPayload, InvoicesPayload, PaymentsPayload } from "../types/events";

const eventSourceURL = `${import.meta.env.VITE_API_URL}/api/events?stream=events`

type EventPayload = InfoPayload | InvoicesPayload | PaymentsPayload

export interface EventMap {
	"info": InfoPayload
	"invoices": InvoicesPayload
	"payments": PaymentsPayload
}

export enum Event {
	Info = "info",
	Invoices = "invoices",
	Payments = "payments"
}

export class SSE {

	private stream: EventSource
	private eventEmitter: Emitter<EventMap>
	// In seconds
	private delay = 3

	constructor() {
		this.stream = new EventSource(eventSourceURL)
		this.stream.addEventListener("error", () => this.reconnect())
		this.stream.addEventListener("open", () => this.delay = 1)

		this.eventEmitter = createEmitter()
		this.forwardEvents()
	}

	private forwardEvents(): void {
		for (const [_, name] of Object.entries(Event)) {
			this.stream.addEventListener(name, (event) => {
				const payload: EventPayload = JSON.parse(event.data)
				this.eventEmitter.emit(name, payload)
			})
		}
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

	Subscribe<T extends Event>(event: T, callback: Listener<EventMap[T]>): void {
		this.eventEmitter.on(event, callback)
	}

	Close(): void {
		this.eventEmitter.clear()
	}
}