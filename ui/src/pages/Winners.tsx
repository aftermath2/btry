import { Component } from 'solid-js';

import Container from "../components/Container";
import WinnersComponent from "../components/Winners";

const Winners: Component = () => {
	return (
		<Container>
			<WinnersComponent showPagination hideTitleLink />
		</Container>
	);
};

export default Winners;