import { Component } from "solid-js";
import { SetStoreFunction } from "solid-js/store";
import { useI18n } from "@solid-primitives/i18n";
import toast from 'solid-toast';

import List from "./List";
import { Winner } from "../types/winners"
import { BeautifyNumber } from "../utils/utils";
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

	const getWinners = async (height: number): Promise<Winner[]> => {
		if (height === undefined) {
			return []
		}
		const resp = await api.GetWinners(height)
		return resp.winners
	}

	const subscribeInfo = (
		setWinners: SetStoreFunction<any>,
		heights: any,
		setHeights: SetStoreFunction<number[]>
	): void => {
		api.Subscribe("info", (payload) => {
			if (payload.winners !== undefined) {
				if (payload.next_height !== undefined) {
					setHeights(heights.length, payload.next_height)
				}

				setWinners(payload.winners)

				let prizes = 0
				payload.winners.forEach((winner) => {
					if (winner.public_key === auth().publicKey) {
						prizes += winner.prize
					}
				})

				toast.success(
					t("congratulations", { prizes: BeautifyNumber(prizes) }),
					{ "id": "0", "duration": 3000 } // Use any id to avoid duplicated pop ups
				)
			}
		})
	}

	return (
		<List
			itemsName="winners"
			tableHeaders={[t("ticket"), t("nickname"), t("prize")]}
			getItems={(height) => getWinners(height)}
			subscribe={subscribeInfo}
			showPagination={props.showPagination}
			hideTitleLink={props.hideTitleLink}
		/>
	)
};

export default Winners;
