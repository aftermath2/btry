import { Bet } from "./bets"
import { LotteryInfo } from "./lottery"
import { Winner } from "./winners"

export type ErrResponse = {
	readonly error: string
}

export type LNURLErrorResponse = {
	readonly status: string
	readonly reason: string
}

export type GetBetsResponse = {
	readonly bets: Bet[]
}

export type GetInfoResponse = LotteryInfo

export type GetInvoiceResponse = {
	readonly invoice: string
	readonly payment_id: number
}

export type LNURLWithdrawResponse = {
	readonly tag: string
	readonly callback: string
	readonly k1: string
	readonly default_description: string
	readonly min_withdrawable: number
	readonly max_withdrawable: number
}

export type GetPrizesResponse = {
	readonly prizes: number
}

export type GetWinnersResponse = {
	readonly winners: Winner[]
}

export type GetHeightsResponse = {
	readonly heights: number[]
}

export type GetLightningAddressResponse = {
	readonly address: string
	readonly has_address: boolean
}

export type SetLightningAddressResponse = {
	readonly success: boolean
}

export type WithdrawResponse = {
	readonly status: string
	readonly payment_id: number
}
