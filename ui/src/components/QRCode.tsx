import type { Component } from 'solid-js';
import { QRCodeSVG } from 'solid-qr-code';

import styles from "./QRCode.module.css"

interface Props {
	value: string
	size?: number
}

const QRCode: Component<Props> = (props) => {
	return (
		<QRCodeSVG
			class={styles.image}
			value={props.value}
			bgColor={"#ffffff"}
			fgColor={"#000000"}
			level={"Q"}
			includeMargin
		/>
	);
}

export default QRCode;
