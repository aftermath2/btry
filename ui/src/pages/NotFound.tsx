import { Component } from "solid-js";

import styles from "./NotFound.module.css";

const NotFound: Component = () => {
	return (
		<div class={styles.container}>
			<h1>404 - Not found</h1>
		</div>
	)
};

export default NotFound;
