import { Component, JSX, Show } from "solid-js";
import { Portal } from "solid-js/web";

import styles from "./Modal.module.css";
import Box from "./Box";

interface Props {
	children: JSX.Element
	show: boolean
	title?: string
	onClose?: () => void
}

const Modal: Component<Props> = (props) => {
	return (
		<Show when={props.show}>
			<Portal>
				<div class={styles.overlay} role="presentation" onClick={props.onClose}>
					<div
						class={styles.modal}
						role="dialog"
						tabIndex={0}
					>
						<Box
							align="center"
							title={props.title}
							maxWidth="500px"
							padding="25px"
							onClick={(e) => e.stopImmediatePropagation()}
						>
							{props.children}
						</Box>
					</div>
				</div>
			</Portal>
		</Show>
	);
};

export default Modal;
