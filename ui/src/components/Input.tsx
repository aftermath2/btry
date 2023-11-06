import { Component, JSX, onMount } from "solid-js";

import styles from "./Input.module.css"

interface Props {
	title: string
	handleInput: JSX.EventHandlerUnion<HTMLInputElement, InputEvent>
	value: any
	onEnter?: (e?: KeyboardEvent) => void
	validate?: (v: string) => boolean
	focus?: boolean
	baseProps?: JSX.InputHTMLAttributes<HTMLInputElement>
	placeholder?: string
}

const Input: Component<Props> = (props: Props) => {
	let ref: HTMLInputElement

	const validateInput: JSX.EventHandlerUnion<HTMLInputElement, KeyboardEvent> = (event) => {
		if (props.validate && !props.validate(event.key)) {
			event.preventDefault()
		}
	}

	const validatePaste: JSX.EventHandlerUnion<HTMLInputElement, ClipboardEvent> = (event) => {
		const pastedData = event.clipboardData?.getData("text").trim()
		if (pastedData && props.validate && !props.validate(pastedData)) {
			event.stopPropagation()
			event.preventDefault()
		}
	}

	onMount(() => {
		if (props.focus) {
			ref?.focus()
		}
	})

	return (
		<div class={styles.container}>
			<p class={styles.title}>{props.title}</p>
			<input
				class={styles.value}
				ref={ref!}
				placeholder={props.placeholder || props.title}
				{...props.baseProps}
				type="text"
				onKeyPress={(e) => validateInput(e)}
				onKeyDown={(e) => {
					if (e.key == "Enter" && props.onEnter) {
						props.onEnter()
					}
				}}
				onPaste={(e) => validatePaste(e)}
				onInput={props.handleInput}
				value={props.value}
			></input>
		</div>
	);
};

export default Input;
