import { Component, JSX, Show } from "solid-js";
import { A } from "@solidjs/router";

import styles from "./Box.module.css";

interface Props {
	title?: string
	titleFontSize?: string
	titleMargin?: string
	titleHref?: string
	children: JSX.Element
	align: "flex-start" | "center" | "flex-end" | "justify-between" | "justify-around"
	padding?: string
	margin?: string
	maxWidth?: string
	minWidth?: string
	flex?: number
	onClick?: JSX.EventHandlerUnion<HTMLDivElement, MouseEvent>
}

const Box: Component<Props> = (props) => {
	return (
		<div
			class={styles.box}
			onClick={props.onClick}
			style={{
				padding: props.padding,
				"max-width": props.maxWidth,
				"min-width": props.minWidth,
				margin: props.margin,
				"justify-content": props.align,
				"align-items": props.align,
				flex: props.flex
			}}
		>
			<Show when={props.title}>
				<Show when={props.titleHref} fallback={
					<p class={styles.title}
						style={{
							"font-size": props.titleFontSize,
							margin: props.titleMargin
						}}>
						{props.title}
					</p>}
				>
					<A href={props.titleHref || ""} class={styles.title_link}
						style={{
							"font-size": props.titleFontSize,
							margin: props.titleMargin
						}}>
						{props.title}
					</A>
				</Show>
			</Show>
			{props.children}
		</div>
	);
};

export default Box;
