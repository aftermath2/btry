import { Component, For } from 'solid-js';

import styles from './FAQ.module.css';
import Container from "../components/Container";
import Box from "../components/Box";

interface FAQ {
	question: string
	answer: string
}

// TODO: 
// - get variable information from the server
// - translate questions and answers
const FAQs: FAQ[] = [
	{
		question: "How do lotteries work?",
		answer: `Users participate for the opportunity of winning the funds that were bet in the same UTC day. 

Winners are announced every day at 00:00:00 UTC using a cryptographically secure random number generator.

Prizes distribution as a percentage of the total prize pool:

1st place: 50% 
2nd place: 25%
3rd place: 12.5%
4th place: 6.25%
5th place: 3.125%
6th place: 1.5625%
7th place: 0.78125%
8th place: 0.390625%

BTRY fee: 0.390625%`,
	},
	{
		question: "How long can I keep my prizes in the account?",
		answer: `Prizes expire after 120 hours (5 days). Consider enabling the notifications to receive a message if you win. 

If BTRY were to hold prizes indefinitely, the node's ability to send and receive payments would eventually be annulled, since all the liquidity would be locked on the same side of the channels. In other words, local balance would be near 100% and remote balance would be near 0%.`,
	},
	{
		question: "Are on-chain payments supported?",
		answer: `Not for now. Although we are aware that the Lightning Network is not the best way to send large payments, we expect the pools to be small and therefore we are initially using this layer only.

However, we do plan to accept on-chain transactions and even provide winners the alternative to withdraw funds from either layer, but that would require us to have more liquidity or features such as splices (not available in LND yet).`,
	},
	{
		question: "What can I do to avoid withdrawal failures?",
		answer: `Sometimes you may run into issues while trying to withdraw large amounts or using a low fee. This is due to BTRY's node having a relatively small capacity and not being connected to a lot of peers, however, there are a few steps that could be decisive to make the payment work.

1. Increase the withdrawal fee to a number near 1500ppm (amount * 1500 / 1000000).

2. Try with a smaller amount. There might not be a route with enough liquidity for a large payment. BTRY is using multi-path-payments to avoid this kind of issues but they generally require higher fees.

3. If you are using a custodial wallet, consider trying with a different one. Some nodes are better connected than others.

4. Open a channel to BTRY's node and swap out the funds to have enough inbound liquidity.`
	},
	{
		question: "How do I connect to BTRY's Lightning node?",
		answer: `Clearnet address: 
032c52ba7d06c6361bc137bf0b78e47ad09b46014ca8d0c1b5d005180207373a37@5.75.184.195:40705

Tor address: 
032c52ba7d06c6361bc137bf0b78e47ad09b46014ca8d0c1b5d005180207373a37@zx4ek52khoefpt4kuztqqd7cybnx66ls3whdwlt4ktjpplgckb3bc6qd.onion:9735`,
	},
]

const FAQ: Component = () => {
	return (
		<Container>
			<Box
				title="FAQ"
				align="flex-start"
				maxWidth="45%"
				padding="30px 50px"
				titleFontSize="30px"
			>
				<For each={FAQs}>
					{(faq, _) =>
						<div>
							<p class={styles.question}>{faq.question}</p>
							<p class={styles.answer}>{faq.answer}</p>
							<br class={styles.space} />
						</div>
					}
				</For>
			</Box>
		</Container>
	);
};

export default FAQ;