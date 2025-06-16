FROM alpine:latest AS with-secret

RUN --mount=type=secret,id=TEST_SECRET1 cat /run/secrets/TEST_SECRET1


FROM alpine:latest AS with-label

RUN echo "With Label Test"

FROM alpine:latest AS with-build-arg

ARG TEST_ARG1
RUN echo "$TEST_ARG1" > testarg.txt

FROM quay.io/skopeo/stable AS with-registry-auth

RUN echo "With Label Test"