FROM alpine:3.9

COPY server/ace_away /
COPY dist /dist

CMD ["/ace_away", "-path", "/dist"]
