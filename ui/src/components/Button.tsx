import { Component } from "solid-js";

import styles from "./Button.module.css"

interface Props {
	text: string
	textFontSize?: string
	onClick?: (e: MouseEvent) => void
	label?: string
	width?: string
}

const Button: Component<Props> = (props: Props) => {
	return (
		<div class={styles.container}>
			<button
				class={styles.button}
				aria-label={props.label}
				onClick={props.onClick}
				style={{
					width: props.width,
					"font-size": props.textFontSize
				}}
			>
				{props.text}
			</button>
		</div>
	);
};

export default Button;
