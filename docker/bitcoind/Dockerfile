FROM polarlightning/bitcoind:26.0

ENV LOCAL_USER_ID=1000

ENV LOCAL_GROUP_ID=1000

RUN groupadd -f -g $LOCAL_GROUP_ID bitcoin && \
	useradd -r -u $LOCAL_USER_ID -g $LOCAL_GROUP_ID bitcoin
	
COPY bitcoin.conf /tmp/bitcoin.conf

COPY entrypoint.sh /root/entrypoint.sh

ENTRYPOINT ["/root/entrypoint.sh"]

CMD ["bitcoind"]
