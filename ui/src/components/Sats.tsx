import { Component } from "solid-js";

import styles from "./Sats.module.css";
import satoshiIcon from "../assets/icons/satoshi.svg"
import { BeautifyNumber } from "../utils/utils";

interface Props {
	num?: number
	align?: string
	fontSize?: string
	fontWeight?: string
}

const Sats: Component<Props> = (props) => {
	return (
		<div class={styles.container} style={{ "justify-content": props.align }}>
			<p
				class={styles.number}
				style={{
					"font-size": props.fontSize,
					"font-weight": props.fontWeight
				}}
			>
				{BeautifyNumber(props.num)}
			</p>

			<img
				class={styles.icon}
				src={satoshiIcon}
				style={{
					width: props.fontSize,
					height: props.fontSize
				}}
			/>
		</div>
	);
};

export default Sats;
