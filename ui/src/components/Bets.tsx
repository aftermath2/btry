import { Component } from "solid-js";
import { SetStoreFunction } from "solid-js/store";
import { useI18n } from "@solid-primitives/i18n";

import List from "./List";
import { Bet } from "../types/bets";
import { Status } from "../types/events";
import { useAPIContext } from "../context/APIContext";

interface Props {
	limit: number
	hideTitleLink?: boolean
	reset?: boolean
	showPagination?: boolean
}

const Bets: Component<Props> = (props) => {
	const api = useAPIContext()
	const [t] = useI18n()

	const getBets = async (height: number): Promise<Bet[]> => {
		if (height === undefined) {
			return []
		}
		const resp = await api.GetBets(height, 0, props.limit, true)
		return resp.bets
	}

	const subscribeInvoices = (bets: any, setBets: SetStoreFunction<any>): void => {
		api.Subscribe("invoices", (payload) => {
			if (payload.status !== Status.Success) {
				return
			}

			if (bets === undefined) {
				bets = []
			}

			const lastIndex = bets.length !== 0 ? bets[0].index : 0
			const newBet = {
				index: lastIndex + payload.amount,
				public_key: payload.public_key,
				tickets: payload.amount,
			}
			bets.unshift(newBet)
			setBets(bets.slice(0, props.limit))
		})
	}

	return (
		<List
			itemsName="bets"
			tableHeaders={[t("index"), t("nickname"), t("tickets")]}
			getItems={(height) => getBets(height)}
			subscribe={subscribeInvoices}
			showPagination={props.showPagination}
			hideTitleLink={props.hideTitleLink}
			reset={props.reset}
		/>
	)
};

export default Bets;
