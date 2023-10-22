import { Component } from "solid-js";

import styles from "./NoItems.module.css";

interface Props {
	text: string
}

const NoItems: Component<Props> = (props) => {
	return (
		<div class={styles.container}>
			<p class={styles.text}>{props.text}</p>
		</div>
	);
};

export default NoItems;
