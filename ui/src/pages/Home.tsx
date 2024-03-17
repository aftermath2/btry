import { Component, createSignal, createEffect, onCleanup, createResource, Show } from 'solid-js';
import { useI18n } from "@solid-primitives/i18n";

import styles from './Home.module.css';

import Winners from "../components/Winners";
import Bets from "../components/Bets";
import { LotteryInfo } from "../types/lottery";
import Loading from "../components/Loading";
import Container from "../components/Container";
import { Status } from "../types/events";
import Sats from "../components/Sats";
import { useAPIContext } from "../context/APIContext";
import { BeautifyNumber } from "../utils/utils";

const numPrizes = 8

const Home: Component = () => {
	const api = useAPIContext()
	const [t] = useI18n()

	const [nextHeight, setNextHeight] = createSignal(0)
	const [resetBets, setResetBets] = createSignal(false)

	const getLotteryInfo = async (): Promise<LotteryInfo> => {
		return await api.GetLottery()
	}
	const [info, infoOptions] = createResource<LotteryInfo>(getLotteryInfo)

	createEffect(() => {
		api.Subscribe("info", (payload) => {
			if (payload.next_height !== undefined) {
				setNextHeight(payload.next_height)
			}

			if (payload.capacity !== undefined && payload.prize_pool !== undefined) {
				const updatedInfo: LotteryInfo = {
					prize_pool: payload.prize_pool,
					capacity: payload.capacity,
					next_height: nextHeight()
				}
				infoOptions.mutate(updatedInfo)
				setResetBets(!resetBets())
			}
		})

		api.Subscribe("invoices", (payload) => {
			if (payload.status !== Status.Success) {
				return
			}

			const inf = info()
			if (inf) {
				infoOptions.mutate({
					prize_pool: inf.prize_pool + payload.amount,
					capacity: inf.capacity - payload.amount,
					next_height: inf.next_height
				})
			}
		})

		api.Subscribe("payments", (payload) => payload.status === Status.Success && infoOptions.refetch())
	})

	onCleanup(() => {
		api.Close()
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
						<p class={`${styles.text} ${styles.capacity}`}>{t("maximum_capacity")}</p>
						<Sats num={info()?.capacity} fontSize="1.875rem" fontWeight="600" />
					</div>
					<div class={styles.info}>
						<p class={`${styles.text} ${styles.block_height}`}>
							{t("target_block_height")}
						</p>
						<h3 class={styles.block_height_num}>
							{BeautifyNumber(info()?.next_height)}
						</h3>
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
