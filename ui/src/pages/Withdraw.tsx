import { Component, JSX, Show, createResource, createSignal, onCleanup, onMount } from 'solid-js';
import { createStore } from "solid-js/store";
import { useI18n } from "@solid-primitives/i18n";
import toast from "solid-toast";

import styles from './Withdraw.module.css';
import { useAuthContext } from "../context/AuthContext";
import { getLNURLWithdrawURL } from "../api/api";
import { Sign } from "../utils/crypto";
import Input from "../components/Input";
import { BeautifyNumber, NumberRegex } from "../utils/utils";
import { HandleError } from "../utils/actions";
import Button from "../components/Button";
import { Invoice, LNURLEncode, ValidateInvoice } from "../utils/lightning";
import QRCode from "../components/QRCode";
import Loading from "../components/Loading";
import Container from "../components/Container";
import Box from "../components/Box";
import { Status } from "../types/events";
import satoshiIcon from "../assets/icons/satoshi.svg"
import { useAPIContext } from "../context/APIContext";

const errNoPrizes = Error("No prizes available to withdraw")
const errInvalidFee = Error("Invalid fee amount")

interface Payment {
	id: number
	hash: string
}

const Withdraw: Component = () => {
	const [auth] = useAuthContext()
	const api = useAPIContext()
	const [t] = useI18n()

	const [invoice, setInvoice] = createSignal("")
	const [fee, setFee] = createSignal(1)
	const [payments, setPayments] = createStore<Payment[]>([])

	const getPrizes = async (): Promise<number> => {
		const resp = await api.GetPrizes()
		return resp.prizes
	}
	const [prizes, prizesOptions] = createResource<number>(getPrizes)

	const getLNURLWithdraw = async (): Promise<string> => {
		const signature = await Sign(auth().privateKey, auth().publicKey)
		const url = getLNURLWithdrawURL(auth().publicKey, signature)
		return LNURLEncode(url)
	}
	const [lnurlWithdraw] = createResource<string>(getLNURLWithdraw)

	const handleFeeInput: JSX.EventHandlerUnion<HTMLInputElement, Event> = (event) => {
		const fee = Number(event.currentTarget.value.trim().replace(/\D/g, ""))
		setFee(fee)
	}

	const handleInvoiceInput: JSX.EventHandlerUnion<HTMLInputElement, Event> = (event) => {
		const invoice = event.currentTarget.value.trim()
		setInvoice(invoice)
	}

	const withdraw = async (): Promise<void> => {
		const availablePrizes = prizes() || 0
		if (availablePrizes === 0) {
			throw errNoPrizes
		}

		if (fee() < 0) {
			throw errInvalidFee
		}

		let inv: Invoice
		try {
			inv = ValidateInvoice(invoice(), undefined, availablePrizes - fee())

			if (payments.some(payment => payment.hash === inv.paymentHash)) {
				throw Error("already used")
			}
		} catch (error: any) {
			throw Error("Invalid invoice: " + error.message)
		}

		const signature = await Sign(auth().privateKey, auth().publicKey)
		const resp = await api.Withdraw(signature, invoice(), auth().publicKey, fee())
		setPayments(payments.length, {
			id: resp.payment_id,
			hash: inv.paymentHash,
		})

		toast.loading(t("withdrawal_request_sent"), { duration: 2000 })

		// Update available funds and input fields
		prizesOptions.mutate(availablePrizes - fee() - inv.amountSat)
		setFee(1)
		setInvoice("")
	}

	onMount(() => {
		api.Subscribe("payments", (payload) => {
			if (!payments.some(payment => payment.id === payload.payment_id)) {
				return
			}

			if (payload.status === Status.Success) {
				toast.success(t("withdrawal_success"), { duration: 3000 })
			} else {
				toast.error(`${t("withdrawal_failed")}: ${payload.error}}`, { duration: 3000 })
				prizesOptions.refetch()
				// Remove payment from the array so the same invoice can be used
				setPayments(payments.filter(payment => payment.id !== payload.payment_id))
			}
		})
	})

	onCleanup(() => {
		api.Close()
	})

	return (
		<Container>
			<Box
				title={t("withdraw").toUpperCase()}
				align="center"
				minWidth="20%"
				padding="30px 25px"
				titleFontSize="24px"
			>
				<div class={styles.available}>
					<p class={styles.text}>{`${t("available")}:`}</p>
					<Show when={!prizes.loading} fallback={<Loading margin="0" width="18px" />}>
						<p class={styles.prizes}>{BeautifyNumber(prizes())}</p>
					</Show>

					<img
						class={styles.sats}
						src={satoshiIcon}
					/>
				</div>

				<div class={styles.withdraw}>
					<Input
						title={`${t("fee")} (sats)`}
						handleInput={handleFeeInput}
						validate={(v) => NumberRegex.test(v)}
						value={BeautifyNumber(fee())}
						focus
						baseProps={{
							min: "1",
							step: "1",
							maxLength: "11",
							autofocus: true,
							required: true
						}}
					/>
					<Input
						title={t("invoice")}
						placeholder={`${t("invoice")} (${t("amount").toLowerCase()} ${t("required").toLowerCase()})`}
						handleInput={handleInvoiceInput}
						value={invoice()}
						baseProps={{
							autofocus: false,
							required: true
						}}
					/>
					<Button
						text={t("withdraw")}
						onClick={HandleError(() => withdraw())}
					/>
				</div>

				<hr class={styles.line} />
				<p class={styles.title}>LNURL WITHDRAW</p>
				<Show when={!lnurlWithdraw.loading} fallback={<Loading />}>
					<QRCode value={lnurlWithdraw()?.toUpperCase() || ""} />
				</Show>
			</Box>
		</Container>
	);
};

export default Withdraw;
