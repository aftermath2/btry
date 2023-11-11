import toast from "solid-toast";

/**
 * HandleError wraps a function inside a try-catch statement and triggers a notification if there
 * is an error.
 * 
 * @param fn function to run that may fail
 */
export const HandleError = (fn: () => Promise<void>) => {
	return async function () {
		try {
			return await fn()
		} catch (error: any) {
			toast.error(error.message, { duration: 3000 })
		}
	}
}

/**
 * WriteClipboard copies text to the clipboard.
 * 
 * @param text
 */
export const WriteClipboard = (text: string, message?: string) => {
	navigator.clipboard.writeText(text)
	toast.success(message)
}
