import { Component, Show, createResource, createSignal, onCleanup, onMount } from "solid-js";
import { useI18n } from "@solid-primitives/i18n";
import toast from 'solid-toast';

import styles from "./Winners.module.css";
import { API } from "../api/api";
import { Winner } from "../types/winners"
import Loading from "./Loading";
import Box from "./Box";
import NoItems from "./NoItems";
import Table from "./Table";
import Pagination from "./Pagination";
import { BeautifyNumber } from "../utils/utils";
import { HandleError } from "../utils/actions";
import { useAuthContext } from "../context/AuthContext";

interface Props {
	showPagination?: boolean
	hideTitleLink?: boolean
}

const Winners: Component<Props> = (props) => {
	const [auth] = useAuthContext()
	const [t] = useI18n()
	const api = new API()

	const oneDaySecs = 86400
	const initialFrom = new Date().setUTCHours(0, 0, 0, 0) / 1000

	const [from, setFrom] = createSignal(initialFrom)

	const getWinners = async (): Promise<Winner[]> => {
		const resp = await api.GetWinnersHistory(from(), from() + oneDaySecs - 1)
		return resp.winners
	}
	const [winners, winnersOptions] = createResource<Winner[]>(getWinners)

	const onPaginationClick = async (next: boolean) => {
		const n = next ? from() + oneDaySecs : from() - oneDaySecs
		setFrom(n)
		const winners = await getWinners()
		winnersOptions.mutate(winners)
	}

	onMount(() => {
		// Do not subscribe to events if we are in the winners page
		if (props.showPagination) {
			return
		}

		api.ListenLotteryInfoEvents((payload) => {
			if (payload.winners !== undefined) {
				winnersOptions.mutate(payload.winners)
				for (let winner of payload.winners) {
					if (winner.public_key === auth().publicKey) {
						toast.success(t("congratulations", { prizes: BeautifyNumber(winner.prizes) }))
					}
				}
			}
		})
	})

	onCleanup(() => {
		api.Abort()
	})

	return (
		<div>
			<Box
				title={t("winners")}
				align="justify-between"
				minWidth="18vw"
				padding="25px"
				margin="0"
				titleFontSize="1.3rem"
				titleHref={props.hideTitleLink ? undefined : "/winners"}
			>
				<Show when={!winners.loading} fallback={<Loading />}>
					<Show when={props.showPagination}>
						<p class={styles.date}>
							{(new Date(from() * 1000)).toLocaleDateString()}
						</p>
					</Show>
					<Show when={winners()} fallback={<NoItems text={t("no_winners")} />}>
						<Table headers={[t("ticket"), t("nickname"), t("prize")]} rows={winners()} />
					</Show>
				</Show>
			</Box>

			<Show when={props.showPagination}>
				<Pagination
					onClickPrev={HandleError(() => onPaginationClick(false))}
					showNext={from() < initialFrom}
					onClickNext={HandleError(() => onPaginationClick(true))}
				/>
			</Show>
		</div>
	);
};

export default Winners;
