FROM node:20-bullseye-slim as development

WORKDIR /usr/src/ui

COPY package.json package.json
COPY package-lock.json package-lock.json

RUN npm install

CMD ["npm", "start"]

FROM development as production

RUN npm run build