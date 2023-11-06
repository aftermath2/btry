import { Show, type Component } from 'solid-js';

import githubIcon from "../assets/icons/github.svg"
import torIcon from "../assets/icons/tor.svg"
import styles from "./Footer.module.css"

// Declare variable that is being replaced at build time
declare const __COMMIT_HASH__: string

const Footer: Component = () => {
	return (
		<footer>
			<div class={styles.container}>
				<div class={styles.logos}>
					<a
						class={styles.logo}
						href="https://github.com/aftermath2/btry"
						type="button"
						rel="noreferrer"
					>
						<img class={styles.logo_github} src={githubIcon} />
					</a>
					<a
						class={styles.logo}
						// href="http://<url>.onion"
						type="button"
						rel="noreferrer"
					>
						<img class={styles.logo_tor} src={torIcon} />
					</a>
				</div>
				<p class={styles.text}>Made with ⚡️ by aftermath</p>
				<Show when={__COMMIT_HASH__ !== ""}>
					<div class={styles.version}>
						<a
							class={styles.version_text}
							href={`https://github.com/aftermath2/btry/commit/${__COMMIT_HASH__}`}
							rel="noreferrer"
						>
							{__COMMIT_HASH__}
						</a>
					</div>
				</Show>
			</div>
		</footer>
	);
};

export default Footer;