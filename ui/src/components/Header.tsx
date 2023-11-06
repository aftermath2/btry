import { Component, Show, createSignal } from 'solid-js';
import { A } from "@solidjs/router";
import { useI18n } from "@solid-primitives/i18n";
import Dismiss from "solid-dismiss";

import styles from "./Header.module.css";
import logo from "../assets/icons/logo.svg"
import angleDownIcon from "../assets/icons/angle_down.svg"
import { useAuthContext } from "../context/AuthContext";
import Menu from "./Menu";

const Header: Component = () => {
	let boxRef: HTMLButtonElement

	const [auth] = useAuthContext()
	const [t] = useI18n()

	const [showMenu, setShowMenu] = createSignal(false)

	return (
		<header class={styles.header}>
			<div class={styles.container}>
				<A href="/" class={styles.brand}>
					<img class={styles.logo} src={logo} />
					<p>BTRY</p>
				</A>

				<div class={styles.pages}>
					<A class={styles.link} href="/bet">{t("bet")}</A>
					<A class={styles.link} href="/withdraw">{t("withdraw")}</A>
					<A class={styles.link} href="/faq">FAQ</A>
				</div>

				<div class={styles.dropdown}>
					<button class={styles.box} ref={boxRef!} >
						<p class={styles.nickname}>{auth().nickname}</p>
						<Show when={showMenu()} fallback={<img class={styles.icon} src={angleDownIcon} />}>
							<img class={styles.icon_up} src={angleDownIcon} />
						</Show>
					</button>
				</div>
				<Dismiss menuButton={boxRef!} open={showMenu} setOpen={setShowMenu}>
					<Menu show={showMenu()} />
				</Dismiss>
			</div>
		</header>
	);
};

export default Header;