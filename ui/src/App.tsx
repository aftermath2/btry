import { type Component, ErrorBoundary, onMount } from 'solid-js';
import { Routes, Route } from "@solidjs/router";
import { Toaster } from 'solid-toast';

import styles from "./App.module.css";
import Home from "./pages/Home";
import Header from "./components/Header";
import Footer from "./components/Footer";
import Bet from "./pages/Bet";
import Withdraw from "./pages/Withdraw";
import FAQ from "./pages/FAQ";
import Bets from "./pages/Bets";
import Winners from "./pages/Winners";
import NotFound from "./pages/NotFound";
import { useAuthContext } from "./context/AuthContext";
import { GenerateKeyPair } from "./utils/crypto";
import { NicknameFromKey } from "./utils/nickname/nickname";

const App: Component = () => {
	const [auth, setAuth] = useAuthContext()

	const createKeyPair = async () => {
		const [privateKey, publicKey] = await GenerateKeyPair()
		const nickname = await NicknameFromKey(publicKey)
		setAuth({
			privateKey: privateKey,
			publicKey: publicKey,
			nickname: nickname,
		})
	}

	onMount(() => {
		if (auth().privateKey === "") {
			createKeyPair()
		}
	})

	const Error: Component<{ error: string }> = (props) => {
		return (
			<div class={styles.error}>
				<h2>Something went wrong</h2>
				<p>{props.error}</p>
			</div>
		)
	}

	return (
		<div class={styles.app}>
			<Toaster />
			<Header />

			<ErrorBoundary fallback={err => <Error error={err.message} />}>
				<Routes>
					<Route path="*" component={NotFound} />
					<Route path="/" component={Home} />
					<Route path="/bet" component={Bet} />
					<Route path="/withdraw" component={Withdraw} />
					<Route path="/faq" component={FAQ} />
					<Route path="/bets" component={Bets} />
					<Route path="/winners" component={Winners} />
				</Routes>
			</ErrorBoundary>

			<Footer />
		</div>
	);
};

export default App;
