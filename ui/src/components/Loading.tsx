import { Component } from "solid-js";

import styles from "./Loading.module.css";

const Loading: Component = () => {
	return (
		<div class={styles.container}>
			<span class={styles.loading}></span>
		</div>
	);
};

export default Loading;
