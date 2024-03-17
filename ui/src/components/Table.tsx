import { Component, For, createEffect, createSignal } from "solid-js";

import styles from "./Table.module.css";
import { Winner } from "../types/winners";
import { Bet } from "../types/bets";
import { NicknameFromKey } from "../utils/nickname/nickname";
import { BeautifyNumber } from "../utils/utils";
import Sats from "./Sats";

interface Row {
	num: number
	nickname: string
	sats: number
}

interface Props {
	headers: string[]
	rows?: Bet[] | Winner[]
}

const Table: Component<Props> = (props) => {

	const [rows, setRows] = createSignal<Row[]>([])

	createEffect(async () => {
		if (!props.rows) {
			return
		}

		const array: Row[] = []

		for (const row of props.rows) {
			const nickname = await NicknameFromKey(row.public_key)

			if ("index" in row) {
				array.push({
					num: row.index,
					nickname: nickname,
					sats: row.tickets
				})
				continue
			}

			array.push({
				num: row.ticket,
				nickname: nickname,
				sats: row.prize
			})
		}

		setRows(array)
	})

	return (
		<table class={styles.container}>
			<thead class={styles.header}>
				<For each={props.headers}>
					{(header, i) => (
						<th class={styles.title} style={{ "text-align": i() === props.headers.length - 1 ? "right" : "left" }}>{header}</th>
					)}
				</For>
			</thead>
			<tbody class={styles.rows}>
				<For each={rows()}>
					{(row, _) => (
						<tr class={styles.row}>
							<td class={styles.text}>{BeautifyNumber(row.num)}</td>
							<td class={styles.text}>{row.nickname}</td>
							<td class={styles.text}><Sats num={row.sats} align="right" /></td>
						</tr>
					)}
				</For>
			</tbody>
		</table>
	);
};

export default Table;
