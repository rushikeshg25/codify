FROM node:20 AS base

FROM base AS final

# The node image already ships an unprivileged uid 1000 ("node"). The pod
# security context pins the container to that uid, so /app has to be owned by
# it and every install has to run as it.
#
# The project is built at /opt/codeground-template rather than in place.
# /app/codeground is only a mount point: an init container seeds the
# codeground's persistent volume from the template the first time it is used,
# and the volume is what the container actually runs from. Building the project
# once, in the template, keeps node_modules out of the image twice over.
RUN mkdir -p /app/codeground /app/wss /opt/codeground-template \
    && chown -R node:node /app /opt/codeground-template

USER node

WORKDIR /opt/codeground-template
COPY --chown=node:node codeground/package*.json ./
RUN npm install
COPY --chown=node:node codeground ./

WORKDIR /app/wss
COPY --chown=node:node wss/package*.json ./
RUN npm install
COPY --chown=node:node wss ./

EXPOSE 5173 9000

WORKDIR /app
COPY --chown=node:node --chmod=755 start.sh /app/start.sh

CMD ["/app/start.sh"]
