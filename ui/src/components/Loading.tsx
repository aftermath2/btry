import { Component } from "solid-js";

import styles from "./Loading.module.css";

interface Props {
	margin?: string
	width?: string
}

const Loading: Component<Props> = (props) => {
	return (
		<div class={styles.container} style={{ margin: props.margin }}>
			<span class={styles.loading} style={{ width: props.width, height: props.width }}></span>
		</div>
	);
};

export default Loading;
