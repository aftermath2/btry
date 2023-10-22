import { Component, JSX } from "solid-js";

import styles from "./Container.module.css";

interface Props {
	children: JSX.Element,
}

const Container: Component<Props> = (props) => {
	return (
		<div class={styles.container}>
			{props.children}
		</div>
	);
};

export default Container;
