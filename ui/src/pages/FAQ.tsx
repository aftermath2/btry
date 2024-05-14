import { Component, For } from 'solid-js';

import styles from './FAQ.module.css';
import Container from "../components/Container";
import Box from "../components/Box";

interface FAQ {
	question: string
	answer: string
}

// TODO:
// - translate questions and answers
const FAQs: FAQ[] = [
	{
		question: "How do lotteries work?",
		answer: `Users participate for the opportunity of winning the funds accumulated in the prize pool by the other participants' bets. Each lottery lasts ${import.meta.env.VITE_LOTTERY_DURATION_BLOCKS} Bitcoin blocks.

Winning tickets are generated using the bytes of the Bitcoin block hash that was mined at the lottery height target.

As soon as the block is mined, winners are announced and any user can generate the winning tickets themselves and verify that the prizes were correctly assigned.

BTRY decodes the block hash and iterates the bytes in reverse, it uses two numbers to calculate each winning ticket. The formula used is (a ^ b) % prizePool.`,
	},
	{
		question: "How are prizes distributed?",
		answer: `Prizes as a percentage of the prize pool:

1st: 50%
2nd: 25%
3rd: 12.5%
4th: 6.25%
5th: 3.125%
6th: 1.5625%
7th: 0.78125%
8th: 0.390625%
		
BTRY fee: 0.390625%`,
	},
	{
		question: "How long can I keep my prizes in the platform?",
		answer: `Prizes expire after ${import.meta.env.VITE_PRIZES_EXPIRATION_BLOCKS} blocks. Consider enabling the notifications to receive a message if you win. 

If BTRY were to hold prizes indefinitely, the node's ability to send and receive payments would eventually be annulled, since all the liquidity would be locked on the same side of the channels. In other words, local balance would be near 100% and remote balance would be near 0%.`,
	},
	{
		question: "Are on-chain payments supported?",
		answer: `Not for now. Although we are aware that the Lightning Network is not the most efficient way of sending large payments, we expect the pools to be small and therefore we are initially using this layer only.

However, we do plan to accept on-chain transactions and even provide winners the alternative to withdraw funds from either layer, but that would require us to have more liquidity or features such as splices (not available in LND yet).`,
	},
	{
		question: "What can I do to avoid withdrawal failures?",
		answer: `Sometimes you may run into issues while trying to withdraw large amounts or using a low fee. This is due to BTRY's node having a relatively small capacity and not being connected to a lot of peers, however, there are a few steps that could be decisive to make the payment work.

1. Increase the withdrawal fee to a number near ${import.meta.env.VITE_WITHDRAWAL_FEE_PPM}ppm (amount * ${import.meta.env.VITE_WITHDRAWAL_FEE_PPM} / 1000000).

2. Try with a smaller amount. There might not be a route with enough liquidity for a large payment. BTRY is using multi-path-payments to avoid this kind of issues but they generally require higher fees.

3. If you are using a custodial wallet, consider trying with a different one. Some nodes are better connected than others.

4. Open a channel to BTRY's node and swap out the funds to have enough inbound liquidity.`
	},
	{
		question: "How do I connect to BTRY's Lightning node?",
		answer: `Clearnet address: ${import.meta.env.VITE_NODE_CLEARNET_ADDRESS}

Tor address: ${import.meta.env.VITE_NODE_TOR_ADDRESS}`,
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
