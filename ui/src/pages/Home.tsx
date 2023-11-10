import { Component, createSignal, createEffect, onCleanup, createResource, Show } from 'solid-js';
import { useI18n } from "@solid-primitives/i18n";

import styles from './Home.module.css';

import Winners from "../components/Winners";
import Bets from "../components/Bets";
import { LotteryInfo } from "../types/lottery";
import Loading from "../components/Loading";
import { FormatTime } from "../utils/utils";
import Container from "../components/Container";
import { Status } from "../types/events";
import Sats from "../components/Sats";
import { useAPIContext } from "../context/APIContext";
import { Event } from "../api/sse";

const numPrizes = 8

const Home: Component = () => {
	const api = useAPIContext()
	const [t] = useI18n()

	const [timeLeft, setTimeLeft] = createSignal("")
	const [resetBets, setResetBets] = createSignal(false)

	const getLotteryInfo = async (): Promise<LotteryInfo> => {
		return await api.GetLottery()
	}
	const [info, infoOptions] = createResource<LotteryInfo>(getLotteryInfo)

	const updateTimer = () => {
		const now = new Date()
		const hours = 23 - now.getUTCHours()
		const minutes = 59 - now.getUTCMinutes()
		const seconds = 59 - now.getUTCSeconds()

		if (hours + minutes + seconds <= 0) {
			setTimeLeft("00:00:00")
			return
		}

		const timeLeft = `${hours}:${minutes}:${seconds}`
		setTimeLeft(FormatTime(timeLeft))
	}

	const timer = setInterval(updateTimer, 1000)

	createEffect(() => {
		updateTimer()

		api.Subscribe(Event.Info, (payload) => {
			const updatedInfo: LotteryInfo = {
				prize_pool: payload.prize_pool,
				capacity: payload.capacity
			}
			infoOptions.mutate(updatedInfo)
			setResetBets(!resetBets())
		})

		api.Subscribe(Event.Invoices, (payload) => {
			if (payload.status !== Status.Success) {
				return
			}

			const inf = info()
			if (inf) {
				infoOptions.mutate({
					prize_pool: inf.prize_pool + payload.amount,
					capacity: inf.capacity - payload.amount
				})
			}
		})

		api.Subscribe(Event.Payments, (payload) => payload.status === Status.Success && infoOptions.refetch())
	})

	onCleanup(() => {
		api.Close()
		clearInterval(timer)
	})

	return (
		<Container>
			<Show when={!info.loading} fallback={<Loading margin="30px 0 100px 0" />}>
				<div class={styles.info} >
					<p class={`${styles.text} ${styles.prize_pool}`}>{t("prize_pool")}</p>
					<Sats num={info()?.prize_pool} fontSize="2.813rem" fontWeight="700" />
				</div>
				<div class={styles.secondary}>
					<div class={styles.info}>
						<p class={`${styles.text} ${styles.capacity}`}>{t("total_capacity")}</p>
						<Sats num={info()?.capacity} fontSize="1.875rem" fontWeight="600" />
					</div>
					<div class={styles.info}>
						<p class={`${styles.text} ${styles.time_left}`}>{t("time_left")}</p>
						<h3 class={styles.time_left_num}>{timeLeft()}</h3>
					</div>
				</div>
			</Show>

			<div class={styles.blocks}>
				<Winners />
				<Bets limit={numPrizes} reset={resetBets()} />
			</div>
		</Container>
	);
};

export default Home;