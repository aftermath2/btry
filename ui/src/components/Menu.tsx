import { Component, Show, createSignal } from "solid-js";
import { useI18n } from "@solid-primitives/i18n";

import styles from "./Menu.module.css";
import copyIcon from "../assets/icons/copy.svg"
import { useAuthContext } from "../context/AuthContext";
import { Hash, HexEncode, HexRegex } from "../utils/utils";
import { HandleError, WriteClipboard } from "../utils/actions";
import Button from "./Button";
import { GenerateKeyPair, GetPublicKey } from "../utils/crypto";
import { NicknameFromKey } from "../utils/nickname/nickname";
import Modal from "./Modal";
import Input from "./Input";
import QRCode from "../components/QRCode";

const errInvalidLength = new Error("The private key must be longer than 12 characters")

interface Props {
	show: boolean
}

const Menu: Component<Props> = (props) => {
	const [auth, setAuth] = useAuthContext()
	const [t] = useI18n()
	const telegramResolve = `tg://resolve?domain=${import.meta.env.VITE_TELEGRAM_BOT_USERNAME}&start=${auth().publicKey}`
	const telegramURL = `https://t.me/${import.meta.env.VITE_TELEGRAM_BOT_USERNAME}?start=${auth().publicKey}`

	const [showRestoreModal, setShowRestoreModal] = createSignal(false)
	const [showTelegramModal, setShowTelegramModal] = createSignal(false)
	const [privateKey, setPrivateKey] = createSignal("")

	const createKeyPair = async () => {
		const [privateKey, publicKey] = await GenerateKeyPair()
		const nickname = await NicknameFromKey(publicKey)
		setAuth({
			privateKey: privateKey,
			publicKey: publicKey,
			nickname: nickname,
		})
	}

	const onSubmit = async () => {
		let privKey = privateKey()
		if (privKey.length < 12) {
			throw errInvalidLength
		}

		const isHex = HexRegex.test(privKey)
		if (isHex) {
			if (privKey.length !== 64) {
				privKey = await Hash(privKey)
			}
		} else {
			if (privKey.length === 32) {
				privKey = HexEncode(privKey)
			} else {
				privKey = await Hash(privKey)
			}
		}

		const publicKey = await GetPublicKey(privKey)
		const nickname = await NicknameFromKey(publicKey)
		setAuth({
			privateKey: privKey,
			publicKey: publicKey,
			nickname: nickname,
		})

		setShowRestoreModal(false)
		setPrivateKey("")
	}

	return (
		<Show when={props.show}>
			<div class={styles.container}>
				<div class={styles.menu}>
					<p class={styles.title}>{t("private_key")}</p>

					<div class={styles.key}>
						<p class={styles.text}>{auth().privateKey}</p>
						<button class={styles.copy} onClick={() => WriteClipboard(auth().privateKey, t("clipboard_copy"))}>
							<img class={styles.copy_icon} src={copyIcon} />
						</button>
					</div>

					<hr class={styles.line} />
					<Button
						text={t("restore")}
						textFontSize="14px"
						onClick={() => setShowRestoreModal(true)}
					/>
					<Button
						text={t("generate_random_keys")}
						textFontSize="14px"
						onClick={() => createKeyPair()}
					/>
					<Button
						text={t("enable_notifications", { service: "telegram" })}
						textFontSize="14px"
						onClick={() => setShowTelegramModal(true)}
					/>
				</div>
			</div>

			<Modal show={showRestoreModal()} onClose={() => setShowRestoreModal(false)}>
				<Input
					title={t("private_key")}
					value={privateKey()}
					handleInput={(e) => setPrivateKey(e.currentTarget.value)}
					focus
					onEnter={HandleError(() => onSubmit())}
				/>
				<Button text="Submit" onClick={HandleError(() => onSubmit())} />
			</Modal>

			<Modal
				title={t("enable_notifications", { service: "telegram" })}
				show={showTelegramModal()}
				onClose={() => setShowTelegramModal(false)}
			>
				<QRCode value={telegramResolve} />

				<div class={styles.invoice}>
					<p class={styles.pr}>{telegramURL}</p>
					<button class={styles.copy}>
						<img
							class={styles.copy_icon}
							src={copyIcon}
							onClick={() => WriteClipboard(telegramURL, t("clipboard_copy"))}
						/>
					</button>
				</div>
			</Modal>
		</Show>
	);
};

export default Menu;
