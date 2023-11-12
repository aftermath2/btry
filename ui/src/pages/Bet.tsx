import { Component, JSX, Show, createSignal, onCleanup, onMount } from 'solid-js';
import { useI18n } from "@solid-primitives/i18n";
import toast from 'solid-toast';
import { createStore } from "solid-js/store";

import styles from './Bet.module.css';
import arrowLeftIcon from "../assets/icons/arrow_left.svg"
import copyIcon from "../assets/icons/copy.svg"
import QRCode from "../components/QRCode";
import Button from "../components/Button";
import { BeautifyNumber, NumberRegex } from "../utils/utils";
import { HandleError, WriteClipboard } from "../utils/actions";
import { GetInvoiceResponse } from "../types/api";
import Input from "../components/Input";
import Container from "../components/Container";
import Box from "../components/Box";
import Modal from "../components/Modal";
import { useAuthContext } from "../context/AuthContext";
import { Status } from "../types/events";
import { useAPIContext } from "../context/APIContext";
import { Event as SSEEvent } from "../api/sse";

const Bet: Component = () => {
	const [auth] = useAuthContext()
	const api = useAPIContext()
	const [t] = useI18n()

	const [capacity, setCapacity] = createSignal(0)
	const [amount, setAmount] = createSignal(1)
	const [invoice, setInvoice] = createSignal("")
	const [showInvoice, setShowInvoice] = createSignal(false)
	const [showWarning, setShowWarning] = createSignal(false)
	const [paymentIDs, setPaymentIDs] = createStore<number[]>([])

	const getInvoice = async (amount: number): Promise<GetInvoiceResponse> => {
		return await api.GetInvoice(amount)
	}

	const getLotteryCapacity = async () => {
		const info = await api.GetLottery()
		setCapacity(info.capacity)
	}

	const handleInput: JSX.EventHandlerUnion<HTMLInputElement, Event> = (event) => {
		const amount = Number(event.currentTarget.value.trim().replace(/\D/g, ""))
		setAmount(amount)
	}

	const validateBet = async (): Promise<void> => {
		if (amount() < 1) {
			throw Error("Invalid amount")
		}
		if (amount() > capacity()) {
			throw Error(`Amount is higher than the available capacity (${BeautifyNumber(capacity())})`)
		}
		setShowWarning(true)
	}

	const handleBet = async (): Promise<void> => {
		const resp = await getInvoice(amount())
		setInvoice(resp.invoice)
		setPaymentIDs(paymentIDs.length, resp.payment_id)

		// Reset input field
		setAmount(1)
	}

	onMount(() => {
		getLotteryCapacity()
		api.Subscribe(SSEEvent.Invoices, (payload) => {
			if (paymentIDs.includes(payload.payment_id)) {
				if (payload.status === Status.Success) {
					toast.success(t("bet_sent"), { duration: 3000 })
					setShowInvoice(false)
				}
				// Remove payment ID from the array
				setPaymentIDs(paymentIDs.filter(id => id !== payload.payment_id))
			}
		})
	})

	onCleanup(() => {
		api.Close()
	})

	return (
		<Container>
			<Modal
				show={showWarning()}
				onClose={() => setShowWarning(false)}
				title={t("backup_private_key")}
			>
				<div class={styles.warning}>
					<p>{t("backup_private_key_message")}</p>
					<div class={styles.private_key}>
						<p>{auth().privateKey}</p>
						<button class={styles.copy}>
							<img
								class={styles.copy_icon}
								src={copyIcon}
								onClick={() => WriteClipboard(auth().privateKey, t("clipboard_copy"))}
							/>
						</button>
					</div>

					<Button
						text={t("continue")}
						width="40%"
						onClick={async () => {
							await HandleError(() => handleBet())()
							setShowWarning(false)
							setShowInvoice(true)
						}}
					/>
				</div>
			</Modal>

			<Box
				title={showInvoice() ? undefined : t("place_bet").toUpperCase()}
				align="center"
				minWidth="20%"
				padding="30px 25px"
				titleFontSize="24px"
				titleMargin="0"
			>
				<Show when={!showInvoice()}>
					<div class={styles.bet}>
						<Input
							title={`${t("amount")} (sats)`}
							handleInput={handleInput}
							onEnter={HandleError(() => validateBet())}
							validate={(v) => NumberRegex.test(v)}
							value={BeautifyNumber(amount())}
							focus
							baseProps={{
								min: "1",
								step: "1",
								maxLength: "11",
								autofocus: true,
								required: true
							}}
						/>
						<Button text={t("bet")} onClick={HandleError(() => validateBet())} />
					</div>
				</Show>

				<Show when={showInvoice()}>
					<div class={styles.invoice_header}>
						<div class={styles.invoice_left} onClick={() => setShowInvoice(false)}>
							<img class={styles.back_icon} src={arrowLeftIcon} />
						</div>
						<p class={styles.title}>{t("pay_invoice").toUpperCase()}</p>
						<div class={styles.invoice_right}></div>
					</div>

					<QRCode value={invoice()} />

					<div class={styles.invoice}>
						<p class={styles.pr}>{invoice().slice(0, 18)}...{invoice().slice(-18)}</p>
					</div>

					<div class={styles.buttons}>
						<Button
							text={t("copy_text")}
							onClick={() => WriteClipboard(invoice(), t("clipboard_copy"))}
						/>
						<Button
							text={t("open_link")}
							onClick={(e) => {
								e.preventDefault()
								window.location.href = "lightning:" + invoice()
							}}
						/>
					</div>
				</Show>
			</Box>
		</Container>
	);
};

export default Bet;
