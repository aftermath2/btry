import { Component, Show, createResource, createSignal, on, onCleanup, onMount } from "solid-js";
import { createStore } from "solid-js/store";
import { useI18n } from "@solid-primitives/i18n";
import toast from 'solid-toast';

import styles from "./Winners.module.css";
import { Winner } from "../types/winners"
import Loading from "./Loading";
import Box from "./Box";
import NoItems from "./NoItems";
import Table from "./Table";
import Pagination from "./Pagination";
import { BeautifyNumber } from "../utils/utils";
import { HandleError } from "../utils/actions";
import { useAuthContext } from "../context/AuthContext";
import { useAPIContext } from "../context/APIContext";

interface Props {
	showPagination?: boolean
	hideTitleLink?: boolean
}

const Winners: Component<Props> = (props) => {
	const [auth] = useAuthContext()
	const api = useAPIContext()
	const [t] = useI18n()

	const [index, setIndex] = createSignal(0)

	const getHeights = async (): Promise<number[]> => {
		const resp = await api.GetHeights()
		setIndex(resp.heights.length - 2)
		return resp.heights
	}
	const [heights, heightsOptions] = createResource<number[]>(getHeights)

	// TODO: fix asynchronous issues
	const getWinners = async (): Promise<Winner[]> => {
		const height = heights()?.at(index())
		if (height === undefined) {
			return []
		}
		const resp = await api.GetWinners(height)
		return resp.winners
	}
	const [winners, winnersOptions] = createResource<Winner[]>(getWinners)

	const onPaginationClick = async (next: boolean) => {
		let i = next ? index() + 1 : index() - 1
		setIndex(i)
		const winners = await getWinners()
		winnersOptions.mutate(winners)
	}

	onMount(() => {
		// Do not subscribe to events if we are in the winners page
		if (props.showPagination) {
			return
		}

		api.Subscribe("info", (payload) => {
			if (payload.winners !== undefined) {
				if (payload.next_height !== undefined) {
					const next_height = payload.next_height
					heightsOptions.mutate((heights) => heights && [next_height, ...heights])
				}

				winnersOptions.mutate(payload.winners)
				payload.winners.forEach((winner) => {
					if (winner.public_key === auth().publicKey) {
						toast.success(
							t("congratulations", { prizes: BeautifyNumber(winner.prizes) }),
							{ "duration": 3000 }
						)
					}
				})
			}
		})
	})

	onCleanup(() => {
		api.Close()
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
						<p class={styles.height}>
							{heights()?.at(index())}
						</p>
					</Show>
					<Show when={winners()?.length} fallback={<NoItems text={t("no_winners")} />}>
						<Table
							headers={[t("ticket"), t("nickname"), t("prize")]}
							rows={winners()}
						/>
					</Show>
				</Show>
			</Box>

			<Show when={props.showPagination}>
				<Pagination
					onClickPrev={HandleError(() => onPaginationClick(false))}
					showPrev={index() > 0}
					showNext={index() < (heights()?.length || 0) - 1}
					onClickNext={HandleError(() => onPaginationClick(true))}
				/>
			</Show>
		</div>
	);
};

export default Winners;
