import { Component, Show } from "solid-js";
import { useI18n } from "@solid-primitives/i18n";

import styles from "./Pagination.module.css";
import Button from "./Button";

interface Props {
	onClickPrev: () => void
	onClickNext: () => void
	showNext?: boolean
	showPrev?: boolean
}

const Pagination: Component<Props> = (props) => {
	const [t] = useI18n()

	return (
		<div class={styles.container}>
			{/* Add empty fallback so the next button stays in place */}
			<Show when={props.showPrev} fallback={<div></div>}>
				<div>
					<Button text={t("previous")} onClick={props.onClickPrev} />
				</div>
			</Show>

			<Show when={props.showNext}>
				<div>
					<Button text={t("next")} onClick={props.onClickNext} />
				</div>
			</Show>
		</div>
	);
};

export default Pagination;
