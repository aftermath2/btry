import { Component } from 'solid-js';

import Container from "../components/Container";
import BetsComponent from "../components/Bets";

const Bets: Component = () => {
	return (
		<Container>
			<BetsComponent limit={50} hideTitleLink />
		</Container>
	);
};

export default Bets;