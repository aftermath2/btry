import { Component, Show, createSignal, onMount } from "solid-js";
import { useI18n } from "@solid-primitives/i18n";

import styles from "./Menu.module.css";
import copyIcon from "../assets/icons/copy.svg"
import { useAPIContext } from "../context/APIContext";
import { useAuthContext } from "../context/AuthContext";
import { Hash, HexEncode, HexRegex } from "../utils/utils";
import { HandleError, WriteClipboard } from "../utils/actions";
import Button from "./Button";
import { GenerateKeyPair, GetPublicKey } from "../utils/crypto";
import { NicknameFromKey } from "../utils/nickname/nickname";
import Modal from "./Modal";
import Input from "./Input";
import QRCode from "../components/QRCode";
import { ValidateLightningAddress } from "../utils/lightning";

const errInvalidLength = new Error("The private key must be longer than 12 characters")

interface Props {
	show: boolean
}

const Menu: Component<Props> = (props) => {
	const api = useAPIContext()
	const [auth, setAuth] = useAuthContext()
	const [t] = useI18n()

	const telegramResolve = `tg://resolve?domain=${import.meta.env.VITE_TELEGRAM_BOT_USERNAME}&start=${auth().publicKey}`
	const telegramURL = `https://t.me/${import.meta.env.VITE_TELEGRAM_BOT_USERNAME}?start=${auth().publicKey}`

	const [privateKey, setPrivateKey] = createSignal("")
	const [lightningAddress, setLightningAddress] = createSignal("")
	const [showRestoreModal, setShowRestoreModal] = createSignal(false)
	const [showTelegramModal, setShowTelegramModal] = createSignal(false)
	const [showLightningAddressModal, setShowLightningAddressModal] = createSignal(false)

	const createKeyPair = async () => {
		setLightningAddress("")
		const [privateKey, publicKey] = await GenerateKeyPair()
		const nickname = await NicknameFromKey(publicKey)
		setAuth({
			privateKey: privateKey,
			publicKey: publicKey,
			nickname: nickname,
		})
	}

	const onPrivateKeySubmit = async () => {
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
		setLightningAddress("")
	}

	const onLightningAddressClick = async () => {
		const resp = await api.GetLightningAddress()
		if (resp.has_address) {
			setLightningAddress(resp.address)
		}

		setShowLightningAddressModal(true)
	}

	const onLightningAddressSubmit = async () => {
		ValidateLightningAddress(lightningAddress())
		await api.SetLightningAddress(lightningAddress())
		setShowLightningAddressModal(false)
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
					<Button
						text={`${t("set")} lightning address`}
						textFontSize="14px"
						onClick={onLightningAddressClick}
					/>
				</div>
			</div>

			<Modal show={showRestoreModal()} onClose={() => setShowRestoreModal(false)}>
				<Input
					title={t("private_key")}
					value={privateKey()}
					handleInput={(e) => setPrivateKey(e.currentTarget.value)}
					focus
					onEnter={HandleError(() => onPrivateKeySubmit())}
				/>
				<Button text="Submit" onClick={HandleError(() => onPrivateKeySubmit())} />
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

			<Modal show={showLightningAddressModal()} onClose={() => setShowLightningAddressModal(false)}>
				<Input
					title="Lightning address"
					value={lightningAddress()}
					handleInput={(e) => setLightningAddress(e.currentTarget.value)}
					focus
					onEnter={HandleError(() => onLightningAddressSubmit())}
				/>
				<Button text="Submit" onClick={HandleError(() => onLightningAddressSubmit())} />
			</Modal>
		</Show>
	);
};

export default Menu;
