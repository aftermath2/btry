/* @refresh reload */
import { render } from 'solid-js/web';
import { I18nContext } from "@solid-primitives/i18n";

import './index.css';
import App from './App';
import { AuthProvider } from "./context/AuthContext";
import { APIProvider } from "./context/APIContext";
import { i18nContext } from "./i18n/i18n";
import { Router } from "@solidjs/router";

const root = document.getElementById('root');

if (import.meta.env.DEV && !(root instanceof HTMLElement)) {
	throw new Error(
		'Root element not found. Did you forget to add it to your index.html? Or maybe the id attribute got misspelled?',
	);
}

render(() => (
	<AuthProvider>
		<I18nContext.Provider value={i18nContext}>
			<APIProvider>
				<App />
			</APIProvider>
		</I18nContext.Provider>
	</AuthProvider>
), root!);
