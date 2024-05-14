import { type Component, ErrorBoundary, onMount, lazy } from 'solid-js';
import { Router, Route } from "@solidjs/router";
import { Toaster } from 'solid-toast';

import styles from "./App.module.css";
import Header from "./components/Header";
import Footer from "./components/Footer";
import { useAuthContext } from "./context/AuthContext";
import { GenerateKeyPair } from "./utils/crypto";
import { NicknameFromKey } from "./utils/nickname/nickname";

// Lazy-load route components
const Bet = lazy(() => import("./pages/Bet"));
const Bets = lazy(() => import("./pages/Bets"));
const FAQ = lazy(() => import("./pages/FAQ"));
const NotFound = lazy(() => import("./pages/NotFound"));
const Winners = lazy(() => import("./pages/Winners"));
const Withdraw = lazy(() => import("./pages/Withdraw"));
const Home = lazy(() => import("./pages/Home"));

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

			<ErrorBoundary fallback={err => <Error error={err.message} />}>
				<Router root={Header}>
					<Route path="*404" component={NotFound} />
					<Route path="/" component={Home} />
					<Route path="/bet" component={Bet} />
					<Route path="/withdraw" component={Withdraw} />
					<Route path="/faq" component={FAQ} />
					<Route path="/bets" component={Bets} />
					<Route path="/winners" component={Winners} />
				</Router>
			</ErrorBoundary>

			<Footer />
		</div>
	);
};

export default App;
