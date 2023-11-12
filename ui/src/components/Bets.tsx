import { Component, Show, createEffect, createResource, on, onCleanup, onMount } from "solid-js";
import { useI18n } from "@solid-primitives/i18n";

import { Bet } from "../types/bets";
import Loading from "./Loading";
import Box from "./Box";
import NoItems from "./NoItems";
import Table from "./Table";
import { Status } from "../types/events";
import { useAPIContext } from "../context/APIContext";

interface Props {
	limit: number
	hideTitleLink?: boolean
	reset?: boolean
}

const Bets: Component<Props> = (props) => {
	const api = useAPIContext()
	const [t] = useI18n()

	const getBets = async (): Promise<Bet[]> => {
		const resp = await api.GetBets(0, props.limit, true)
		return resp.bets
	}
	const [bets, betsOptions] = createResource<Bet[]>(getBets)

	// Reset the bets list when the property changes (received when the lottery has finished)
	createEffect(
		on(
			() => props.reset,
			() => betsOptions.mutate(undefined)
		)
	)

	onMount(() => {
		api.Subscribe("invoices", (payload) => {
			if (payload.status !== Status.Success) {
				return
			}

			const b = bets()
			const lastIndex = b ? b[0].index : 0
			const newBet = {
				index: lastIndex + payload.amount,
				public_key: payload.public_key,
				tickets: payload.amount,
			}
			betsOptions.mutate((bets) => bets && [newBet, ...bets.slice(0, props.limit)])
		})
	})

	onCleanup(() => {
		api.Close()
	})

	return (
		<Box
			title={t("bets")}
			align="justify-between"
			minWidth="18vw"
			padding="25px"
			margin="0"
			titleFontSize="1.3rem"
			titleHref={props.hideTitleLink ? undefined : "/bets"}
		>
			<Show when={!bets.loading} fallback={<Loading />}>
				<Show when={bets()} fallback={<NoItems text={t("no_bets")} />}>
					<Table headers={[t("index"), t("nickname"), t("tickets")]} rows={bets()} />
				</Show>
			</Show>
		</Box>
	);
};

export default Bets;
