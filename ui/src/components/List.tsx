import { Component, Show, createEffect, createSignal, on, onCleanup, onMount } from "solid-js";
import { SetStoreFunction, Store, createStore } from "solid-js/store";
import { useI18n } from "@solid-primitives/i18n";

import styles from "./List.module.css";
import Loading from "./Loading";
import Box from "./Box";
import NoItems from "./NoItems";
import Table from "./Table";
import Pagination from "./Pagination";
import { HandleError } from "../utils/actions";
import { useAPIContext } from "../context/APIContext";

interface Props {
	itemsName: "bets" | "winners"
	tableHeaders: string[]
	getItems: (height: number) => Promise<any[]>
	subscribe: (items: Store<any>, setItems: SetStoreFunction<any>) => void
	showPagination?: boolean
	hideTitleLink?: boolean
	reset?: boolean
}

const List: Component<Props> = (props) => {
	const api = useAPIContext()
	const [t] = useI18n()
	const heightOffset = props.itemsName === "bets" ? 1 : 2

	const [index, setIndex] = createSignal(0)
	// Not using resources because the values depend on each other
	const [heights, setHeigths] = createStore<number[]>([])
	const [items, setItems] = createStore<any[]>([])

	const getHeights = async (): Promise<number[]> => {
		const resp = await api.GetHeights()
		setIndex(resp.heights.length - heightOffset)
		return resp.heights
	}

	const onPaginationClick = async (next: boolean) => {
		let i = next ? index() + 1 : index() - 1
		setIndex(i)
		const items = await props.getItems(heights[index()])
		if (items === undefined) {
			setItems([])
			return
		}
		setItems(items)
	}

	// Reset the items list when the property changes
	createEffect(
		on(
			() => props.reset,
			() => setItems([])
		)
	)

	onMount(async () => {
		const heights = await getHeights()
		setHeigths(heights)
		const items = await props.getItems(heights[index()])
		if (items !== undefined) {
			setItems(items)
		}

		// Do not subscribe to events if we are in the items page
		if (props.showPagination) {
			return
		}

		props.subscribe(items, setItems)
	})

	onCleanup(() => {
		api.Close()
	})

	return (
		<div
			class={styles.container}
			style={{ "flex-direction": props.showPagination ? "column" : "row" }}
		>
			<Box
				title={t(props.itemsName)}
				align="justify-between"
				minWidth="18vw"
				padding="25px"
				margin="0"
				titleFontSize="1.3rem"
				titleHref={props.hideTitleLink ? undefined : `/${props.itemsName}`}
			>
				<Show when={items} fallback={<Loading />}>
					<Show when={props.showPagination}>
						<p class={styles.height}>
							{`${t("block_height")}: ${heights[index()] || 0}`}
						</p>
					</Show>
					<Show
						when={items.length > 0}
						fallback={<NoItems text={t(`no_${props.itemsName}`)} />}
					>
						<Table
							headers={props.tableHeaders}
							rows={items}
						/>
					</Show>
				</Show>
			</Box>

			<Show when={props.showPagination}>
				<Pagination
					onClickPrev={HandleError(() => onPaginationClick(false))}
					showPrev={index() > 0}
					showNext={index() < heights.length - heightOffset}
					onClickNext={HandleError(() => onPaginationClick(true))}
				/>
			</Show>
		</div>
	);
};

export default List;
