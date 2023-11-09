import {
	GetBetsResponse, GetInfoResponse, GetInvoiceResponse,
	GetPrizesResponse, GetWinnersResponse, LNURLWithdrawResponse, WithdrawResponse
} from "../types/api";
import { InfoPayload, InvoicesPayload, PaymentsPayload } from "../types/events";
import { HTTP } from "./http";
import { EventName, SSE } from "./sse";

export const API_URL = `${import.meta.env.VITE_API_URL}/api`

export const getLNURLWithdrawURL = (publicKey: string, signature: string): string => {
	return `${API_URL}/lnurl/withdraw?pubkey=${publicKey}&signature=${signature}`
}

export class API {

	private abortController: AbortController
	private sse?: SSE

	constructor() {
		this.abortController = new AbortController()
	}

	Abort(): void {
		this.abortController.abort()
		this.sse && this.sse.close()
	}

	async GetBets(offset: number = 0, limit: number = 0, reverse: boolean = false): Promise<GetBetsResponse> {
		return await HTTP.get<GetBetsResponse>({
			url: `${API_URL}/bets?offset=${offset}&limit=${limit}&reverse=${reverse}`,
			keepalive: true,
			signal: this.abortController.signal
		})
	}

	async GetLottery(): Promise<GetInfoResponse> {
		return await HTTP.get<GetInfoResponse>({
			url: `${API_URL}/lottery`,
			keepalive: true,
			signal: this.abortController.signal
		})
	}

	async GetInvoice(amount: number): Promise<GetInvoiceResponse> {
		return await HTTP.get<GetInvoiceResponse>({
			url: `${API_URL}/invoice?amount=${amount}`,
			keepalive: true,
			signal: this.abortController.signal
		})
	}

	async GetPrizes(): Promise<GetPrizesResponse> {
		return await HTTP.get<GetPrizesResponse>({
			url: `${API_URL}/prizes`,
			keepalive: true,
			signal: this.abortController.signal
		})
	}

	async GetWinners(): Promise<GetWinnersResponse> {
		return await HTTP.get<GetWinnersResponse>({
			url: `${API_URL}/winners`,
			keepalive: true,
			signal: this.abortController.signal
		})
	}

	async GetWinnersHistory(from: number = 0, to: number = 0): Promise<GetWinnersResponse> {
		return await HTTP.get<GetWinnersResponse>({
			url: `${API_URL}/winners/history?from=${from}&to=${to}`,
			keepalive: true,
			signal: this.abortController.signal
		})
	}

	async LNURLWithdraw(publicKey: string, signature: string): Promise<LNURLWithdrawResponse> {
		return await HTTP.get<LNURLWithdrawResponse>({
			url: getLNURLWithdrawURL(publicKey, signature),
			keepalive: true,
			signal: this.abortController.signal
		})
	}

	async Withdraw(k1: string, pr: string, publicKey: string, fee: number): Promise<WithdrawResponse> {
		return await HTTP.post<WithdrawResponse>({
			url: `${API_URL}/withdraw?k1=${k1}&pr=${pr}&pubkey=${publicKey}&fee=${fee}`,
			keepalive: true,
			signal: this.abortController.signal
		})
	}

	private newSSE(): void {
		if (!this.sse) {
			this.sse = new SSE()
		}
	}

	ListenLotteryInfoEvents(onEvent: (payload: InfoPayload) => void): void {
		this.newSSE()
		this.sse && this.sse.listen<InfoPayload>(EventName.Info, onEvent)
	}

	ListenInvoicesEvents(onEvent: (payload: InvoicesPayload) => void): void {
		this.newSSE()
		this.sse && this.sse.listen<InvoicesPayload>(EventName.Invoices, onEvent)
	}

	ListenPaymentsEvents(onEvent: (payload: PaymentsPayload) => void): void {
		this.newSSE()
		this.sse && this.sse.listen<PaymentsPayload>(EventName.Payments, onEvent)
	}
}