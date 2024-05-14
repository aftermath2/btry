import { Component, Show, createSignal } from 'solid-js';
import { useI18n } from "@solid-primitives/i18n";
import Dismiss from "solid-dismiss";

import styles from "./Header.module.css";
import logo from "../assets/icons/logo.svg"
import angleDownIcon from "../assets/icons/angle_down.svg"
import { useAuthContext } from "../context/AuthContext";
import Menu from "./Menu";
import { A, RouteSectionProps } from "@solidjs/router";

const highligthTabStyle = { "border-bottom": "3px solid var(--red)" }

const Header: Component<RouteSectionProps<unknown>> = (props) => {
	let boxRef: HTMLButtonElement

	const [auth] = useAuthContext()
	const [t] = useI18n()

	const [showMenu, setShowMenu] = createSignal(false)

	return (
		<>
			<nav class={styles.header}>
				<div class={styles.container}>
					<div class={styles.brand}>
						<A class={styles.home} href="/">
							<img class={styles.logo} src={logo} />
							<p class={styles.name}>BTRY</p>
						</A>
					</div>

					<div class={styles.pages}>
						<A
							class={styles.link}
							href="/bet"
							style={props.location.pathname === '/bet' ? highligthTabStyle : {}}
						>
							{t("bet")}
						</A>
						<A
							class={styles.link}
							href="/withdraw"
							style={props.location.pathname === '/withdraw' ? highligthTabStyle : {}}
						>
							{t("withdraw")}
						</A>
						<A
							class={styles.link}
							href="/faq"
							style={props.location.pathname === '/faq' ? highligthTabStyle : {}}
						>
							FAQ
						</A>
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
			</nav>
			{props.children}
		</>
	);
};

export default Header;
