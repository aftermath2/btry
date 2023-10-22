import { Component, Show } from "solid-js";
import { useI18n } from "@solid-primitives/i18n";

import styles from "./Pagination.module.css";
import Button from "./Button";

interface Props {
	onClickPrev: () => void
	onClickNext: () => void
	showNext?: boolean
}

const Pagination: Component<Props> = (props) => {
	const [t] = useI18n()

	return (
		<div class={styles.container}>
			<div>
				<Button text={t("previous")} onClick={props.onClickPrev} />
			</div>

			<Show when={props.showNext}>
				<div>
					<Button text={t("next")} onClick={props.onClickNext} />
				</div>
			</Show>
		</div>
	);
};

export default Pagination;
