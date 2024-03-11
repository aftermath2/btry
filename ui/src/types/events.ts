import { Winner } from "./winners"

export enum Status {
	Failed = 0,
	Success = 1
}

export type InfoPayload = {
	readonly winners?: Winner[]
	readonly capacity?: number
	readonly prize_pool?: number
	readonly next_height?: number
	readonly block_height?: number
}

export type InvoicesPayload = {
	readonly error: string
	readonly payment_id: number
	readonly public_key: string
	readonly amount: number
	readonly status: Status
}

export type PaymentsPayload = {
	readonly error: string
	readonly payment_id: number
	readonly status: Status
}
